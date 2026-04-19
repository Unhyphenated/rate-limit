package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal tracks how many requests have been made.
	// 'status' differentiates between 'accepted' and 'denied' requests.
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed by the rate limiter.",
		},
		[]string{"status", "endpoint"},
	)

	// RequestDuration tracks how long the rate limiting logic takes.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiter",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Latency of the rate limiting check in seconds.",
		},
		[]string{"endpoint"},
	)

	// RedisOpsTotal tracks interactions with the Redis backend.
	RedisOpsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Subsystem: "redis",
			Name:      "operations_total",
			Help:      "Total number of Redis operations performed.",
		},
		[]string{"operation", "status"},
	)

	// FailOpenTotal tracks how often the limiter failed but allowed traffic anyway.
	FailOpenTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ratelimiter",
			Subsystem: "limiter",
			Name:      "fail_open_total",
			Help:      "Total number of times the limiter failed open due to errors.",
		},
	)
)

func Init() {}