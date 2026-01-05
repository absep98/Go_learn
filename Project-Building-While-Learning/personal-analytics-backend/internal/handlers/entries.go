package handlers

import (
	"encoding/json"
	"net/http"
	"personal-analytics-backend/internal/db"
)

// CreateEntryRequest represents the incoming request body
type CreateEntryRequest struct {
	UserID int    `json:"user_id"`
	Text   string `json:"text"`
	Mood   int    `json:"mood"`
}

// The json:"..." tags:

// Tell Go: "When JSON comes in, put user_id into the UserID field"
// It's a translator between JSON (what clients send) and Go structs (what our code uses)
// Why we need this:
// Without it, Go wouldn't know how to convert JSON into a Go struct.
// CreateEntryResponse represents the API response
type CreateEntryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      int64  `json:"id,omitempty"`
}

// The omitempty tag:

// Means: "If ID is 0 (empty), don't include it in the JSON"
// Used for error responses where there's no ID

// CreateEntry handles POST /entries
func CreateEntry(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req CreateEntryRequest

	// "Read the JSON data from the request and put it into our req variable."
	// Step-by-step:

	// r.Body = the data the client sent
	// json.NewDecoder(r.Body) = create a JSON reader
	// .Decode(&req) = convert JSON into our struct, & means put the decoded data directly into req not a copy.
	// if it fails : JSON malformed typo, mussing comma, etc or send back error "invalid request body"
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, CreateEntryResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate input
	// checks if userid is valid if not reject it
	// UserID of 0 or -ve makes no sense text being empty is useless mood outisde 1-10 is invalid
	if req.UserID <= 0 {
		respondJSON(w, http.StatusBadRequest, CreateEntryResponse{
			Success: false,
			Message: "user_id must be positive",
		})
		return
	}

	if req.Text == "" {
		respondJSON(w, http.StatusBadRequest, CreateEntryResponse{
			Success: false,
			Message: "text cannot be empty",
		})
		return
	}

	if req.Mood < 1 || req.Mood > 10 {
		respondJSON(w, http.StatusBadRequest, CreateEntryResponse{
			Success: false,
			Message: "mood must be between 1 and 10",
		})
		return
	}

	// Insert into database
	id, err := db.InsertEntry(req.UserID, req.Text, req.Mood)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, CreateEntryResponse{
			Success: false,
			Message: "Failed to save entry",
		})
		return
	}
	// All above are checks if passed then only allow to save it
	// Success response
	respondJSON(w, http.StatusCreated, CreateEntryResponse{
		Success: true,
		Message: "Entry created successfully",
		ID:      id,
	})
}

// GetEntries handles GET /entries
func GetEntries(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all entries from database
	entries, err := db.GetAllEntries()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to fetch entries",
		})
		return
	}

	// Handle empty case - return empty array instead of null
	if entries == nil {
		entries = []map[string]interface{}{}
	}

	// Success response with entries
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(entries),
		"entries": entries,
	})
}

// Helper function to send JSON responses
// "Sends back a JSON response with a specific status code."
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	// Tell the client the response i'm sending is json
	w.Header().Set("Content-Type", "application/json")
	// set the response code 200, 400 etc
	w.WriteHeader(status)
	// Covert our go struct to json and send it to client
	json.NewEncoder(w).Encode(data)
}
