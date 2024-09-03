package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
	"github.com/GetStream/stream-backend-homework-assignment/api/schemas"
	"github.com/GetStream/stream-backend-homework-assignment/api/types"
	"github.com/GetStream/stream-backend-homework-assignment/api/utils"
	"github.com/GetStream/stream-backend-homework-assignment/api/validators"
)

func (a *API) FindMessages(ctx context.Context, pageNumber, offset, pageSize int) ([]Message, error) {
	if pageNumber == 1 {
		msgsFromCache, err := a.Cache.ListMessages(ctx, pageSize)
		if err != nil {
			return nil, fmt.Errorf("error fetching messages from cache: %w", err)
		}
		if len(msgsFromCache) == pageSize {
			return msgsFromCache, nil
		}

		remain := pageSize - len(msgsFromCache)
		a.Logger.Info("Fetched partial records from cache, fetching remaining from DB", "remaining", remain)

		dbMsgs, err := a.DB.ListMessages(ctx, len(msgsFromCache), remain)
		if err != nil {
			return nil, fmt.Errorf("error fetching remaining messages from DB: %w", err)
		}
		return append(msgsFromCache, dbMsgs...), nil
	}

	a.Logger.Info("Fetching messages from DB...", "offset", offset, "limit", pageSize)
	dbMsgs, err := a.DB.ListMessages(ctx, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("error fetching messages from DB: %w", err)
	}
	a.Logger.Info("Got messages from DB", "count", len(dbMsgs))
	return dbMsgs, nil
}

func TransformMessagesToResponse(msgs []Message) []types.Message {
	out := make([]types.Message, len(msgs))
	for i, msg := range msgs {
		out[i] = types.Message{
			ID:              msg.ID,
			Text:            msg.Text,
			UserID:          msg.UserID,
			ListOfReactions: msg.ListOfReactions,
			ReactionScore:   msg.ReactionScore,
			CreatedAt:       msg.CreatedAt.Format(time.RFC1123),
		}
	}
	return out
}

func (a *API) listMessagesV2(w http.ResponseWriter, r *http.Request) {
	pageNumber, err := utils.GetPageNumber(r.URL.Query().Get("page"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Invalid Page")
		return
	}

	// input validation
	if err := validators.ValidateListQueryRequest(schemas.ListRequestQuery{
		PageNumber: pageNumber,
	}); err != nil {
		a.respondError(w, http.StatusBadRequest, err, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	// TODO: can be moved to env or dedicated config manager
	const pageSize = 10
	offset := (pageNumber - 1) * pageSize

	msgs, err := a.FindMessages(r.Context(), pageNumber, offset, pageSize)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
		return
	}

	out := TransformMessagesToResponse(msgs)

	a.respond(w, http.StatusOK, schemas.ListResponse{Messages: out})
}

func (a *API) createMessageV2(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("[v2] create message api called")
	var body schemas.CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Could not decode request body")
		return
	}
	r.Body.Close()

	// input validation
	if err := validators.ValidateCreateMessageRequest(body); err != nil {
		a.respondError(w, http.StatusBadRequest, err, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	msg, err := a.DB.InsertMessage(r.Context(), Message{
		Text:            body.Text,
		UserID:          body.UserID,
		ReactionScore:   0,
		ListOfReactions: []constants.ReactionType{},
		CreatedAt:       time.Now(),
	})
	if err != nil {
		a.Logger.Debug("[v2] Could not insert message", "err", err)
		a.respondError(w, http.StatusInternalServerError, err, "Could not insert v2 message")
		return
	}

	if err := a.Cache.InsertMessage(r.Context(), msg); err != nil {
		a.Logger.Error("Could not cache message", "error", err.Error())
	}

	res := schemas.CreateMessageResponseV2{
		ID:              msg.ID,
		Text:            msg.Text,
		UserID:          msg.UserID,
		ListOfReactions: msg.ListOfReactions,
		ReactionScore:   msg.ReactionScore,
		CreatedAt:       msg.CreatedAt.Format(time.RFC1123),
	}
	a.respond(w, http.StatusCreated, res)
}
