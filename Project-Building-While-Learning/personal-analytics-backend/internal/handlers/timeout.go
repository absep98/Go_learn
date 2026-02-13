package handlers

/*
=== REQUEST TIMEOUT MIDDLEWARE ===

Problem: If database or Redis hangs, request blocks FOREVER.
The goroutine handling that request never returns = goroutine leak.
After many hung requests: thousands of leaked goroutines = server crash.

Solution: Wrap every request with context.WithTimeout.
If handler doesn't respond within the deadline, return 504 Gateway Timeout.

=== HOW CONTEXT TIMEOUT WORKS ===

context.WithTimeout creates a "ticking bomb" context:
  ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)

After 10 seconds, ctx.Done() channel closes → signals all listeners to stop.

Any operation using this context (DB queries, Redis calls) will abort:
  db.QueryContext(ctx, "SELECT ...")  ← Will return error after 10s
  redis.Client.Get(ctx, key)          ← Will return error after 10s

=== THE SELECT PATTERN ===

  select {
  case <-done:        // Handler finished normally
  case <-ctx.Done():  // Timeout expired!
  }

select waits for WHICHEVER happens first:
- Handler finishes → great, send response
- Timeout fires → return 504 error

=== WHY USE A CHANNEL FOR DONE? ===

The handler runs in a goroutine (go func() { ... }).
We need to know when it finishes. Channel is the Go way to
signal "I'm done" between goroutines.

=== IMPORTANT: GOROUTINE LEAK PREVENTION ===

Even after timeout, the handler goroutine keeps running until it naturally
finishes (we can't forcefully kill goroutines in Go). But the CLIENT gets
a fast 504 response instead of waiting forever.

To truly cancel work inside the handler, pass ctx to DB/Redis calls:
  db.QueryContext(ctx, ...)  ← Aborts when context is cancelled
*/

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// RequestTimeout is the maximum time a request can take
// Set from config in main.go (default: 10 seconds)
var RequestTimeout = 10 * time.Second

// TimeoutMiddleware wraps handlers with a deadline
// If handler takes longer than RequestTimeout, returns 504
func TimeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a timeout context from the request's existing context
		// This preserves any values already in context (request_id, user_id)
		ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
		defer cancel() // Always cancel to release resources (prevents context leak)

		// Replace the request's context with our timeout context
		// Now any downstream code using r.Context() gets the timeout
		r = r.WithContext(ctx)

		// Channel to signal when handler completes
		done := make(chan struct{})

		// Run the handler in a goroutine
		go func() {
			next(w, r)
			close(done) // Signal completion
		}()

		// Wait for handler to finish OR timeout to expire
		select {
		case <-done:
			// Handler completed normally — response already written
			return

		case <-ctx.Done():
			// Timeout! Handler took too long
			slog.Warn("Request timeout",
				"method", r.Method,
				"path", r.URL.Path,
				"timeout", RequestTimeout.String(),
			)
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			return
		}
	}
}
