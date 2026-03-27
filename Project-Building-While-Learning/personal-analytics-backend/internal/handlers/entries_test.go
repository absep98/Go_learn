package handlers

/*
=== ENTRY HANDLER TESTS ===

WHAT WE ARE TESTING:
End-to-end handler tests for CreateEntry — from HTTP request to DB write to JSON response.
These tests call the real handler with a real in-memory SQLite DB (no mocks for DB layer).

KEY CONCEPTS:

1. TestMain — PACKAGE-LEVEL SETUP
   Runs once before all tests. Initializes the in-memory DB and a stub Redis client.
   Both are required because CreateEntry internally calls db.InsertEntry AND cache.Delete.

2. WHY REDIS CLIENT (stub, not real Redis)
   CreateEntry calls cache.Delete after every insert (cache invalidation).
   cache.Delete uses redis.Client. If Client is nil → panic.
   We use goredis.NewClient() directly (no ping/retry) to make Client non-nil instantly.
   The circuit breaker handles the actual "connection refused" gracefully without panicking.
   goredis = import alias for "github.com/redis/go-redis/v9" (avoids name conflict with our internal/redis package)

3. INJECTING user_id INTO CONTEXT
   In production: AuthMiddleware reads the JWT and puts user_id into context.
   In tests: we skip AuthMiddleware and inject user_id manually:
     ctx := context.WithValue(req.Context(), "user_id", int64(1))
     req = req.WithContext(ctx)

4. CHECKING BOTH STATUS CODE AND RESPONSE BODY
   rec.Code         → HTTP status (201, 400, etc.)
   rec.Body         → JSON written by the handler
   json.NewDecoder(rec.Body).Decode(&resp) → parse JSON into struct for assertions

TEST SCENARIOS:
- TestCreateEntry          → valid input → 201, success:true, id > 0
- TestCreateEntry_MissingText  → empty text → 400, success:false
- TestCreateEntry_InvalidMood  → mood=400 (out of range) → 400, success:false
*/

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/redis"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestMain(m *testing.M) {

	err := db.InitDB(":memory:")
	if err != nil {
		panic("failed to init test DB: " + err.Error())
	}

	// Skip ping/retry — just make Client non-nil so cache.Delete doesn't panic
	redis.Client = goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})

	os.Exit(m.Run())
}

func TestCreateEntry(t *testing.T) {

	body := strings.NewReader(`{"text":"feeling good, feeling alright","mood":8,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPost, "/entries", body)

	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	CreateEntry(rec, req)

	var resp CreateEntryResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if resp.ID <= 0 {
		t.Errorf("expected ID > 0, got %d", resp.ID)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, get %d", rec.Code)
	}
}

func TestCreateEntry_MissingText(t *testing.T) {

	body := strings.NewReader(`{"text":"","mood":8,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPost, "/entries", body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))

	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	CreateEntry(rec, req)

	var resp CreateEntryResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Success == true {
		t.Errorf("expected success false but got true")
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, get %d", rec.Code)
	}
}

func TestCreateEntry_InvalidMood(t *testing.T) {

	body := strings.NewReader(`{"text":"feeling good","mood":400,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPost, "/entries", body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))

	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	CreateEntry(rec, req)

	var resp CreateEntryResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Success == true {
		t.Errorf("expected success false but got true")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, get %d", rec.Code)
	}
}
