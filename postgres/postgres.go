package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/GetStream/stream-backend-homework-assignment/api/constants"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// Postgres provides storage in PostgreSQL.
type Postgres struct {
	bun *bun.DB
}

// Connect connects to the database and ping the DB to ensure the connection is
// working.
func Connect(ctx context.Context, connStr string) (*Postgres, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connStr)))
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	db := bun.NewDB(sqlDB, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook())

	fmt.Println(db)
	return &Postgres{
		bun: db,
	}, nil
}

// ListMessages returns all messages in the database.
func (pg *Postgres) ListMessages(ctx context.Context, offset, limit int) ([]api.Message, error) {
	var msgs []message
	q := pg.bun.NewSelect().
		Model(&msgs).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)
	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	out := make([]api.Message, len(msgs))
	for i, m := range msgs {
		out[i] = m.APIMessageV2()
	}
	return out, nil
}

// InsertMessage inserts a message into the database. The returned message
// holds auto generated fields, such as the message id.
func (pg *Postgres) InsertMessage(ctx context.Context, msg api.Message) (api.Message, error) {
	m := &message{
		MessageText:     msg.Text,
		UserID:          msg.UserID,
		ListOfReactions: msg.ListOfReactions,
		ReactionScore:   msg.ReactionScore,
	}
	if _, err := pg.bun.NewInsert().Model(m).Exec(ctx); err != nil {
		return api.Message{}, fmt.Errorf("insert: %w", err)
	}
	return m.APIMessageV2(), nil
}

func (pg *Postgres) FindMessage(ctx context.Context, msgId string) error {
	var msg message

	q := pg.bun.NewSelect().
		Model(&msg).
		Where("id=?", msgId)

	if err := q.Scan(ctx); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	return nil
}

func (pg *Postgres) InsertReactionAndUpdateMessage(ctx context.Context, react api.ReactionV2) (api.ReactionV2, error) {
	tx, err := pg.bun.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return api.ReactionV2{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	r := &reaction{
		ReactionType:  react.ReactionType,
		ReactionScore: react.ReactionScore,
		UserID:        react.UserID,
		MessageID:     react.MessageID,
	}

	if _, err := tx.NewInsert().Model(r).Exec(ctx); err != nil {
		return api.ReactionV2{}, fmt.Errorf("insert: %w", err)
	}

	var msg message
	q := pg.bun.NewSelect().
		Model(&msg).
		Where("id = ?", react.MessageID)

	if err := q.Scan(ctx); err != nil {
		return api.ReactionV2{}, fmt.Errorf("scan: %w", err)
	}

	listOfReactions := append(msg.ListOfReactions, constants.ReactionType(react.ReactionType))
	reactionScore := msg.ReactionScore + 1

	_, err = tx.NewUpdate().Model(&msg).
		Set("list_of_reactions = ?", listOfReactions).
		Set("reaction_score = ?", reactionScore).
		Where("id = ?", react.MessageID).
		Exec(ctx)

	if err != nil {
		fmt.Println("Error updating message:", err)
		return api.ReactionV2{}, err
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		return api.ReactionV2{}, err
	}

	return r.MessageReaction(), nil
}

func (pg *Postgres) InsertReaction(ctx context.Context, reaction api.Reaction) (api.Reaction, error) {
	panic("not implemented")
}
