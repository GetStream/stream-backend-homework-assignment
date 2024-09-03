package types

import "github.com/GetStream/stream-backend-homework-assignment/api/constants"

type Message struct {
	ID              string                   `json:"id"`
	Text            string                   `json:"text"`
	UserID          string                   `json:"user_id"`
	ListOfReactions []constants.ReactionType `json:"list_of_reactions"`
	ReactionScore   int                      `json:"reaction_score"`
	CreatedAt       string                   `json:"created_at"`
}
