package middleware

import (
	"net"
	"net/http"

	"github.com/Unhyphenated/rate-limit/internal/limiter"
)

func RateLimit(l *limiter.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API Key
		key := r.Header.Get("X-API-KEY")
		if key == "" {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			key = host
		}

		if !l.Limit(r.Context(), key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit exceeded", "retry_after": 60}`))
			return
		}

		next.ServeHTTP(w, r)
	}
}