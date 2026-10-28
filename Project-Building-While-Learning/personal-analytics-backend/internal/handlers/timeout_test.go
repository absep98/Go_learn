package handlers

/*
=== TIMEOUT MIDDLEWARE TESTS ===

These pin down three properties the original hand-rolled middleware violated:

1. A handler that keeps running after the deadline must not be able to write
   to the client's response. (Original: the handler goroutine and the
   middleware both wrote to the same http.ResponseWriter concurrently — a data
   race that `go test -race` reports, and the "late" body leaked into the
   response.)

2. A panic inside a handler must surface on the serving goroutine, where
   net/http's per-connection recover() can catch it. (Original: the handler ran
   in its own goroutine, so one panicking request crashed the whole process.)

3. A fast handler's status, headers and body pass through unchanged.

Run with: go test -race ./internal/handlers/ -run TimeoutMiddleware
*/

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func withRequestTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	old := RequestTimeout
	RequestTimeout = d
	t.Cleanup(func() { RequestTimeout = old })
}

func TestTimeoutMiddleware_LateHandlerWriteDoesNotReachClient(t *testing.T) {
	withRequestTimeout(t, 20*time.Millisecond)

	handlerDone := make(chan struct{})
	slow := func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		<-r.Context().Done() // simulate work (e.g. a slow DB call) that outlives the deadline
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("late body")) //nolint:errcheck
	}

	rec := httptest.NewRecorder()
	TimeoutMiddleware(slow)(rec, httptest.NewRequest(http.MethodGet, "/entries", nil))
	<-handlerDone // wait until the late write has definitely happened

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 on timeout, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "late body") {
		t.Errorf("late handler output leaked into the timed-out response: %q", rec.Body.String())
	}
}

func TestTimeoutMiddleware_HandlerPanicSurfacesOnServingGoroutine(t *testing.T) {
	withRequestTimeout(t, time.Second)

	boom := func(w http.ResponseWriter, r *http.Request) { panic("boom") }

	defer func() {
		if recover() == nil {
			t.Errorf("expected the handler panic to propagate to the caller so net/http can recover it")
		}
	}()
	TimeoutMiddleware(boom)(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestTimeoutMiddleware_FastHandlerPassesThrough(t *testing.T) {
	withRequestTimeout(t, time.Second)

	fast := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Entry-ID", "42")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success":true}`)) //nolint:errcheck
	}

	rec := httptest.NewRecorder()
	TimeoutMiddleware(fast)(rec, httptest.NewRequest(http.MethodPost, "/entries", nil))

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Entry-ID"); got != "42" {
		t.Errorf("expected header X-Entry-ID=42, got %q", got)
	}
	if rec.Body.String() != `{"success":true}` {
		t.Errorf("unexpected body %q", rec.Body.String())
	}
}
