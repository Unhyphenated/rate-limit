package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/cache"
	"github.com/Unhyphenated/rate-limit/internal/handlers"
	"github.com/Unhyphenated/rate-limit/internal/limiter"
	"github.com/Unhyphenated/rate-limit/internal/metrics"
	"github.com/Unhyphenated/rate-limit/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func main() {	
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })

	logger := slog.New(handler)
    slog.SetDefault(logger)

	slog.Info("api_node_starting", 
        slog.String("version", "1.0.0"),
        slog.Int("port", 8080),
    )
	
	rateStr := getEnv("RATE_LIMIT_FPS", "10")
	maxStr := getEnv("RATE_LIMIT_MAX", "100")

	rate, _ := strconv.ParseInt(rateStr, 10, 64)
    max, _ := strconv.ParseInt(maxStr, 10, 64)

	// Start Redis cache
	redisUrl := getEnv("REDIS_URL", "redis://localhost:6379")
	cache, err := cache.NewCache(redisUrl)
	if err != nil {
		slog.Error("failed_to_initialize_redis_cache", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer cache.Close()

	limiter := limiter.NewLimiter(cache, rate, max)
	
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/prices", middleware.RateLimit(limiter, handlers.GetPrices))

	// Handle Prometheus metrics using promhttp
	metrics.Init()
	mux.Handle("/metrics", promhttp.Handler())


	// Get the number of buckets every 30 seconds from Redis
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			count, err := cache.Count(context.Background(), "*")
			if err == nil {
				metrics.ActiveBuckets.Set(float64(count))
			}
		}
	}()

	slog.Info("server_listening", slog.Int("port", 8080))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("server_forced_to_shutdown", slog.String("error", err.Error()))
	}
}
