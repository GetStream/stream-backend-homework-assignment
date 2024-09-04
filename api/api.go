package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api/schemas"
	"github.com/GetStream/stream-backend-homework-assignment/api/validators"
)

// A DB provides a storage layer that persists messages.
type DB interface {
	ListMessages(ctx context.Context, offset, limit int) ([]Message, error)
	InsertMessage(ctx context.Context, msg Message) (Message, error)
	InsertReactionAndUpdateMessage(ctx context.Context, react ReactionV2) (ReactionV2, error)
	FindMessage(ctx context.Context, msgId string) error
}

// A Cache provides a storage layer that caches messages.
type Cache interface {
	ListMessages(ctx context.Context, limit int) ([]Message, error)
	InsertMessage(ctx context.Context, msg Message) error
}

// API provides the REST endpoints for the application.
type API struct {
	Logger *slog.Logger
	DB     DB
	Cache  Cache
	once   sync.Once
	mux    *http.ServeMux
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /messages", a.listMessages)
	mux.HandleFunc("POST /messages", a.createMessage)
	mux.HandleFunc("POST /messages/{messageID}/reactions", a.createReaction)

	// v2 apis with requested changes. Taking this approach to avoid any issues for existing clients
	mux.HandleFunc("POST /v2/messages", a.createMessageV2)
	mux.HandleFunc("GET /v2/messages", a.listMessagesV2)

	a.mux = mux
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.Logger.Info("Request received", "method", r.Method, "path", r.URL.Path)
	a.mux.ServeHTTP(w, r)
}

func (a *API) respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		a.Logger.Error("Could not encode JSON body", "error", err.Error())
	}
}

func (a *API) respondError(w http.ResponseWriter, status int, err error, msg string) {
	type response struct {
		Error string `json:"error"`
	}
	a.Logger.Error("Error", "error", err.Error())
	a.respond(w, status, response{Error: msg})
}

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	type message struct {
		ID        string `json:"id"`
		Text      string `json:"text"`
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
	}
	type response struct {
		Messages []message `json:"messages"`
	}

	pageSize := 10

	msgs, err := a.Cache.ListMessages(r.Context(), pageSize)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
		return
	}

	remain := pageSize - len(msgs)
	a.Logger.Info("Got messages from cache", "count", len(msgs), "remain", remain)

	if remain > 0 {
		dbMsgs, err := a.DB.ListMessages(r.Context(), len(msgs), remain)
		if err != nil {
			a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
			return
		}
		a.Logger.Info("Got remaining messages from DB", "count", len(dbMsgs))
		msgs = append(msgs, dbMsgs...)
	}

	out := make([]message, len(msgs))
	for i, msg := range msgs {
		out[i] = message{
			ID:        msg.ID,
			Text:      msg.Text,
			UserID:    msg.UserID,
			CreatedAt: msg.CreatedAt.Format(time.RFC1123),
		}
	}
	res := response{
		Messages: out,
	}
	a.respond(w, http.StatusOK, res)
}

func (a *API) createMessage(w http.ResponseWriter, r *http.Request) {
	var body schemas.CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Could not decode request body")
		return
	}
	r.Body.Close()

	msg, err := a.DB.InsertMessage(r.Context(), Message{
		Text:      body.Text,
		UserID:    body.UserID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not insert message")
		return
	}

	if err := a.Cache.InsertMessage(r.Context(), msg); err != nil {
		a.Logger.Error("Could not cache message", "error", err.Error())
	}

	res := schemas.CreateMessageResponse{
		ID:        msg.ID,
		Text:      msg.Text,
		UserID:    msg.UserID,
		CreatedAt: msg.CreatedAt.Format(time.RFC1123),
	}
	a.respond(w, http.StatusCreated, res)
}

func (a *API) createReaction(w http.ResponseWriter, r *http.Request) {
	messageID := r.PathValue("messageID")
	var body schemas.MessageReactionRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Could not decode request body")
		return
	}
	r.Body.Close()
	// input validation
	if err := validators.ValidateReactionRequest(body); err != nil {
		a.respondError(w, http.StatusBadRequest, err, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	if body.ReactionScore == nil {
		a.Logger.Debug("setting the default value for reaction score")
		defaultScore := 1
		body.ReactionScore = &defaultScore
	}

	// update message with reaction and reactionType. In an ideal scenario we should update the message reaction count and type via async communication (msg queue, eg message queue/fan out method)
	react, err := a.DB.InsertReactionAndUpdateMessage(r.Context(), ReactionV2{
		ReactionType:  string(body.ReactionType),
		ReactionScore: *body.ReactionScore,
		UserID:        body.UserID,
		MessageID:     messageID,
		CreatedAt:     time.Now(),
	})

	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not insert reaction")
		return
	}

	res := schemas.CreateReactionResponse{
		ID:            react.ID,
		MessageID:     react.MessageID,
		UserID:        react.UserID,
		ReactionScore: react.ReactionScore,
		ReactionType:  react.ReactionType,
		CreatedAt:     react.CreatedAt.Format(time.RFC1123),
	}
	// logs?
	a.respond(w, http.StatusCreated, res)
}
