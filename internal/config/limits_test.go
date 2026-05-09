package config

import "testing"

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
