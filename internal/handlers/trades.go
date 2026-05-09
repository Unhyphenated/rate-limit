package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type Trade struct {
	ID        string  `json:"id"`
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Side      string  `json:"side"`
	Timestamp int64   `json:"timestamp"`
}

type TradesResponse struct {
	Trades []Trade `json:"trades"`
	Total  int     `json:"total"`
}

func GetTrades(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Unix()
	
	trades := []Trade{
		{
			ID:        "t1001",
			Symbol:    "BTCUSDT",
			Price:     67234.50,
			Quantity:  0.245,
			Side:      "buy",
			Timestamp: now - 120,
		},
		{
			ID:        "t1002",
			Symbol:    "ETHUSDT",
			Price:     3456.78,
			Quantity:  2.5,
			Side:      "sell",
			Timestamp: now - 90,
		},
		{
			ID:        "t1003",
			Symbol:    "SOLUSDT",
			Price:     142.35,
			Quantity:  15.0,
			Side:      "buy",
			Timestamp: now - 60,
		},
		{
			ID:        "t1004",
			Symbol:    "BTCUSDT",
			Price:     67189.25,
			Quantity:  0.5,
			Side:      "sell",
			Timestamp: now - 30,
		},
		{
			ID:        "t1005",
			Symbol:    "ETHUSDT",
			Price:     3462.10,
			Quantity:  1.25,
			Side:      "buy",
			Timestamp: now - 10,
		},
	}

	response := TradesResponse{
		Trades: trades,
		Total:  len(trades),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
