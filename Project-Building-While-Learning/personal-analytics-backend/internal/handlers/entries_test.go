package handlers

/*
=== ENTRY HANDLER TESTS ===

WHAT WE ARE TESTING:
End-to-end handler tests for CreateEntry — from HTTP request to DB write to JSON response.
These tests call the real handler with a real Postgres DB via env vars

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
	"fmt"
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

	// Build DSN from env vars (same as main.go does)
	host := os.Getenv("DB_HOST")
	if host == "" {
		fmt.Println("DB_HOST not set, skipping DB tests")
		os.Exit(0)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	err := db.InitDB(dsn)
	if err != nil {
		panic("failed to init test DB: " + err.Error())
	}

	// Clean up tables before running tests — ensures fresh state every run
	_, err = db.DB.Exec("TRUNCATE TABLE entries, users RESTART IDENTITY CASCADE")
	if err != nil {
		panic("failed to clean test DB: " + err.Error())
	}

	// Skip ping/retry — just make Client non-nil so cache.Delete doesn't panic
	redis.Client = goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})

	os.Exit(m.Run())
}

// helper — inserts one entry for userID=1, returns its ID
func createTestEntry(t *testing.T) int64 {
	t.Helper()
	id, err := db.InsertEntry(context.Background(), 1, "test text", 5, "mood")
	if err != nil {
		t.Fatalf("failed to create test entry: %v", err)
	}
	return id
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

func TestGetEntries(t *testing.T) {

	// Setup: Insert 2 entries
	// Every call inserts a new row with a different auto inc ID. There's no uniqueness constraint on text or mood only email in the user table in unique
	createTestEntry(t)
	createTestEntry(t)

	// Action: Call GetEntries
	req := httptest.NewRequest(http.MethodGet, "/entries", nil)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	GetEntries(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}

	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

func TestGetEntries_Empty(t *testing.T) {

	createTestEntry(t)

	req := httptest.NewRequest(http.MethodGet, "/entries", nil)
	ctx := context.WithValue(req.Context(), userIDKey, int64(10))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	GetEntries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}

	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}

}

func TestUpdateEntry(t *testing.T) {

	id := createTestEntry(t)

	body := strings.NewReader(`{"text":"updated text","mood":7,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/entries?id=%d", id), body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	UpdateEntry(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}

	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

func TestUpdateEntry_NotFound(t *testing.T) {
	id := createTestEntry(t)

	body := strings.NewReader(`{"text":"updated text","mood":7,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/entries?id=%d", id), body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(10))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	UpdateEntry(rec, req)

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var resp map[string]interface{}

	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["success"] != false {
		t.Errorf("expected success false")
	}
}

func TestUpdateEntry_MissingText(t *testing.T) {

	id := createTestEntry(t)

	body := strings.NewReader(`{"text":"","mood":10,"category":"mood"}`)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/entries?id=%d", id), body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	UpdateEntry(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]interface{}

	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["success"] != false {
		t.Errorf("expected success false")
	}
}

func TestDeleteEntry(t *testing.T) {

	id := createTestEntry(t)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/entries?id=%d", id), nil)

	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	DeleteEntry(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestDeleteEntry_NotFound(t *testing.T) {

	req := httptest.NewRequest(http.MethodDelete, "/entries?id=9999", nil)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))

	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	DeleteEntry(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
