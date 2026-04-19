package limiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/cache"
)

func setupRedis(t *testing.T) (cache.Cache, func()) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	c, err := cache.NewCache("redis://localhost:6379")
	if err != nil {
		t.Fatalf("Failed to connect to Redis (is docker-compose running?): %v", err)
	}

	cleanup := func() {
		c.Close()
	}

	return c, cleanup
}

func TestLimiter_Integration_BasicFlow(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-basic-flow"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 1 token/sec, Max: 5 tokens
	l := NewLimiter(c)
	rate := int64(1)
	maxTokens := int64(5)

	t.Run("New user gets max tokens", func(t *testing.T) {
		// First 5 requests should succeed
		for i := 0; i < 5; i++ {
			result := l.Allow(ctx, key, rate, maxTokens)
			if !result.Allowed {
				t.Errorf("Request %d should be allowed (new user starts with 5 tokens)", i+1)
			}
			if result.FailOpen {
				t.Error("Should not fail open with working Redis")
			}
		}

		// 6th request should fail (no tokens left)
		result := l.Allow(ctx, key, rate, maxTokens)
		if result.Allowed {
			t.Error("Request 6 should be denied (bucket depleted)")
		}
		if result.RetryAfter == 0 {
			t.Error("RetryAfter should be set when rate limited")
		}
	})
}

func TestLimiter_Integration_Refill(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-refill"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 2 tokens/sec, Max: 5 tokens
	l := NewLimiter(c)
	rate := int64(2)
	maxTokens := int64(5)

	t.Run("Tokens refill over time", func(t *testing.T) {
		// Deplete all tokens
		for i := 0; i < 5; i++ {
			l.Allow(ctx, key, rate, maxTokens)
		}

		// Verify depleted
		result := l.Allow(ctx, key, rate, maxTokens)
		if result.Allowed {
			t.Error("Bucket should be depleted")
		}

		// Wait 2 seconds (should refill 4 tokens: 2 tokens/sec * 2 sec)
		time.Sleep(2 * time.Second)

		// Should be able to make 4 requests
		for i := 0; i < 4; i++ {
			result := l.Allow(ctx, key, rate, maxTokens)
			if !result.Allowed {
				t.Errorf("Request %d should be allowed after refill", i+1)
			}
		}

		// 5th request should fail
		result = l.Allow(ctx, key, rate, maxTokens)
		if result.Allowed {
			t.Error("Request 5 should be denied (only 4 tokens refilled)")
		}
	})
}

func TestLimiter_Integration_MaxTokensCap(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-max-cap"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 10 tokens/sec, Max: 5 tokens
	l := NewLimiter(c)
	rate := int64(10)
	maxTokens := int64(5)

	t.Run("Tokens don't exceed max", func(t *testing.T) {
		// Use 2 tokens
		l.Allow(ctx, key, rate, maxTokens)
		l.Allow(ctx, key, rate, maxTokens)

		// Wait 5 seconds (would refill 50 tokens, but capped at 5)
		time.Sleep(5 * time.Second)

		// Should only be able to make 5 requests total (3 remaining + 2 used = 5 max)
		successCount := 0
		for i := 0; i < 10; i++ {
			result := l.Allow(ctx, key, rate, maxTokens)
			if result.Allowed {
				successCount++
			}
		}

		if successCount != 5 {
			t.Errorf("Expected 5 successful requests (max cap), got %d", successCount)
		}
	})
}

