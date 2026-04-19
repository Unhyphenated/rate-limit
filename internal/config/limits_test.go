package config

import "testing"

func TestGetLimits_ValidEndpoints(t *testing.T) {
	tests := []struct {
		endpoint  string
		wantRate  int64
		wantMax   int64
	}{
		{"/api/v1/prices", 100, 1000},
		{"/api/v1/trades", 50, 500},
		{"/api/v1/orders", 10, 100},
		{"/api/v1/wallet", 5, 50},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint, func(t *testing.T) {
			limits := GetLimits(tt.endpoint)
			if limits.Rate != tt.wantRate {
				t.Errorf("GetLimits(%s).Rate = %d, want %d", tt.endpoint, limits.Rate, tt.wantRate)
			}
			if limits.MaxTokens != tt.wantMax {
				t.Errorf("GetLimits(%s).MaxTokens = %d, want %d", tt.endpoint, limits.MaxTokens, tt.wantMax)
			}
		})
	}
}

func TestGetLimits_UnknownEndpoint(t *testing.T) {
	limits := GetLimits("/api/v1/unknown")
	
	// Should return fallback
	if limits.Rate != 10 {
		t.Errorf("GetLimits(unknown).Rate = %d, want 10 (fallback)", limits.Rate)
	}
	if limits.MaxTokens != 50 {
		t.Errorf("GetLimits(unknown).MaxTokens = %d, want 50 (fallback)", limits.MaxTokens)
	}
}

func TestLimitConfig_AllEndpointsDefined(t *testing.T) {
	requiredEndpoints := []string{
		"/api/v1/prices",
		"/api/v1/trades",
		"/api/v1/orders",
		"/api/v1/wallet",
	}

	for _, endpoint := range requiredEndpoints {
		if _, exists := LimitConfig[endpoint]; !exists {
			t.Errorf("LimitConfig missing required endpoint: %s", endpoint)
		}
	}
}
