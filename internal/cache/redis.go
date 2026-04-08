package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Unhyphenated/rate-limit/internal/models"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (*models.Bucket, error)
	Set(ctx context.Context, key string, bucket *models.Bucket) error
	Eval(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error)
	Delete(ctx context.Context, key string) error
	Close()
}

type Redis struct {
	Client *redis.Client
}

func NewCache(redisURL string) (Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	slog.Info("redis_client_initialized")

	return &Redis{Client: client}, nil
}

func (c *Redis) Get(ctx context.Context, key string) (*models.Bucket, error) {
	res := c.Client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to get value from Redis: %w", err)
	}

	if len(res.Val()) == 0 {
		return nil, nil
	}

	var bucket models.Bucket
	if err := res.Scan(&bucket); err != nil {
		return nil, fmt.Errorf("failed to scan redis hash: %w", err)
	}

	return &bucket, nil
}

func (c *Redis) Set(ctx context.Context, key string, bucket *models.Bucket) error {
	err := c.Client.HSet(ctx, key, bucket).Err()
	if err != nil {
        return fmt.Errorf("failed to set bucket: %w", err)
    }
    return nil
}

func (c *Redis) Eval(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
	res, err := script.Run(ctx, c.Client, keys, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to run Lua script: %w", err)
	}
	return res, nil
}

func (c *Redis) Delete(ctx context.Context, key string) error {
	_, err := c.Client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}
	return nil
}

func (c *Redis) Close() {
	err := c.Client.Close()
	if err != nil {
		slog.Error("redis_client_close_failed", slog.Any("error", err))
		return
	}
	slog.Info("redis_client_closed")
}
