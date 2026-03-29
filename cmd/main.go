package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Unhyphenated/rate-limit/internal/cache"
	"github.com/Unhyphenated/rate-limit/internal/limiter"
	"github.com/Unhyphenated/rate-limit/internal/middleware"
)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func getPrices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    // Mock data for a trading platform
    prices := `{"BTC_USD": 65000.50, "ETH_USD": 3500.20, "timestamp": 1710712200}`
    w.Write([]byte(prices))
}

func main() {	
	rateStr := getEnv("RATE_LIMIT_FPS", "10")
	maxStr := getEnv("RATE_LIMIT_MAX", "100")

	rate, _ := strconv.ParseInt(rateStr, 10, 64)
    max, _ := strconv.ParseInt(maxStr, 10, 64)

	// Start Redis cache
	redisUrl := getEnv("REDIS_URL", "redis://localhost:6379")
	cache, err := cache.NewCache(redisUrl)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}

	defer cache.Close()

	limiter := limiter.NewLimiter(cache, rate, max)
	
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/prices", middleware.RateLimit(limiter, getPrices))

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
