package handlers

import (
	"encoding/json"
    "net/http"
)


func GetMetrics(w http.ResponseWriter, r *http.Request) {

	// Step 1: Check if method is GET (return 405 if not)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Call AppMetrics.GetSnapshot() to get the data

	getMetricsSnapshot := AppMetrics.GetSnapshot()

	// Step 3: Set Content-Type header to "application/json"
	w.Header().Set("Content-Type", "application/json")

	// Step 4: Encode data as JSON using json.NewEncoder(w).Encode(data)
	json.NewEncoder(w).Encode(getMetricsSnapshot)
}
