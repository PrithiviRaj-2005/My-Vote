package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vote represents a recorded vote in MongoDB permanent storage
type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID  string             `bson:"optionId" json:"optionId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// VoteRequest represents the payload submitted by audience to cast a vote
type VoteRequest struct {
	OptionID string `json:"optionId"`
}

// VotePubSubMessage represents the message published to Redis channel `poll:<shareCode>`
type VotePubSubMessage struct {
	OptionID  string                 `json:"optionId"`
	Increment int64                  `json:"increment"`
	ShareCode string                 `json:"shareCode,omitempty"`
	Results   *PollResultsResponse   `json:"results,omitempty"`
}
