package limiter

import (
	"context"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/cache"
	"github.com/Unhyphenated/rate-limit/internal/models"
)

type Limiter struct {
	cache 		cache.Cache
	rate 		int64
	maxTokens 	int64
}

func NewLimiter(c cache.Cache, rate int64, max int64) *Limiter {
	return &Limiter{
		cache: c,
		rate: rate,
		maxTokens: max,
	}
}

func (l *Limiter) Limit(ctx context.Context, key string) bool {
	bucket, err := l.cache.Get(ctx, key)
	if err != nil {
		bucket = &models.Bucket{
			Tokens: l.maxTokens,
			LastRefill: time.Now().Unix(),
		}
	}

	tokens := l.Refill(bucket)

	if (tokens >= 1) {
		tokens -= 1
		bucket.Tokens = tokens
		bucket.LastRefill = time.Now().Unix()
		l.cache.Set(ctx, key, bucket)
		return true
	}

	return false
}

func (l *Limiter) Refill(bucket *models.Bucket) int64 {
	currentTime := time.Now().Unix()
	lastRefilled := bucket.LastRefill
	elapsed := currentTime - lastRefilled

	tokens := min(l.maxTokens, bucket.Tokens + l.rate * (elapsed))
	return tokens
}