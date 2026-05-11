package cache

import (
	"context"

	"github.com/Unhyphenated/rate-limit/internal/breaker"
	"github.com/Unhyphenated/rate-limit/internal/models"
	"github.com/redis/go-redis/v9"
)

type CircuitBreakerCache struct {
	inner   Cache
	breaker *breaker.CircuitBreaker
}

func NewCircuitBreakerCache(inner Cache, cb *breaker.CircuitBreaker) *CircuitBreakerCache {
	return &CircuitBreakerCache{
		inner:   inner,
		breaker: cb,
	}
}

func (c *CircuitBreakerCache) Get(ctx context.Context, key string) (*models.Bucket, error) {
	res, err := c.breaker.Execute(func() (any, error) {
		return c.inner.Get(ctx, key)
	})
	if err != nil {
		return nil, err
	}
	b, _ := res.(*models.Bucket)
	return b, nil
}

func (c *CircuitBreakerCache) Set(ctx context.Context, key string, bucket *models.Bucket) error {
	_, err := c.breaker.Execute(func() (any, error) {
		return nil, c.inner.Set(ctx, key, bucket)
	})
	return err
}

func (c *CircuitBreakerCache) Eval(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
	return c.breaker.Execute(func() (any, error) {
		return c.inner.Eval(ctx, script, keys, args)
	})
}

func (c *CircuitBreakerCache) Count(ctx context.Context, pattern string) (int64, error) {
	res, err := c.breaker.Execute(func() (any, error) {
		return c.inner.Count(ctx, pattern)
	})
	if err != nil {
		return 0, err
	}
	n, _ := res.(int64)
	return n, nil
}

func (c *CircuitBreakerCache) Delete(ctx context.Context, key string) error {
	_, err := c.breaker.Execute(func() (any, error) {
		return nil, c.inner.Delete(ctx, key)
	})
	return err
}

func (c *CircuitBreakerCache) Close() {
	c.inner.Close()
}
