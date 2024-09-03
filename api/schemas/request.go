package schemas

import (
	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
)

type MessageReactionRequest struct {
	ReactionType  constants.ReactionType `json:"reaction_type" validate:"required,oneof=Like Love Wow Angry"`
	ReactionScore *int                   `json:"reaction_score,omitempty"`
	UserID        string                 `json:"user_id" validate:"required"`
}

type CreateMessageRequest struct {
	Text   string `json:"text" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

type ListRequestQuery struct {
	PageNumber int `json:"page" validate:"required,min=1"`
}
