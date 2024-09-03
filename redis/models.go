package redis

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
)

// A message represents a message in the database.
type message struct {
	ID              string                   `json:"id"`
	Text            string                   `json:"text"`
	UserID          string                   `json:"user_id"`
	ReactionScore   int                      `json:"reaction_score"`
	ListOfReactions []constants.ReactionType `json:"list_of_reactions"`
	CreatedAt       time.Time                `json:"created_at"`
}

func (m message) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.Text,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}
