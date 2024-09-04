package api

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
)

// A Message represents a persisted message.
type Message struct {
	ID              string
	Text            string
	UserID          string
	ReactionScore   int
	ListOfReactions []constants.ReactionType
	CreatedAt       time.Time
}

// A Reaction represents a reaction to a message such as a like.
type Reaction struct {
	ID        string
	MessageID string
	Type      string
	Score     int
	UserID    string
	CreatedAt time.Time
}

type ReactionV2 struct {
	ID            string
	MessageID     string
	ReactionType  string
	ReactionScore int
	UserID        string
	CreatedAt     time.Time
}
