package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/breaker"
	"github.com/Unhyphenated/rate-limit/internal/models"
	"github.com/redis/go-redis/v9"
)

type stubCache struct {
	evalCalls atomic.Int32
	closeN    atomic.Int32
	evalFn    func(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error)
	countFn   func(ctx context.Context, pattern string) (int64, error)
}

func (s *stubCache) Get(ctx context.Context, key string) (*models.Bucket, error) {
	return nil, errors.New("not implemented")
}

func (s *stubCache) Set(ctx context.Context, key string, bucket *models.Bucket) error {
	return errors.New("not implemented")
}

func (s *stubCache) Eval(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
	s.evalCalls.Add(1)
	if s.evalFn != nil {
		return s.evalFn(ctx, script, keys, args)
	}
	return nil, nil
}

func (s *stubCache) Count(ctx context.Context, pattern string) (int64, error) {
	if s.countFn != nil {
		return s.countFn(ctx, pattern)
	}
	return 0, errors.New("not implemented")
}

func (s *stubCache) Delete(ctx context.Context, key string) error {
	return errors.New("not implemented")
}

func (s *stubCache) Close() {
	s.closeN.Add(1)
}

func TestCircuitBreakerCache_EvalDoesNotReachInnerWhenOpen(t *testing.T) {
	cb := breaker.NewCircuitBreaker(breaker.Config{
		FailureThreshold:  2,
		CooldownPeriod:    time.Hour,
		HalfOpenSuccesses: 1,
	})
	inner := &stubCache{
		evalFn: func(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
			return nil, errors.New("down")
		},
	}
	wrapped := NewCircuitBreakerCache(inner, cb)
	ctx := context.Background()
	script := redis.NewScript(`return 1`)

	for i := 0; i < 2; i++ {
		_, err := wrapped.Eval(ctx, script, nil, nil)
		if err == nil || errors.Is(err, breaker.ErrCircuitOpen) {
			t.Fatalf("iter %d: expected inner error, got %v", i, err)
		}
	}

	_, err := wrapped.Eval(ctx, script, nil, nil)
	if !errors.Is(err, breaker.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if inner.evalCalls.Load() != 2 {
		t.Fatalf("expected 2 inner Evals, got %d", inner.evalCalls.Load())
	}
}

func TestCircuitBreakerCache_EvalPropagatesResult(t *testing.T) {
	cb := breaker.NewCircuitBreaker(breaker.Config{
		FailureThreshold:  10,
		CooldownPeriod:    time.Second,
		HalfOpenSuccesses: 1,
	})
	inner := &stubCache{
		evalFn: func(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
			return []any{int64(1)}, nil
		},
	}
	wrapped := NewCircuitBreakerCache(inner, cb)
	ctx := context.Background()
	script := redis.NewScript(`return 1`)

	v, err := wrapped.Eval(ctx, script, []string{"k"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	slc, ok := v.([]any)
	if !ok || len(slc) != 1 || slc[0].(int64) != 1 {
		t.Fatalf("unexpected value %v", v)
	}
	if inner.evalCalls.Load() != 1 {
		t.Fatalf("inner calls: %d", inner.evalCalls.Load())
	}
}

func TestCircuitBreakerCache_CloseBypassesBreaker(t *testing.T) {
	cb := breaker.NewCircuitBreaker(breaker.Config{
		FailureThreshold:  1,
		CooldownPeriod:    time.Hour,
		HalfOpenSuccesses: 1,
	})
	ctx := context.Background()
	script := redis.NewScript(`return 1`)

	inner := &stubCache{
		evalFn: func(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
			return nil, errors.New("down")
		},
	}
	wrapped := NewCircuitBreakerCache(inner, cb)

	_, _ = wrapped.Eval(ctx, script, nil, nil)
	wrapped.Close()
	if inner.closeN.Load() != 1 {
		t.Fatalf("expected inner Close once, got %d", inner.closeN.Load())
	}

	cb2 := breaker.NewCircuitBreaker(breaker.Config{
		FailureThreshold:  1,
		CooldownPeriod:    time.Hour,
		HalfOpenSuccesses: 1,
	})
	inner2 := &stubCache{
		evalFn: func(ctx context.Context, script *redis.Script, keys []string, args []any) (any, error) {
			return nil, errors.New("down")
		},
	}
	wrappedOpen := NewCircuitBreakerCache(inner2, cb2)
	_, _ = wrappedOpen.Eval(ctx, script, nil, nil)
	_, err := wrappedOpen.Eval(ctx, script, nil, nil)
	if !errors.Is(err, breaker.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen when open: %v", err)
	}
	if inner2.evalCalls.Load() != 1 {
		t.Fatalf("expected 1 inner Eval (second short-circuited), got %d", inner2.evalCalls.Load())
	}
	wrappedOpen.Close()
	if inner2.closeN.Load() != 1 {
		t.Fatal("Close should reach inner while breaker open")
	}
}

func TestCircuitBreakerCache_CountDelegates(t *testing.T) {
	cb := breaker.NewCircuitBreaker(breaker.Config{
		FailureThreshold:  5,
		CooldownPeriod:    time.Second,
		HalfOpenSuccesses: 1,
	})
	var countCalls atomic.Int32
	inner := &stubCache{
		countFn: func(ctx context.Context, pattern string) (int64, error) {
			countCalls.Add(1)
			return 42, nil
		},
	}

	wrapped := NewCircuitBreakerCache(inner, cb)
	ctx := context.Background()

	n, err := wrapped.Count(ctx, "*")
	if err != nil {
		t.Fatal(err)
	}
	if n != 42 {
		t.Fatalf("got %d", n)
	}
	if countCalls.Load() != 1 {
		t.Fatalf("inner Count calls: %d", countCalls.Load())
	}
}
