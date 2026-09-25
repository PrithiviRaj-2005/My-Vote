package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"pulsevote/models"
	"pulsevote/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollService handles poll creation, retrieval, updates, and voting workflows
type PollService struct {
	pollRepo *repository.PollRepository
	realtime *RealtimeService
}

// NewPollService creates a new PollService instance
func NewPollService(pollRepo *repository.PollRepository, realtime *RealtimeService) *PollService {
	return &PollService{
		pollRepo: pollRepo,
		realtime: realtime,
	}
}

// generateShareCode creates a random 6-character alphanumeric code
func generateShareCode() (string, error) {
	const charset = "abcdefghjkmnpqrstuvwxyz23456789" // Exclude easily confused chars like 0, O, 1, l
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

// CreatePoll validates input, assigns unique options, and stores the new poll
func (s *PollService) CreatePoll(ctx context.Context, userID primitive.ObjectID, req models.CreatePollRequest) (*models.Poll, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, errors.New("poll question is required")
	}

	// Validate options: min 2, max 10
	if len(req.Options) < 2 {
		return nil, errors.New("a poll must have at least 2 options")
	}
	if len(req.Options) > 10 {
		return nil, errors.New("a poll can have at most 10 options")
	}

	seen := make(map[string]bool)
	pollOptions := make([]models.PollOption, 0, len(req.Options))

	for i, opt := range req.Options {
		trimmed := strings.TrimSpace(opt)
		if trimmed == "" {
			return nil, fmt.Errorf("option #%d cannot be empty", i+1)
		}
		lower := strings.ToLower(trimmed)
		if seen[lower] {
			return nil, fmt.Errorf("duplicate option: '%s'", trimmed)
		}
		seen[lower] = true

		pollOptions = append(pollOptions, models.PollOption{
			ID:   fmt.Sprintf("opt_%d", i+1),
			Text: trimmed,
		})
	}

	// Generate a unique share code
	var shareCode string
	for attempts := 0; attempts < 5; attempts++ {
		code, err := generateShareCode()
		if err != nil {
			return nil, err
		}
		existing, _ := s.pollRepo.FindByShareCode(ctx, code)
		if existing == nil {
			shareCode = code
			break
		}
	}
	if shareCode == "" {
		return nil, errors.New("failed to generate unique share code, please try again")
	}

	poll := &models.Poll{
		Question:  question,
		Options:   pollOptions,
		CreatedBy: userID,
		ShareCode: shareCode,
		IsActive:  true,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	// Initialize empty vote counts in Redis
	initialCounts := make(map[string]int64)
	for _, opt := range pollOptions {
		initialCounts[opt.ID] = 0
	}
	_ = s.realtime.SetOptionCounts(ctx, shareCode, initialCounts)

	return poll, nil
}

// GetUserPolls returns all polls created by a specific user
func (s *PollService) GetUserPolls(ctx context.Context, userID primitive.ObjectID) ([]models.Poll, error) {
	return s.pollRepo.FindByUserID(ctx, userID)
}

// GetPollByShareCode retrieves the public poll structure
func (s *PollService) GetPollByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByShareCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, errors.New("poll not found")
	}
	return poll, nil
}

