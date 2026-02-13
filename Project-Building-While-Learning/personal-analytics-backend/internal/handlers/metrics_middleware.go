package handlers

import (
	"fmt"
	"net/http"
	"personal-analytics-backend/internal/metrics"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode // Save it
	// DEBUG: Uncomment to see wrapper intercepting
	fmt.Printf("🔍 Wrapper intercepted status: %d\n", statusCode)
	rw.ResponseWriter.WriteHeader(statusCode) // Pass it through
}

var AppMetrics = metrics.NewMetrics()

func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		AppMetrics.RequestStarted(path)
		timeStarted := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default if WriteHeader not called
		}
		next(wrapped, r)

		// Calculate duration
		duration := time.Since(timeStarted).Milliseconds()

		// Record completion
		AppMetrics.RequestCompleted(path, float64(duration), wrapped.statusCode)
	})
}
