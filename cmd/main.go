package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Unhyphenated/rate-limit/internal/cache"
	"github.com/Unhyphenated/rate-limit/internal/limiter"
)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
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
	
	fmt.Println("Sever listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
