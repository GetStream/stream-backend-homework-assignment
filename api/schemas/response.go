package schemas

import (
	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
	"github.com/GetStream/stream-backend-homework-assignment/api/types"
)

type CreateMessageResponse struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type CreateMessageResponseV2 struct {
	ID              string                   `json:"id"`
	Text            string                   `json:"text"`
	UserID          string                   `json:"user_id"`
	ListOfReactions []constants.ReactionType `json:"reaction_list"`
	ReactionScore   int                      `json:"reaction_score"`
	CreatedAt       string                   `json:"created_at"`
}

type CreateReactionResponse struct {
	ID            string `json:"id"`             // reaction ID
	MessageID     string `json:"message_id"`     // message ID
	ReactionType  string `json:"reaction_type"`  // reaction type, for example 'like', 'laugh', 'wow', 'thumbs_up'
	ReactionScore int    `json:"reaction_score"` // reaction score should default to 1 if not specified, but can be any positive integer. Think of claps on Medium.com
	UserID        string `json:"user_id"`        // the user ID submitting the reaction
	CreatedAt     string `json:"created_at"`     // the date/time the reaction was created
}

type ListResponse struct {
	Messages []types.Message `json:"messages"`
}
