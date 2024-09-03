package validators

import (
	"github.com/GetStream/stream-backend-homework-assignment/api/schemas"
	"github.com/go-playground/validator"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateReactionRequest(req schemas.MessageReactionRequest) error {
	return validate.Struct(req)
}

func ValidateCreateMessageRequest(req schemas.CreateMessageRequest) error {
	return validate.Struct(req)
}

func ValidateListQueryRequest(req schemas.ListRequestQuery) error {
	return validate.Struct(req)
}
