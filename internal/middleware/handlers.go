package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/config"
	"github.com/Unhyphenated/rate-limit/internal/limiter"
	"github.com/Unhyphenated/rate-limit/internal/metrics"
)

func RateLimit(l *limiter.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API Key
		key := r.Header.Get("X-API-KEY")
		if key == "" {
			key = getClientIP(r)
		}

		endpoint := r.URL.Path
		limits := config.GetLimits(endpoint)
		start := time.Now()

		result := l.Allow(r.Context(), key, limits.Rate, limits.MaxTokens)

		metrics.RequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds())

		if result.FailOpen {
			metrics.RequestsTotal.WithLabelValues("accepted", endpoint).Inc()
			metrics.FailOpenTotal.Inc()
			next.ServeHTTP(w, r)
			return
		}

		if !result.Allowed {
			metrics.RequestsTotal.WithLabelValues("denied", endpoint).Inc()
			w.Header().Set("Retry-After", strconv.FormatInt(result.RetryAfter, 10))
			w.WriteHeader(429)
			return
		}

		metrics.RequestsTotal.WithLabelValues("accepted", endpoint).Inc()
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