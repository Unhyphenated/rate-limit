package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/Unhyphenated/rate-limit/internal/models"
)

// MockCache implements the Cache interface for testing
type MockCache struct {
	data map[string]*models.Bucket
}

func (m *MockCache) Get(ctx context.Context, key string) (*models.Bucket, error) {
	return m.data[key], nil
}

func (m *MockCache) Set(ctx context.Context, key string, b *models.Bucket) error {
	m.data[key] = b
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error { return nil }
func (m *MockCache) Close()                                     {}

func TestLimiter_Limit(t *testing.T) {
	mock := &MockCache{data: make(map[string]*models.Bucket)}
	// Rate: 1 token/sec, Max: 5 tokens
	l := NewLimiter(mock, 1, 5)
	ctx := context.Background()
	key := "test-trader"

	t.Run("New user is allowed and gets max tokens", func(t *testing.T) {
		allowed := l.Limit(ctx, key)
		if !allowed {
			t.Error("Expected new user to be allowed")
		}
		if mock.data[key].Tokens != 4 { // 5 - 1
			t.Errorf("Expected 4 tokens left, got %d", mock.data[key].Tokens)
		}
	})

	t.Run("Depleted user is rejected", func(t *testing.T) {
		mock.data[key].Tokens = 0
		allowed := l.Limit(ctx, key)
		if allowed {
			t.Error("Expected depleted user to be rejected")
		}
	})

	t.Run("Refill works after time passes", func(t *testing.T) {
		// Set bucket to 0 tokens and 2 seconds ago
		mock.data[key].Tokens = 0
		mock.data[key].LastRefill = time.Now().Unix() - 2
		
		allowed := l.Limit(ctx, key)
		if !allowed {
			t.Error("Expected refill to allow request")
		}
	})
}