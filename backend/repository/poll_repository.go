package repository

import (
	"context"
	"errors"
	"time"

	"pulsevote/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PollRepository handles poll and vote persistence operations in MongoDB
type PollRepository struct {
	pollsCol *mongo.Collection
	votesCol *mongo.Collection
}

// NewPollRepository initializes a new PollRepository instance
func NewPollRepository(pollsCol *mongo.Collection, votesCol *mongo.Collection) *PollRepository {
	return &PollRepository{
		pollsCol: pollsCol,
		votesCol: votesCol,
	}
}

// Create inserts a new poll into MongoDB
func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.ID = primitive.NewObjectID()
	poll.IsActive = true
	poll.CreatedAt = time.Now()
	poll.UpdatedAt = time.Now()

	_, err := r.pollsCol.InsertOne(ctx, poll)
	return err
}

// FindByUserID retrieves all polls created by a specific user, sorted newest first
func (r *PollRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Poll, error) {
	filter := bson.M{"createdBy": userID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.pollsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

// FindByShareCode searches for a poll by its unique public shareCode
func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	var poll models.Poll
	filter := bson.M{"shareCode": shareCode}

	err := r.pollsCol.FindOne(ctx, filter).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

// FindByID searches for a poll by its MongoDB ObjectID
func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	filter := bson.M{"_id": id}

	err := r.pollsCol.FindOne(ctx, filter).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &poll, nil
}

// UpdateQuestion updates a poll's question if owned by the user
func (r *PollRepository) UpdateQuestion(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, question string) (*models.Poll, error) {
	filter := bson.M{"_id": id, "createdBy": userID}
	update := bson.M{
		"$set": bson.M{
			"question":  question,
			"updatedAt": time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedPoll models.Poll
	err := r.pollsCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedPoll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("poll not found or unauthorized")
		}
		return nil, err
	}

	return &updatedPoll, nil
}

// ClosePoll marks a poll as inactive (closed)
func (r *PollRepository) ClosePoll(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*models.Poll, error) {
	filter := bson.M{"_id": id, "createdBy": userID}
	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var closedPoll models.Poll
	err := r.pollsCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&closedPoll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("poll not found or unauthorized")
		}
		return nil, err
	}

	return &closedPoll, nil
}

// DeletePoll removes a poll and associated votes permanently
func (r *PollRepository) DeletePoll(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": id, "createdBy": userID}
	res, err := r.pollsCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("poll not found or unauthorized")
	}

	// Also clean up votes recorded for this poll
	_, _ = r.votesCol.DeleteMany(ctx, bson.M{"pollId": id})
	return nil
}

// SaveVote records a single vote permanently in MongoDB
func (r *PollRepository) SaveVote(ctx context.Context, vote *models.Vote) error {
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()

	_, err := r.votesCol.InsertOne(ctx, vote)
	return err
}

// AggregateOptionCounts calculates counts directly from permanent MongoDB votes collection
func (r *PollRepository) AggregateOptionCounts(ctx context.Context, pollID primitive.ObjectID) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "pollId", Value: pollID}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$optionId"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.votesCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			OptionID string `bson:"_id"`
			Count    int64  `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			counts[doc.OptionID] = doc.Count
		}
	}

	return counts, nil
}
