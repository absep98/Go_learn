package handlers

import (
	"net/http"
	"time"
)

// LoggingMiddleware logs each request with method, path, and duration
// This is now structured logging - much easier to parse and analyze!
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the actual handler
		next(w, r)

		// Calculate how long the request took
		duration := time.Since(start)

		// Get logger with request_id (if available)
		logger := GetLoggerWithRequestID(r)

		// Log structured data (key-value pairs)
		// This outputs JSON like:
		// {"request_id":"abc123","time":"...","level":"INFO","msg":"Request","method":"GET","path":"/health","duration_ms":5}
		logger.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", duration.Milliseconds(),
		)
	}
}
