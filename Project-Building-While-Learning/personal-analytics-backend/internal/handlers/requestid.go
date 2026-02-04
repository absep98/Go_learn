package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"log/slog"
)

// RequestMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate a request ID using GenerateRequestID()
		requestID := GenerateRequestID()

		// Store it in context using context.WithValue()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next(w, r.WithContext(ctx))
	}
}

// Generate request ID creates a unique request ID
func GenerateRequestID() string {

	// Create a byte slice of length 8
	byteSlice := make([]byte, 8)

	// Fill it with random bytes using rand.Read()
	_, err := rand.Read(byteSlice)

	if err != nil {
		// If we can't generate random bytes, something is seriously wrong
		// Better to crash than to continue with predictable IDs
		panic("failed to generate request ID: " + err.Error())
	}
	// Convert bytes to hex string using hex.EncodeToString()
	hexstring := hex.EncodeToString(byteSlice)

	return hexstring
}


// GetLoggerWithRequestID returns a logger that includes the request_id
func GetLoggerWithRequestID(r *http.Request) *slog.Logger {
	// Get the value from context (might be nil if not set)
	value := r.Context().Value("request_id")  // ← Capital C!

	// Type assert to string
	requestID, ok := value.(string)
	if !ok {  // ← !ok means "if not ok" (if assertion failed)
		// if Type assertion failed, use empty string
		requestID = ""  // ← Set the variable, don't return
	}
	// return a logger that always includes request_id
	// slog.Default().With() creates a new logger with this field in every log.
	return slog.Default().With("request_id", requestID)
}
