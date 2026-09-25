package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollOption represents a single choice in a poll
type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// Poll represents a poll stored in MongoDB
type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question  string             `bson:"question" json:"question"`
	Options   []PollOption       `bson:"options" json:"options"`
	CreatedBy primitive.ObjectID `bson:"createdBy" json:"createdBy"`
	ShareCode string             `bson:"shareCode" json:"shareCode"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// CreatePollRequest represents payload to create a new poll
type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

// UpdatePollRequest represents payload to edit a poll question
type UpdatePollRequest struct {
	Question string `json:"question"`
}

// OptionResult represents an option along with its live vote counts and percentages
type OptionResult struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Votes      int64   `json:"votes"`
	Percentage float64 `json:"percentage"`
}

// PollResultsResponse represents the full live results of a poll
type PollResultsResponse struct {
	PollID     string         `json:"pollId"`
	Question   string         `json:"question"`
	ShareCode  string         `json:"shareCode"`
	IsActive   bool           `json:"isActive"`
	TotalVotes int64          `json:"totalVotes"`
	Options    []OptionResult `json:"options"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}
