package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type Price struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
}

func GetPrices(w http.ResponseWriter, r *http.Request) {
	data := []Price{
		{Symbol: "BTC/USD", Price: 68432.21, Change24h: 1.52},
		{Symbol: "ETH/USD", Price: 3842.15, Change24h: -0.84},
		{Symbol: "SOL/USD", Price: 145.10, Change24h: 5.12},
	}

	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().UnixMilli(),
		"data":      data,
	}

	json.NewEncoder(w).Encode(response)
}