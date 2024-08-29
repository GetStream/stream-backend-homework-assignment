package redis

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
)

// A message represents a message in the database.
type message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"create_at"`
}

func (m message) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.Text,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}
