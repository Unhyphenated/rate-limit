package limiter

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/cache"
	"github.com/Unhyphenated/rate-limit/internal/models"
	"github.com/redis/go-redis/v9"
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

func (l *Limiter) Allow(ctx context.Context, key string) *models.RateLimitResult {
	rlResult := &models.RateLimitResult{
		Allowed: true, 
		Limit: l.maxTokens,
		Remaining: 0,
		ResetAt: 0,
		RetryAfter: 0,
		FailOpen: false,
	}

	result, err := l.cache.Eval(ctx, getLimit, []string{key}, []any{l.rate, l.maxTokens, time.Now().Unix()})
	// Fail-open 
	if err != nil {
		slog.Warn("rate_limit_script_failed", slog.String("key", key), slog.Any("error", err))
		rlResult.FailOpen = true
		return rlResult
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		slog.Warn("unexpected_script_result_type", slog.Any("result", result))
		rlResult.FailOpen = true
		return rlResult
	}

	rlResult.Allowed = values[0].(int64) == 1
	rlResult.Remaining = values[1].(int64)
	rlResult.ResetAt = values[2].(int64)

	if !rlResult.Allowed {
		rlResult.RetryAfter = int64(math.Ceil(1.0 / float64(l.rate)))
	}

	return rlResult
}

var getLimit = redis.NewScript(`
	local key = KEYS[1]
	local rate = tonumber(ARGV[1])
	local maxTokens = tonumber(ARGV[2])
	local currentTime = tonumber(ARGV[3])

	local data = redis.call("HGETALL", key)
	local bucket = {}
	
	if #data == 0 then
		bucket["tokens"] = maxTokens
		bucket["last_refill"] = currentTime
	else
		for i = 1, #data, 2 do
			bucket[data[i]] = tonumber(data[i + 1])
		end
	end

	local elapsed = currentTime - bucket["last_refill"]
	bucket["tokens"] = math.min(maxTokens, bucket["tokens"] + (elapsed * rate))

	local tokensNeeded = maxTokens - bucket["tokens"]
	local resetAt = currentTime + math.ceil(tokensNeeded / rate)

	if bucket["tokens"] >= 1 then
		bucket["tokens"] = bucket["tokens"] - 1
		bucket["last_refill"] = currentTime
		redis.call("HSET", key, "tokens", bucket["tokens"], "last_refill", bucket["last_refill"])
		redis.call("EXPIRE", key, 3600)

		return {1, bucket["tokens"], resetAt}
	end

	local timeToNextToken = currentTime + math.ceil(1 / rate)

	return {0, 0, timeToNextToken}
`)