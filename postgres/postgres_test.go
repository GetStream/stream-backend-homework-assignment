//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/google/go-cmp/cmp"
)

func TestPostgres_ListMessages(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(pg *Postgres) error
		offset int
		limit  int
		want   []api.Message
	}{
		{
			name:   "Empty",
			offset: 0,
			limit:  10,
			want:   []api.Message{},
		},
		{
			name: "One",
			setup: func(pg *Postgres) error {
				msgs := []message{
					{
						ID:          "388d74ea-cc39-4566-860f-0df6068f3330",
						MessageText: "hello",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				_, err := pg.bun.NewInsert().Model(&msgs).Exec(context.Background())
				return err
			},
			offset: 0,
			limit:  10,
			want: []api.Message{
				{
					ID:        "388d74ea-cc39-4566-860f-0df6068f3330",
					Text:      "hello",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Two",
			setup: func(pg *Postgres) error {
				msgs := []message{
					{
						ID:          "4562fe69-42b3-46e5-b990-11581182f57c",
						MessageText: "hello",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "7c6d956b-58d6-4ac3-9984-f341346edc37",
						MessageText: "world",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}
				_, err := pg.bun.NewInsert().Model(&msgs).Exec(context.Background())
				return err
			},
			offset: 0,
			limit:  10,
			want: []api.Message{
				{ // First because of DESC sorting on the created_at column.
					ID:        "7c6d956b-58d6-4ac3-9984-f341346edc37",
					Text:      "world",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "4562fe69-42b3-46e5-b990-11581182f57c",
					Text:      "hello",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Page",
			setup: func(pg *Postgres) error {
				msgs := []message{
					{
						ID:          "a691e2d8-316b-42ae-85c9-5691d0b7dcd2",
						MessageText: "foo",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "02a7b937-3450-40f4-a0c6-83da4985081f",
						MessageText: "bar",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "1923601a-53f2-48ba-a61b-bd7bb9ece431",
						MessageText: "baz",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "609153c8-867b-4d40-859b-d2acf3a9d232",
						MessageText: "qux",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
					},
				}
				_, err := pg.bun.NewInsert().Model(&msgs).Exec(context.Background())
				return err
			},
			offset: 1,
			limit:  2,
			want: []api.Message{
				{
					ID:        "1923601a-53f2-48ba-a61b-bd7bb9ece431",
					Text:      "baz",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "02a7b937-3450-40f4-a0c6-83da4985081f",
					Text:      "bar",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			pg := connect(t)
			if tt.setup != nil {
				if err := tt.setup(pg); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			got, err := pg.ListMessages(ctx, tt.offset, tt.limit)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Diff (-got +want)\n%s", diff)
			}
		})
	}
}

func TestPostgres_InsertMessage(t *testing.T) {
	tests := []struct {
		name  string
		msg   api.Message
		check func(t *testing.T, pg *Postgres)
	}{
		{
			name: "OK",
			msg: api.Message{
				Text:   "Hello",
				UserID: "testuser",
			},
			check: func(t *testing.T, pg *Postgres) {
				var got message
				if err := pg.bun.NewSelect().Model(&got).Scan(context.Background()); err != nil {
					t.Fatal(err)
				}

				if got.MessageText != "Hello" {
					t.Errorf("Stored message text does not match; got %q, want %q", got.MessageText, "Hello")
				}
				if got.UserID != "testuser" {
					t.Errorf("Stored message user id does not match; got %q, want %q", got.UserID, "testuser")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			pg := connect(t)
			got, err := pg.InsertMessage(ctx, tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, pg)

			if got.ID == "" {
				t.Error("Returned message has empty ID")
			}
			if got.CreatedAt.IsZero() {
				t.Error("Returned message does not have a CreatedAt field")
			}
		})
	}
}

func connect(t *testing.T) *Postgres {
	t.Helper()
	connStr := "postgres://message-api:message-api@localhost:5432/message-api?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pg, err := Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Could not connect to PostgreSQL: %v", err)
	}

	// Truncate the table before each test.
	if _, err := pg.bun.NewTruncateTable().Model((*message)(nil)).Exec(ctx); err != nil {
		t.Fatalf("Could not truncate table: %v", err)
	}

	return pg
}
