package cache

import (
	"context"

	"github.com/Unhyphenated/rate-limit/internal/models"
	// "github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (*models.Bucket, error)
	Set(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
	Close()
}