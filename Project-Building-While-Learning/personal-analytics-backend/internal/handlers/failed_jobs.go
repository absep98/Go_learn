package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"personal-analytics-backend/internal/db"
)

func GetFailedJobs(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	failedJobs, err := db.GetFailedJobs()
	if err != nil {
		slog.Error("failed to get failed jobs", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(failedJobs); err != nil {
		slog.Error("failed to encode failed jobs response", "error", err)
	}
}