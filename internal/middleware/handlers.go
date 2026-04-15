package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Unhyphenated/rate-limit/internal/limiter"
)

func RateLimit(l *limiter.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API Key
		key := r.Header.Get("X-API-KEY")
		if key == "" {
			key = getClientIP(r)
		}

		result := l.Allow(r.Context(), key)

		if result.FailOpen {
			next.ServeHTTP(w, r)
			return
		}

		if !result.Allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(result.RetryAfter, 10))
			w.WriteHeader(429)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt, 10))

		next.ServeHTTP(w, r)
	}
}

func getClientIP(r *http.Request) string {
	numProxies := 1
	xff := r.Header.Get("X-Forwarded-For")

	if xff != "" {
		parts := strings.Split(xff, ",")

		targetIndex := len(parts) - numProxies
		if targetIndex >= 0 {
			return strings.TrimSpace(parts[targetIndex])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
	return host
}