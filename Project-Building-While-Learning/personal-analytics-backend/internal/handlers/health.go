package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/redis"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`   // "healthy" or "unhealthy"
	Database string `json:"database"` // "connected" or "disconnected"
	Redis    string `json:"redis"`    // "connected" or "disconnected"
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:   "healthy",
		Database: "connected",
		Redis:    "connected",
	}

	// Check Redis: Use existing connection, just ping it
	err := redis.Client.Ping(context.Background()).Err()
	if err != nil {
		response.Redis = "disconnected"
		response.Status = "unhealthy"
	}

	// Check Database: Ping the existing connection
	err = db.DB.Ping()
	if err != nil {
		response.Database = "disconnected"
		response.Status = "unhealthy"
	}

	// Set response headers and status code
	w.Header().Set("Content-Type", "application/json")
	if response.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable) // 503
	} else {
		w.WriteHeader(http.StatusOK) // 200
	}

	// Send JSON response
	json.NewEncoder(w).Encode(response)
}
