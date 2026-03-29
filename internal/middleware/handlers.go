package middleware

import (
	"net"
	"net/http"
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

		if !l.Limit(r.Context(), key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit exceeded", "retry_after": 60}`))
			return
		}

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