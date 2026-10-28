package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"personal-analytics-backend/internal/models"
	"personal-analytics-backend/internal/worker"
	"strconv"
	"time"
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
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerWithRequestID(r)
	logger.Info("Request received", "method", "POST", "path", "/entries")

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, "Invalid authentication data")
		return
	}

	logger.Debug("Creating entry", "user_id", userID)

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
	id, err := h.Store.InsertEntry(r.Context(), int(userID), req.Text, req.Mood, req.Category)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to save entry")
		return
	}

	// reason to delete as if existing data is there its now stale as we created newly this is called the cache invalidaton.
	h.invalidateEntriesCache(r.Context(), userID)

	// Add background job to process this entry (async)
	// This returns immediately - worker processes it in background
	worker.AddJob("entry_created", map[string]interface{}{
		"event":     "entry_created",
		"entry_id":  id,
		"user_id":   userID,
		"text":      req.Text,
		"mood":      req.Mood,
		"category":  req.Category,
		"timestamp": time.Now().UTC(),
	})

	// All above are checks if passed then only allow to save it
	// Success response
	logger.Info("Entry created", "entry_id", id, "user_id", userID)
	respondJSON(w, http.StatusCreated, CreateEntryResponse{
		Success: true,
		Message: "Entry created successfully",
		ID:      id,
	})
}

// GetEntries handles GET /entries
// Returns entries for the authenticated user only
// Supports pagination: ?page=1&limit=10
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", "method", "GET", "path", "/entries")

	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination params with defaults
	// If not provided or invalid, use defaults instead of returning error
	page := 1   // Default page
	limit := 10 // Default limit

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get authenticated user_id from context (middleware puts it there)
	userIDValue := r.Context().Value(userIDKey)
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
		slog.Error("Failed to extract user_id from context", "value", userIDValue)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Invalid authentication data",
		})
		return
	}

	slog.Debug("Fetching entries", "user_id", userID, "page", page, "limit", limit)

	offset := (page - 1) * limit

	category := r.URL.Query().Get("category")

	moodMin := 0
	if moodMinStr := r.URL.Query().Get("mood_min"); moodMinStr != "" {
		if m, err := strconv.Atoi(moodMinStr); err == nil && m > 0 {
			moodMin = m
		}
	}

	// building cacheKey
	cacheKey := fmt.Sprintf("entries:user:%d:page:%d:limit:%d", userID, page, limit)

	if category == "" && moodMin == 0 {
		// Get from cache - two return values
		cached, err := h.Cache.Get(r.Context(), cacheKey)

		// if found - write raw JSON to response, return
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, cached)
			return
		}

	}
	var entries []models.Entry
	var err error

	if category != "" || moodMin > 0 {
		entries, err = h.Store.GetEntriesFiltered(r.Context(), userID, category, moodMin, limit, offset)
	} else {
		entries, err = h.Store.GetAllEntriesPaginated(r.Context(), userID, limit, offset)
	}

	if err != nil {
		slog.Error("failed to get entres", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false, "message": "Failed to fetch entries",
		})
		return
	}

	// Handle empty case - return empty array instead of null
	if entries == nil {
		entries = []models.Entry{}
	}

	total, err := h.Store.GetEntryCount(r.Context(), userID)

	if err != nil {
		slog.Error("failed to get entry count", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false, "message": "Failed to fetch entry count",
		})
		return
	}
	// Calculate total pages using integer ceiling division
	// Formula: (total + limit - 1) / limit
	// Example: 25 total, limit 10 → (25 + 10 - 1) / 10 = 34 / 10 = 3 pages
	totalPages := (total + limit - 1) / limit

	responseData := map[string]interface{}{
		"success":    true,
		"entries":    entries,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": totalPages,
	}

	if category == "" && moodMin == 0 {
		if data, err := json.Marshal(responseData); err == nil {
			if err := h.Cache.Set(r.Context(), cacheKey, string(data), 60*time.Second); err != nil {
				slog.Warn("Failed to cache entries page", "user_id", userID, "error", err)
			}
		}
	}

	// Success response with entries and pagination metadata
	slog.Info("Entries returned", "count", len(entries), "user_id", userID, "page", page, "total_pages", totalPages)
	respondJSON(w, http.StatusOK, responseData)
}

// Helper function to send JSON responses
// "Sends back a JSON response with a specific status code."
// invalidateEntriesCache drops cached reads for a user after a write.
// A failed delete is logged, not fatal: the write already succeeded and the
// stale entry expires with its 60s TTL.
// Known gap: only page 1 / limit 10 is invalidated — other cached pages stay
// stale until their TTL. See docs/ENGINEERING-NOTES.md ("Known limitations").
func (h *Handler) invalidateEntriesCache(ctx context.Context, userID int64) {
	for _, key := range []string{
		fmt.Sprintf("count:user:%d", userID),
		fmt.Sprintf("entries:user:%d:page:1:limit:10", userID),
	} {
		if err := h.Cache.Del(ctx, key); err != nil {
			slog.Warn("Cache invalidation failed", "key", key, "error", err)
		}
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	// Tell the client the response i'm sending is json
	w.Header().Set("Content-Type", "application/json")
	// set the response code 200, 400 etc
	w.WriteHeader(status)
	// Covert our go struct to json and send it to client
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) UpdateEntry(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entryId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		slog.Warn("Invalid entry ID", "error", err)
		errorResponse(w, http.StatusBadRequest, "Invalid or missing entry ID")
		return
	}

	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, "Invalid authentication data")
		return
	}
	slog.Debug("Updating entry", "entry_id", entryId, "user_id", userID)

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
	rowsAffected, err := h.Store.UpdateEntry(r.Context(), entryId, userID, req.Text, req.Mood, req.Category)
	if err != nil {
		slog.Error("Database error on update", "error", err, "entry_id", entryId)
		errorResponse(w, http.StatusInternalServerError, "Failed to update entry")
		return
	}

	// If no rows affected, entry doesn't exist or doesn't belong to user
	if rowsAffected == 0 {
		errorResponse(w, http.StatusNotFound, "Entry not found or access denied")
		return
	}

	h.invalidateEntriesCache(r.Context(), userID)

	// Success response
	slog.Info("Entry updated", "entry_id", entryId, "user_id", userID)
	respondJSON(w, http.StatusOK, CreateEntryResponse{
		Success: true,
		Message: "Entry updated successfully",
	})
}

func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entryId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		slog.Warn("Invalid entry ID", "error", err)
		errorResponse(w, http.StatusBadRequest, "Invalid or missing entry ID")
		return
	}

	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, "Invalid authenticaiton data")
		return
	}

	slog.Debug("Deleting entry", "entry_id", entryId, "user_id", userID)

	rowsAffected, err := h.Store.DeleteEntry(r.Context(), entryId, userID)

	if err != nil {
		slog.Error("Database error on delete", "error", err, "entry_id", entryId)
		errorResponse(w, http.StatusInternalServerError, "Failed to delete entry")
		return
	}

	if rowsAffected == 0 {
		errorResponse(w, http.StatusNotFound, "Entry not found or access denied")
		return
	}

	h.invalidateEntriesCache(r.Context(), userID)

	worker.AddJob("entry_deleted", map[string]interface{}{
		"event":     "entry_deleted",
		"entry_id":  entryId,
		"user_id":   userID,
		"timestamp": time.Now().UTC(),
	})

	slog.Info("Entry deleted", "entry_id", entryId, "user_id", userID)

	respondJSON(w, http.StatusOK, CreateEntryResponse{
		Success: true,
		Message: "Entry deleted successfully",
	})
}
