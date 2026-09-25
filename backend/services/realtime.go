package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"pulsevote/models"
	ws "pulsevote/websocket"

	"github.com/redis/go-redis/v9"
)

// RealtimeService manages Redis live vote counts, Pub/Sub channels, and WebSocket broadcasting
// Falls back gracefully to in-memory caching and direct WebSocket broadcast if Redis is unavailable
type RealtimeService struct {
	rdb       *redis.Client
	hub       *ws.Hub
	memMu     sync.RWMutex
	memCounts map[string]map[string]int64
}

// NewRealtimeService initializes a new RealtimeService
func NewRealtimeService(rdb *redis.Client, hub *ws.Hub) *RealtimeService {
	return &RealtimeService{
		rdb:       rdb,
		hub:       hub,
		memCounts: make(map[string]map[string]int64),
	}
}

// resultsKey returns the Redis Hash key for storing live option vote counts
func resultsKey(shareCode string) string {
	return fmt.Sprintf("poll:%s:results", shareCode)
}

// pubsubChannel returns the Redis Pub/Sub channel for a poll
func pubsubChannel(shareCode string) string {
	return fmt.Sprintf("poll:%s", shareCode)
}

// IncrementOptionCount increments the vote count of an option in Redis Hash (or memory)
func (s *RealtimeService) IncrementOptionCount(ctx context.Context, shareCode string, optionID string) (int64, error) {
	if s.rdb == nil {
		s.memMu.Lock()
		defer s.memMu.Unlock()
		if _, ok := s.memCounts[shareCode]; !ok {
			s.memCounts[shareCode] = make(map[string]int64)
		}
		s.memCounts[shareCode][optionID]++
		return s.memCounts[shareCode][optionID], nil
	}

	key := resultsKey(shareCode)
	// HINCRBY poll:<shareCode>:results <optionId> 1
	newCount, err := s.rdb.HIncrBy(ctx, key, optionID, 1).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment vote count in Redis: %w", err)
	}
	return newCount, nil
}

// GetOptionCounts retrieves all option vote counts for a poll from Redis Hash (or memory)
func (s *RealtimeService) GetOptionCounts(ctx context.Context, shareCode string) (map[string]int64, error) {
	if s.rdb == nil {
		s.memMu.RLock()
		defer s.memMu.RUnlock()
		if pollCounts, ok := s.memCounts[shareCode]; ok {
			counts := make(map[string]int64, len(pollCounts))
			for k, v := range pollCounts {
				counts[k] = v
			}
			return counts, nil
		}
		return nil, nil
	}

	key := resultsKey(shareCode)
	// HGETALL poll:<shareCode>:results
	data, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch option counts from Redis: %w", err)
	}

	counts := make(map[string]int64)
	for optID, countStr := range data {
		count, _ := strconv.ParseInt(countStr, 10, 64)
		counts[optID] = count
	}
	return counts, nil
}

// SetOptionCounts populates or syncs the Redis Hash with initial counts
func (s *RealtimeService) SetOptionCounts(ctx context.Context, shareCode string, counts map[string]int64) error {
	if len(counts) == 0 {
		return nil
	}

	if s.rdb == nil {
		s.memMu.Lock()
		defer s.memMu.Unlock()
		if _, ok := s.memCounts[shareCode]; !ok {
			s.memCounts[shareCode] = make(map[string]int64)
		}
		for optID, count := range counts {
			s.memCounts[shareCode][optID] = count
		}
		return nil
	}

	key := resultsKey(shareCode)
	pipe := s.rdb.Pipeline()
	for optID, count := range counts {
		pipe.HSet(ctx, key, optID, count)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// PublishVote broadcasts a vote update via Redis Pub/Sub to poll:<shareCode> (or direct WebSocket)
func (s *RealtimeService) PublishVote(ctx context.Context, shareCode string, optionID string, results *models.PollResultsResponse) error {
	msg := models.VotePubSubMessage{
		OptionID:  optionID,
		Increment: 1,
		ShareCode: shareCode,
		Results:   results,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal pubsub message: %w", err)
	}

	if s.rdb == nil {
		s.hub.Broadcast(shareCode, payload)
		log.Printf("[Realtime (In-Memory)] Broadcasted vote update for poll [%s] via WebSocket hub", shareCode)
		return nil
	}

	channel := pubsubChannel(shareCode)
	// Publish to Redis channel: poll:<shareCode>
	err = s.rdb.Publish(ctx, channel, payload).Err()
	if err != nil {
		// Fallback to local broadcast so connected peers still receive update
		s.hub.Broadcast(shareCode, payload)
		return fmt.Errorf("failed to publish vote event to Redis: %w", err)
	}

	log.Printf("[Redis Pub/Sub] Published vote event for poll [%s] on channel [%s]", shareCode, channel)
	return nil
}

// PublishPollClosed broadcasts a poll closed event via Redis Pub/Sub (or direct WebSocket)
func (s *RealtimeService) PublishPollClosed(ctx context.Context, shareCode string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"event":     "poll_closed",
		"shareCode": shareCode,
		"isActive":  false,
	})
	if err != nil {
		return err
	}

	if s.rdb == nil {
		s.hub.Broadcast(shareCode, payload)
		log.Printf("[Realtime (In-Memory)] Broadcasted poll closed event for poll [%s] via WebSocket hub", shareCode)
		return nil
	}

	channel := pubsubChannel(shareCode)
	err = s.rdb.Publish(ctx, channel, payload).Err()
	if err != nil {
		s.hub.Broadcast(shareCode, payload)
	}
	return err
}

// DeletePollData cleans up Redis keys and in-memory caches when a poll is deleted
func (s *RealtimeService) DeletePollData(ctx context.Context, shareCode string) {
	if s.rdb == nil {
		s.memMu.Lock()
		delete(s.memCounts, shareCode)
		s.memMu.Unlock()
		return
	}
	key := resultsKey(shareCode)
	_ = s.rdb.Del(ctx, key).Err()
}

// StartSubscriber starts listening for Redis Pub/Sub events on "poll:*" and forwards them to WebSocket clients
func (s *RealtimeService) StartSubscriber(ctx context.Context) {
	if s.rdb == nil {
		log.Println("[Realtime] Running in direct WebSocket mode without Redis Pub/Sub subscriber.")
		return
	}

	go func() {
		log.Println("[Redis Pub/Sub] Starting background subscriber pattern 'poll:*'...")
		pubsub := s.rdb.PSubscribe(ctx, "poll:*")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Redis Pub/Sub] Subscriber context canceled. Stopping subscriber.")
				return
			case msg, ok := <-ch:
				if !ok {
					log.Println("[Redis Pub/Sub] Channel closed, attempting reconnect in 2 seconds...")
					time.Sleep(2 * time.Second)
					pubsub = s.rdb.PSubscribe(ctx, "poll:*")
					ch = pubsub.Channel()
					continue
				}

				// Channel name format: poll:<shareCode>
				// Ignore internal hash keys or subkeys
				channelParts := strings.Split(msg.Channel, ":")
				if len(channelParts) != 2 {
					continue
				}
				shareCode := channelParts[1]

				// Forward message to all connected Gorilla WebSocket clients for this poll
				s.hub.Broadcast(shareCode, []byte(msg.Payload))
				log.Printf("[Redis -> WebSocket] Broadcasted update to clients viewing poll [%s]", shareCode)
			}
		}
	}()
}