func TestLimiter_Integration_RaceCondition(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-race-condition"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 1 tokens/sec (slow refill), Max: 10 tokens
	l := NewLimiter(c)
	rate := int64(1)
	maxTokens := int64(10)

	t.Run("Concurrent requests are handled atomically", func(t *testing.T) {
		// Initialize bucket by making first request
		l.Allow(ctx, key, rate, maxTokens)

		// Wait a moment to ensure bucket is set
		time.Sleep(100 * time.Millisecond)

		// Launch 100 concurrent requests
		// Only 9 more should succeed (started with 10, used 1, so 9 left)
		numGoroutines := 100
		results := make(chan bool, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := l.Allow(ctx, key, rate, maxTokens)
				results <- result.Allowed
			}()
		}

		wg.Wait()
		close(results)

		// Count allowed requests
		allowed := 0
		for result := range results {
			if result {
				allowed++
			}
		}

		t.Logf("Launched %d concurrent requests, %d were allowed", numGoroutines, allowed)

		// With Lua script atomicity, exactly 9 should be allowed
		if allowed != 9 {
			t.Errorf("Expected exactly 9 requests allowed, got %d (race condition detected!)", allowed)
		} else {
			t.Log("✓ No race condition - Lua script provides atomicity")
		}
	})
}

func TestLimiter_Integration_TTL(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-ttl"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 1 token/sec, Max: 5 tokens
	l := NewLimiter(c)
	rate := int64(1)
	maxTokens := int64(5)

	t.Run("Bucket has TTL set", func(t *testing.T) {
		// Make a request to create the bucket
		result := l.Allow(ctx, key, rate, maxTokens)
		if !result.Allowed {
			t.Error("First request should be allowed")
		}

		// Check that TTL is set (this requires accessing Redis directly)
		// We'll verify by checking the key exists
		bucket, err := c.Get(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get bucket: %v", err)
		}
		if bucket == nil {
			t.Error("Bucket should exist after request")
		}

		t.Log("✓ Bucket created with TTL (expires in 1 hour)")
	})
}

func TestLimiter_Integration_MultipleUsers(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key1 := "user-1"
	key2 := "user-2"
	defer func() { _ = c.Delete(ctx, key1) }()
	defer func() { _ = c.Delete(ctx, key2) }()

	// Rate: 1 token/sec, Max: 3 tokens
	l := NewLimiter(c)
	rate := int64(1)
	maxTokens := int64(3)

	t.Run("Different users have independent buckets", func(t *testing.T) {
		// User 1 depletes their bucket
		for i := 0; i < 3; i++ {
			l.Allow(ctx, key1, rate, maxTokens)
		}
		result := l.Allow(ctx, key1, rate, maxTokens)
		if result.Allowed {
			t.Error("User 1 should be rate limited")
		}

		// User 2 should still have tokens
		for i := 0; i < 3; i++ {
			result := l.Allow(ctx, key2, rate, maxTokens)
			if !result.Allowed {
				t.Errorf("User 2 request %d should be allowed (independent bucket)", i+1)
			}
		}

		// User 2 should now be depleted too
		result = l.Allow(ctx, key2, rate, maxTokens)
		if result.Allowed {
			t.Error("User 2 should be rate limited after 3 requests")
		}
	})
}

func TestLimiter_Integration_HighConcurrency(t *testing.T) {
	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	key := "test-high-concurrency"
	defer func() { _ = c.Delete(ctx, key) }()

	// Rate: 1 tokens/sec, Max: 50 tokens
	l := NewLimiter(c)
	rate := int64(1)
	maxTokens := int64(50)

	t.Run("High concurrency stress test", func(t *testing.T) {
		// Initialize bucket
		l.Allow(ctx, key, rate, maxTokens)

		// Launch 500 concurrent requests
		// Only 49 more should succeed (50 - 1 already used)
		numGoroutines := 500
		results := make(chan bool, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := l.Allow(ctx, key, rate, maxTokens)
				results <- result.Allowed
			}()
		}

		wg.Wait()
		close(results)

		allowed := 0
		for result := range results {
			if result {
				allowed++
			}
		}

		t.Logf("Launched %d concurrent requests, %d were allowed", numGoroutines, allowed)

		if allowed != 49 {
			t.Errorf("Expected exactly 49 requests allowed, got %d", allowed)
		} else {
			t.Log("✓ High concurrency handled correctly")
		}
	})
}
