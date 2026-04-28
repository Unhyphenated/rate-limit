package config

type EndpointLimits struct {
	Rate int64
	MaxTokens int64
}

var LimitConfig = map[string]EndpointLimits{
    "/api/v1/prices": {Rate: 2, MaxTokens: 1000},
    "/api/v1/trades": {Rate: 2, MaxTokens: 500},
    "/api/v1/orders": {Rate: 2, MaxTokens: 100},
    "/api/v1/wallet": {Rate: 1, MaxTokens: 50},
}

func GetLimits(endpoint string) EndpointLimits {
	if limits, exists := LimitConfig[endpoint]; exists {
		return limits
	}
	return EndpointLimits{Rate: 10, MaxTokens: 50}
}