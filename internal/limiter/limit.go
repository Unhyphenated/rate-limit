package limiter

import (
	"github.com/Unhyphenated/rate-limit/internal/cache"
)


type Limiter struct {
	cache 		cache.Cache
	rate 		int
	maxTokens 	int
}

func NewLimiter(c cache.Cache, rate int, max int) *Limiter {
	return &Limiter{
		cache: c,
		rate: rate,
		maxTokens: max,
	}
}