package handlers

/*
=== REQUEST TIMEOUT MIDDLEWARE ===

Problem: If database or Redis hangs, request blocks FOREVER.
The goroutine handling that request never returns = goroutine leak.
After many hung requests: thousands of leaked goroutines = server crash.

Solution: Wrap every request with context.WithTimeout.
If handler doesn't respond within the deadline, return 503 Service Unavailable.
(The notes below describe the first, goroutine+select version and why it was
replaced by http.TimeoutHandler — see "UPDATE" at the bottom of this file.)

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
	"log/slog"
	"net/http"
	"time"
)

// RequestTimeout is the maximum time a request can take
// Set from config in main.go (default: 10 seconds)
var RequestTimeout = 10 * time.Second

// TimeoutMiddleware wraps handlers with a deadline.
// If the handler takes longer than RequestTimeout, the client gets 503.
//
// FIX (was a data race): the first version ran next(w, r) in a goroutine and,
// on timeout, wrote a 504 to the SAME ResponseWriter the still-running handler
// was writing to. `go test -race` flagged it, and the client could receive the
// handler's late 200 + body instead of the timeout. A handler panic also
// crashed the whole process, because it happened off the serving goroutine
// where net/http's recover() can't see it.
//
// http.TimeoutHandler solves all three: the handler writes into a buffer
// guarded by a mutex, the buffer is copied to the real writer only if the
// handler finishes in time, writes after the deadline get ErrHandlerTimeout,
// and a handler panic is re-raised on the serving goroutine.
// It still cancels r.Context() at the deadline, so ctx-aware DB/Redis calls abort.
func TimeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		http.TimeoutHandler(next, RequestTimeout, "Request timed out").ServeHTTP(w, r)

		if time.Since(start) >= RequestTimeout {
			slog.Warn("Request timeout",
				"method", r.Method,
				"path", r.URL.Path,
				"timeout", RequestTimeout.String(),
			)
		}
	}
}

/*
=== INTERVIEW ANSWER: REQUEST TIMEOUT MIDDLEWARE ===

WHAT:
TimeoutMiddleware wraps every handler with a deadline. If the handler doesn't
respond within the limit (10s), the client gets a 504 Gateway Timeout immediately
instead of waiting forever.

WHY:
If Redis or DB hangs, the goroutine handling that request blocks forever —
a goroutine leak. With thousands of concurrent requests all hanging, the server
runs out of memory and crashes. Timeout middleware caps the damage.

HOW:
1. context.WithTimeout(r.Context(), 10s) — creates a "ticking bomb" context.
   After 10s, ctx.Done() channel closes, signalling all listeners to stop.
2. r.WithContext(ctx) — replaces the request's context so downstream DB/Redis
   calls using r.Context() automatically get the timeout.
3. Handler runs in a GOROUTINE — critical. Without this, next(w,r) would block
   the current goroutine and the select could never fire concurrently.
4. select races done channel vs ctx.Done():
   - done closes first → handler finished normally, return
   - ctx.Done() fires first → 10s exceeded, write 504 and return
5. defer cancel() — always called, releases context resources (prevents leak).

WHY GOROUTINE IS REQUIRED:
Two things must run concurrently: the handler doing work, and the middleware
watching the clock. A goroutine is the only way to do both at the same time.
Without it, the select would never be reached until the handler already finished.

WHY close(done) NOT done <- true:
close() unblocks ALL readers at once. A send unblocks only one.
struct{} carries zero bytes — it's a pure signal with no data.

GOROUTINE LEAK CAVEAT:
After timeout fires, the handler goroutine keeps running — you cannot force-kill
goroutines in Go. The CLIENT gets the fast 504, but the goroutine lives until it
naturally finishes. To truly cancel work: pass ctx to DB/Redis calls — they
abort automatically when the context is cancelled.

TRADE-OFFS:
- Response may already be partially written when timeout fires — production would
  use a ResponseWriter wrapper to detect this and avoid writing headers twice.
- 10s is configurable via REQUEST_TIMEOUT env var.

UPDATE — WHAT WAS WRONG WITH THE DESIGN ABOVE (found with go test -race):
The "partially written" trade-off was really a data race. Two goroutines wrote
to one ResponseWriter with no synchronization. Now replaced by http.TimeoutHandler,
which is exactly that "ResponseWriter wrapper": it buffers the handler's output,
and whichever side (handler or deadline) takes the mutex first wins.
Costs of the stdlib version:
- Status is 503 Service Unavailable, not 504. 504 is what a gateway (nginx)
  returns when the UPSTREAM is slow; 503 says the app itself shed the request.
- The response is buffered in memory, and the wrapper does not implement
  http.Flusher/Hijacker — so no streaming/SSE/WebSocket routes behind it.
  This API has none.
Regression tests: timeout_test.go.
*/
