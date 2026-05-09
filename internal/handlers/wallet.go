package handlers

import (
	"encoding/json"
	"net/http"
)

type Balance struct {
	Asset     string  `json:"asset"`
	Free      float64 `json:"free"`
	Locked    float64 `json:"locked"`
	Total     float64 `json:"total"`
	USDValue  float64 `json:"usd_value"`
}

type WalletResponse struct {
	Balances     []Balance `json:"balances"`
	TotalUSD     float64   `json:"total_usd"`
	LastUpdated  int64     `json:"last_updated"`
}

func GetWallet(w http.ResponseWriter, r *http.Request) {
	balances := []Balance{
		{
			Asset:    "BTC",
			Free:     1.5432,
			Locked:   0.25,
			Total:    1.7932,
			USDValue: 120567.89,
		},
		{
			Asset:    "ETH",
			Free:     15.234,
			Locked:   2.0,
			Total:    17.234,
			USDValue: 59612.45,
		},
		{
			Asset:    "USDT",
			Free:     25000.00,
			Locked:   0.0,
			Total:    25000.00,
			USDValue: 25000.00,
		},
		{
			Asset:    "SOL",
			Free:     150.5,
			Locked:   10.0,
			Total:    160.5,
			USDValue: 22847.18,
		},
		{
			Asset:    "BNB",
			Free:     50.25,
			Locked:   5.0,
			Total:    55.25,
			USDValue: 16575.00,
		},
	}

	totalUSD := 0.0
	for _, balance := range balances {
		totalUSD += balance.USDValue
	}

	response := WalletResponse{
		Balances:    balances,
		TotalUSD:    totalUSD,
		LastUpdated: 1713484800, // Static timestamp for demo
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
