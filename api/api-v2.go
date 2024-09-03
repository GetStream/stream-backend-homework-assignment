package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
	"github.com/GetStream/stream-backend-homework-assignment/api/schemas"
	"github.com/GetStream/stream-backend-homework-assignment/api/types"
	"github.com/GetStream/stream-backend-homework-assignment/api/validators"
)

func (a *API) listMessagesV2(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	pageNumber := 1
	if page != "" {
		pageNumberInt, err := strconv.Atoi(page)
		if err != nil {
			a.respondError(w, http.StatusBadRequest, err, "Invalid Page")

		}
		pageNumber = pageNumberInt
	}

	// input validation
	if err := validators.ValidateListQueryRequest(schemas.ListRequestQuery{
		PageNumber: pageNumber,
	}); err != nil {
		a.respondError(w, http.StatusBadRequest, err, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	pageSize := 10
	var msgs []Message
	offset := (pageNumber - 1) * pageSize

	if pageNumber == 1 {
		msgsFromCache, err := a.Cache.ListMessages(r.Context(), pageSize)
		if err != nil {
			a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
			return
		}
		msgs = append(msgs, msgsFromCache...)
		remain := pageSize - len(msgs)
		if remain > 0 {
			a.Logger.Info("Fetched some records from cache, remaining: ", "remain", remain)
			dbMsgs, err := a.DB.ListMessages(r.Context(), len(msgsFromCache), remain)
			if err != nil {
				a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
				return
			}
			msgs = append(msgs, dbMsgs...)
		}
	} else {
		a.Logger.Info("Fetching messages from db...", fmt.Sprint(offset), pageSize)
		dbMsgs, err := a.DB.ListMessages(r.Context(), offset, pageSize)
		if err != nil {
			a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
			return
		}
		a.Logger.Info("Got messages from DB", "count", len(dbMsgs))
		msgs = append(msgs, dbMsgs...)
	}

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
	res := schemas.ListResponse{
		Messages: out,
	}

	a.respond(w, http.StatusOK, res)
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
