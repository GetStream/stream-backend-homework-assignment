//go:build integration

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
)

func TestRedis_ListMessages(t *testing.T) {
	tests := []struct {
		name  string
		setup func(r *Redis) error
		limit int
		want  []api.Message
	}{
		{
			name:  "Empty",
			limit: 10,
			want:  []api.Message{},
		},
		{
			name: "One",
			setup: func(r *Redis) error {
				members := []redis.Z{
					z(t, message{
						ID:        "9cbf8127-299b-4a84-8920-cd35ea0c084c",
						Text:      "hello",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}),
				}
				return r.cli.ZAdd(context.Background(), messagePrefix, members...).Err()
			},
			limit: 10,
			want: []api.Message{
				{
					ID:        "9cbf8127-299b-4a84-8920-cd35ea0c084c",
					Text:      "hello",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Two",
			setup: func(r *Redis) error {
				members := []redis.Z{
					z(t, message{
						ID:        "1bb3fbd9-01b8-41ed-ac45-3f7c6235e657",
						Text:      "hello",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}),
					z(t, message{
						ID:        "7f1f1803-d3cf-46a9-acd2-6aa9d4b8b4c0",
						Text:      "world",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					}),
				}
				return r.cli.ZAdd(context.Background(), messagePrefix, members...).Err()
			},
			limit: 10,
			want: []api.Message{
				{ // First because of DESC sorting on score (timestamp)
					ID:        "7f1f1803-d3cf-46a9-acd2-6aa9d4b8b4c0",
					Text:      "world",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "1bb3fbd9-01b8-41ed-ac45-3f7c6235e657",
					Text:      "hello",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Limit",
			setup: func(r *Redis) error {
				members := []redis.Z{
					z(t, message{
						ID:        "7676696e-96e2-4115-bbef-bef2a313c7d3",
						Text:      "foo",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}),
					z(t, message{
						ID:        "44df7659-1a04-42a8-8a3c-c607f79c1210",
						Text:      "bar",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					}),
					z(t, message{
						ID:        "6ae0833f-dfe1-40b8-9cd2-c40c636120b5",
						Text:      "baz",
						UserID:    "test",
						CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
					}),
				}
				return r.cli.ZAdd(context.Background(), messagePrefix, members...).Err()
			},
			limit: 2,
			want: []api.Message{
				{ // Latest 2, by score (timestamp)
					ID:        "6ae0833f-dfe1-40b8-9cd2-c40c636120b5",
					Text:      "baz",
					UserID:    "test",
					CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "44df7659-1a04-42a8-8a3c-c607f79c1210",
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

			r := connect(t)
			if tt.setup != nil {
				if err := tt.setup(r); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			got, err := r.ListMessages(ctx, tt.limit)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Diff (-got +want)\n%s", diff)
			}
		})
	}
}

func TestRedis_InsertMessage(t *testing.T) {
	tests := []struct {
		name  string
		msg   api.Message
		check func(t *testing.T, r *Redis)
	}{
		{
			name: "OK",
			msg: api.Message{
				Text:   "Hello",
				UserID: "testuser",
			},
			check: func(t *testing.T, r *Redis) {
				vals, err := r.cli.ZRange(context.Background(), messagePrefix, 0, 10).Result()
				if err != nil {
					t.Fatal(err)
				}
				if len(vals) != 1 {
					t.Fatal("No items in Redis")
				}

				var got message
				if err := json.Unmarshal([]byte(vals[0]), &got); err != nil {
					t.Fatalf("Could not unmarshal JSON: %v", err)
				}

				if got.Text != "Hello" {
					t.Errorf("Stored message text does not match; got %q, want %q", got.Text, "Hello")
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

			r := connect(t)
			err := r.InsertMessage(ctx, tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, r)
		})
	}
}

func TestRedis_InsertMessage_MaxSize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r := connect(t)
	// Insert 11 items.
	for i := 0; i <= maxSize; i++ {
		msg := api.Message{
			Text:      fmt.Sprintf("Message %d", i+1),
			UserID:    "testuser",
			CreatedAt: time.Now().Add(time.Millisecond * time.Duration(i)),
		}
		if err := r.InsertMessage(ctx, msg); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Fetching all 11 items should return 10 items because no more than 10 messages should be stored.
	vals, err := r.cli.ZRevRange(ctx, messagePrefix, 0, maxSize).Result()

	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != maxSize {
		t.Fatalf("Expected %d items in Redis, got %d", maxSize, len(vals))
	}
	for i, val := range vals {
		var got message
		if err := json.Unmarshal([]byte(val), &got); err != nil {
			t.Fatalf("Could not unmarshal JSON: %v", err)
		}
		// First message in the list should be #11, then #10, ..., the last one #2.
		want := fmt.Sprintf("Message %d", maxSize+1-i)
		if got.Text != want {
			t.Errorf("Stored message text does not match; got %q, want %q", got.Text, want)
		}
	}
}

func connect(t *testing.T) *Redis {
	t.Helper()
	addr := "localhost:6379"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := Connect(ctx, addr)
	if err != nil {
		t.Fatalf("Could not connect to Redis: %v", err)
	}

	if err := r.cli.FlushAll(context.Background()).Err(); err != nil {
		t.Fatalf("Could not flush Redis: %v", err)
	}

	return r
}

func z(t *testing.T, msg message) redis.Z {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Could not marshal JSON: %v", err)
	}
	return redis.Z{
		Score:  float64(msg.CreatedAt.UnixNano()),
		Member: data,
	}
}
