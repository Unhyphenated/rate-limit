package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Filled    float64 `json:"filled"`
	Status    string  `json:"status"`
	CreatedAt int64   `json:"created_at"`
}

type OrdersResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Unix()
	
	orders := []Order{
		{
			OrderID:   "ord_123456",
			Symbol:    "BTCUSDT",
			Side:      "buy",
			Type:      "limit",
			Price:     66500.00,
			Quantity:  0.5,
			Filled:    0.0,
			Status:    "open",
			CreatedAt: now - 3600,
		},
		{
			OrderID:   "ord_123457",
			Symbol:    "ETHUSDT",
			Side:      "sell",
			Type:      "limit",
			Price:     3500.00,
			Quantity:  2.0,
			Filled:    0.0,
			Status:    "open",
			CreatedAt: now - 1800,
		},
		{
			OrderID:   "ord_123458",
			Symbol:    "SOLUSDT",
			Side:      "buy",
			Type:      "market",
			Price:     0.0,
			Quantity:  10.0,
			Filled:    10.0,
			Status:    "filled",
			CreatedAt: now - 600,
		},
		{
			OrderID:   "ord_123459",
			Symbol:    "BTCUSDT",
			Side:      "buy",
			Type:      "limit",
			Price:     67000.00,
			Quantity:  0.25,
			Filled:    0.1,
			Status:    "partial",
			CreatedAt: now - 300,
		},
	}

	response := OrdersResponse{
		Orders: orders,
		Total:  len(orders),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
