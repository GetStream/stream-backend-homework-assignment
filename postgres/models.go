package postgres

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
)

// A message represents a message in the database.
type message struct {
	ID              string                   `bun:",pk,type:uuid,default:uuid_generate_v4()"`
	MessageText     string                   `bun:"message_text,notnull"`
	UserID          string                   `bun:",notnull"`
	ReactionScore   int                      `bun:",notnull"`
	ListOfReactions []constants.ReactionType `bun:",notnull"`
	CreatedAt       time.Time                `bun:",nullzero,default:now()"`
}

type reaction struct {
	ID            string    `bun:",pk,type:uuid,default:uuid_generate_v4()"`
	UserID        string    `bun:",notnull"`
	MessageID     string    `bun:",type:uuid,notnull"`
	ReactionType  string    `bun:",notnull"`
	ReactionScore int       `bun:",notnull"`
	CreatedAt     time.Time `bun:",nullzero,default:now()"`

	//foreign keys
	Message *message `bun:"rel:belongs-to,join:message_id=id"`
}

func (m message) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.MessageText,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}

func (m message) APIMessageV2() api.Message {
	return api.Message{
		ID:              m.ID,
		Text:            m.MessageText,
		UserID:          m.UserID,
		ReactionScore:   m.ReactionScore,
		ListOfReactions: m.ListOfReactions,
		CreatedAt:       m.CreatedAt,
	}
}

func (r reaction) MessageReaction() api.ReactionV2 {
	return api.ReactionV2{
		ID:            r.ID,
		UserID:        r.UserID,
		MessageID:     r.MessageID,
		ReactionType:  r.ReactionType,
		ReactionScore: r.ReactionScore,
		CreatedAt:     r.CreatedAt,
	}
}
