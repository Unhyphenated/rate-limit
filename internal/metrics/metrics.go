package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/collectors"
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
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
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

	RedisOpsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ratelimiter",
			Subsystem: "redis",
			Name: "operation_duration_seconds",
			Help: "Duration of Redis operations",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
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

	// ActiveBuckets tracks the number of active token buckets in Redis.
	ActiveBuckets = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ratelimiter",
			Subsystem: "limiter",
			Name:      "active_buckets",
			Help:      "Total number of active token buckets in Redis.",
		},
	)
)

func Init() {
	prometheus.MustRegister(collectors.NewGoCollector())
	prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}
