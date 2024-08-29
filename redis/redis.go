package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/redis/go-redis/v9"
)

// Redis provides caching in Redis.
type Redis struct {
	cli *redis.Client
}

// Connect connects to the Redis server and pings the server to ensure the
// connection is working.
func Connect(ctx context.Context, addr string) (*Redis, error) {
	cli := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Redis{
		cli: cli,
	}, nil
}

const (
	messagePrefix = "messages:"
	maxSize       = 10
)

// ListMessages returns a list of message from Redis. The messages are sorted
// by the timestamp in descending order.
func (r *Redis) ListMessages(ctx context.Context, limit int) ([]api.Message, error) {
	vals, err := r.cli.ZRevRange(ctx, messagePrefix, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("zrange: %w", err)
	}

	out := make([]api.Message, len(vals))
	for i, val := range vals {
		var msg message
		if err := json.Unmarshal([]byte(val), &msg); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		out[i] = msg.APIMessage()
	}

	return out, nil
}

// InsertMessage adds the message to Redis sorted set.
func (r *Redis) InsertMessage(ctx context.Context, msg api.Message) error {
	m := message(msg)
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err = r.cli.ZAdd(ctx, messagePrefix, redis.Z{
		Score:  float64(msg.CreatedAt.UnixNano()),
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("zadd: %w", err)
	}
	if err = r.cli.ZRemRangeByRank(ctx, messagePrefix, 0, int64(-maxSize-1)).Err(); err != nil {
		return fmt.Errorf("zremrangebyrank: %w", err)
	}
	return nil
}
