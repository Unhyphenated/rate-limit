package limiter

import (
	"context"
	"log/slog"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/cache"
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

func (l *Limiter) Limit(ctx context.Context, key string) bool {
	result, err := l.cache.Eval(ctx, getLimit, []string{key}, []any{l.rate, l.maxTokens, time.Now().Unix()})
	if err != nil {
		slog.Warn("rate_limit_script_failed", slog.String("key", key), slog.Any("error", err))
		return true
	}
	allowed, ok := result.(int64)
	if !ok {
		slog.Warn("unexpected_script_result_type", slog.Any("result", result))
		return true
	}
	return allowed == 1
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

	if bucket["tokens"] >= 1 then
		bucket["tokens"] = bucket["tokens"] - 1
		bucket["last_refill"] = currentTime
		redis.call("HSET", key, "tokens", bucket["tokens"], "last_refill", bucket["last_refill"])
		redis.call("EXPIRE", key, 3600)
		return 1
	end

	return 0
`)