// GetPollResults retrieves current live results (combining Redis counts with poll metadata)
func (s *PollService) GetPollResults(ctx context.Context, shareCode string) (*models.PollResultsResponse, error) {
	poll, err := s.pollRepo.FindByShareCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, errors.New("poll not found")
	}

	// Fetch live option counts from Redis Hash
	redisCounts, err := s.realtime.GetOptionCounts(ctx, shareCode)
	if err != nil || len(redisCounts) == 0 {
		// Fallback / Warmup: query MongoDB permanent votes
		mongoCounts, mErr := s.pollRepo.AggregateOptionCounts(ctx, poll.ID)
		if mErr == nil {
			redisCounts = mongoCounts
			_ = s.realtime.SetOptionCounts(ctx, shareCode, mongoCounts)
		}
	}

	var totalVotes int64 = 0
	for _, count := range redisCounts {
		totalVotes += count
	}

	optionResults := make([]models.OptionResult, len(poll.Options))
	for i, opt := range poll.Options {
		votes := redisCounts[opt.ID]
		var percentage float64 = 0
		if totalVotes > 0 {
			percentage = (float64(votes) / float64(totalVotes)) * 100.0
		}

		optionResults[i] = models.OptionResult{
			ID:         opt.ID,
			Text:       opt.Text,
			Votes:      votes,
			Percentage: percentage,
		}
	}

	return &models.PollResultsResponse{
		PollID:     poll.ID.Hex(),
		Question:   poll.Question,
		ShareCode:  poll.ShareCode,
		IsActive:   poll.IsActive,
		TotalVotes: totalVotes,
		Options:    optionResults,
		UpdatedAt:  time.Now(),
	}, nil
}

// SubmitVote records a vote permanently to MongoDB, increments Redis, and publishes Pub/Sub
func (s *PollService) SubmitVote(ctx context.Context, shareCode string, optionID string) (*models.PollResultsResponse, error) {
	// 1. Validate poll exists
	poll, err := s.pollRepo.FindByShareCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, errors.New("poll not found")
	}

	// 2. Validate poll is active
	if !poll.IsActive {
		return nil, errors.New("this poll is closed and no longer accepting votes")
	}

	// 3. Validate option ID exists in poll
	var optionValid bool
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			optionValid = true
			break
		}
	}
	if !optionValid {
		return nil, errors.New("invalid option selected")
	}

	// 4. Save vote permanently to MongoDB
	vote := &models.Vote{
		PollID:   poll.ID,
		OptionID: optionID,
	}
	if err := s.pollRepo.SaveVote(ctx, vote); err != nil {
		return nil, fmt.Errorf("failed to save vote: %w", err)
	}

	// 5. Increment Redis live result count (HINCRBY poll:<shareCode>:results <optionId> 1)
	_, err = s.realtime.IncrementOptionCount(ctx, shareCode, optionID)
	if err != nil {
		// Log warning, permanent vote was recorded
		fmt.Printf("[Realtime] Warning: Redis increment failed: %v\n", err)
	}

	// 6. Get updated results to broadcast
	results, err := s.GetPollResults(ctx, shareCode)
	if err != nil {
		return nil, err
	}

	// 7. Publish to Redis Pub/Sub channel poll:<shareCode>
	_ = s.realtime.PublishVote(ctx, shareCode, optionID, results)

	return results, nil
}

// UpdatePoll updates poll question if owned by the user
func (s *PollService) UpdatePoll(ctx context.Context, pollID primitive.ObjectID, userID primitive.ObjectID, req models.UpdatePollRequest) (*models.Poll, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, errors.New("question cannot be empty")
	}
	return s.pollRepo.UpdateQuestion(ctx, pollID, userID, question)
}

// ClosePoll closes poll (isActive = false) if owned by the user and broadcasts status
func (s *PollService) ClosePoll(ctx context.Context, pollID primitive.ObjectID, userID primitive.ObjectID) (*models.Poll, error) {
	closedPoll, err := s.pollRepo.ClosePoll(ctx, pollID, userID)
	if err != nil {
		return nil, err
	}

	// Broadcast poll closed event over Redis Pub/Sub
	_ = s.realtime.PublishPollClosed(ctx, closedPoll.ShareCode)
	return closedPoll, nil
}

// DeletePoll deletes poll and associated data if owned by the user
func (s *PollService) DeletePoll(ctx context.Context, pollID primitive.ObjectID, userID primitive.ObjectID) error {
	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil {
		return err
	}
	if poll == nil {
		return errors.New("poll not found")
	}

	if err := s.pollRepo.DeletePoll(ctx, pollID, userID); err != nil {
		return err
	}

	// Clean up Redis keys
	s.realtime.DeletePollData(ctx, poll.ShareCode)
	return nil
}
