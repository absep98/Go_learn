package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"personal-analytics-backend/internal/db"
	"strconv"
)

// CreateEntryRequest represents the incoming request body
type CreateEntryRequest struct {
	Text     string `json:"text"`
	Mood     int    `json:"mood"`
	Category string `json:"category"`
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

// errorResponse is a helper function to send error responses consistently
// Reduces code duplication across all validation checks
func errorResponse(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, CreateEntryResponse{
		Success: false,
		Message: message,
	})
}

// CreateEntry handles POST /entries
func CreateEntry(w http.ResponseWriter, r *http.Request) {
	log.Println("📨 POST /entries - Request received")

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, "Invalid authentication data")
		return
	}

	log.Printf("🔍 Creating entry for user ID: %d", userID)

	// Parse JSON request body
	var req CreateEntryRequest

	// "Read the JSON data from the request and put it into our req variable."
	// Step-by-step:

	// r.Body = the data the client sent
	// json.NewDecoder(r.Body) = create a JSON reader
	// .Decode(&req) = convert JSON into our struct, & means put the decoded data directly into req not a copy.
	// if it fails : JSON malformed typo, mussing comma, etc or send back error "invalid request body"
	err := json.NewDecoder(r.Body).Decode(&req)

	// Validate input
	// checks if userid is valid if not reject it
	// UserID of 0 or -ve makes no sense text being empty is useless mood outisde 1-10 is invalid
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Text == "" {
		errorResponse(w, http.StatusBadRequest, "text cannot be empty")
		return
	}

	if req.Mood < 1 || req.Mood > 10 {
		errorResponse(w, http.StatusBadRequest, "mood must be between 1 and 10")
		return
	}

	if req.Category == "" {
		errorResponse(w, http.StatusBadRequest, "category cannot be empty")
		return
	}

	// Insert into database
	id, err := db.InsertEntry(int(userID), req.Text, req.Mood, req.Category)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to save entry")
		return
	}
	// All above are checks if passed then only allow to save it
	// Success response
	log.Printf("✅ Entry created successfully with ID: %d", id)
	respondJSON(w, http.StatusCreated, CreateEntryResponse{
		Success: true,
		Message: "Entry created successfully",
		ID:      id,
	})
}

// GetEntries handles GET /entries
// Returns entries for the authenticated user only
func GetEntries(w http.ResponseWriter, r *http.Request) {
	log.Println("📚 GET /entries - Request received")

	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user_id from context (middleware puts it there)
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		// This shouldn't happen if middleware is working, but check anyway
		respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		log.Printf("❌ Failed to extract user_id from context: %v", userIDValue)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Invalid authentication data",
		})
		return
	}

	log.Printf("🔍 Fetching entries for user ID: %d", userID)

	// Get entries for this user only
	entries, err := db.GetEntriesByUser(userID)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
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
	log.Printf("✅ Returning %d entries for user %d", len(entries), userID)
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

func UpdateEntry(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entryId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("❌ Invalid entry ID: %v", err)
		errorResponse(w, http.StatusBadRequest, "Invalid or missing entry ID")
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, "Invalid authentication data")
		return
	}
	log.Printf("🔍 Updating entry for user ID: %d", userID)

	var req CreateEntryRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Text == "" {
		errorResponse(w, http.StatusBadRequest, "text cannot be empty")
		return
	}

	if req.Mood < 1 || req.Mood > 10 {
		errorResponse(w, http.StatusBadRequest, "mood must be between 1 and 10")
		return
	}

	if req.Category == "" {
		errorResponse(w, http.StatusBadRequest, "category cannot be empty")
		return
	}

	// Call database to update entry
	rowsAffected, err := db.UpdateEntry(entryId, userID, req.Text, req.Mood, req.Category)
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to update entry")
		return
	}

	// If no rows affected, entry doesn't exist or doesn't belong to user
	if rowsAffected == 0 {
		errorResponse(w, http.StatusNotFound, "Entry not found or access denied")
		return
	}

	// Success response
	log.Printf("✅ Entry %d updated successfully for user %d", entryId, userID)
	respondJSON(w, http.StatusOK, CreateEntryResponse{
		Success: true,
		Message: "Entry updated successfully",
	})
}
