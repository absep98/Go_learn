package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
=== AUTH MIDDLEWARE TEST FLOW ===

This test uses a "Mocking" strategy to verify the Security Guard (Middleware).

 1. THE TARGET (dummyHandler):
    A simple "Mannequin" that always says "200 OK".
    If the request reaches this, it means the Guard let them in.

 2. THE GUARD (AuthMiddleware):
    We wrap the Guard around the Mannequin.

3. THE ATTACK (httptest):

  - NewRequest: A "Fake Letter" sent to the server.

  - NewRecorder: A "Secret Tape Recorder" that catches the server's reply.

    4. THE VERDICT:
    We check the Tape Recorder.

  - If we sent NO TOKEN, we expect "401 Unauthorized".

  - If we sent a VALID TOKEN, we expect "200 OK" (from the dummy).

DIAGRAM:
Request -> [ AuthMiddleware ] -> (if valid) -> [ dummyHandler ] -> "200 OK"

	[      ^       ]
	[ (if invalid) ] -> "401 Unauthorized"
*/
func TestAuthMiddleware_NoHeader(t *testing.T) {
	// Step 1: Create the "Goal Line" (The Dummy)
	// This only runs if the Middleware calls 'next(w, r)'
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Step 2: Wrap the Dummy with our "Security Guard"
	h := &Handler{JWTSecret: "test-secret"}
	handler := h.AuthMiddleware(dummyHandler)

	// Step 3: Create a Fake Request (The Bullet) and a Recorder (The Target Paper)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Step 4: Fire the Request at the Wrapped Handler
	handler(rec, req)

	// Step 5: Check the "Target Paper" (The Recorder)
	// Since we didn't set an "Authorization" header, the Guard should block it.
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Security Breach! Expected 401, but the Guard let them through with: %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {

	h := &Handler{JWTSecret: "test-secret"}

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := h.AuthMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer this.is.garbage")

	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {

	h := &Handler{JWTSecret: "test-secret"}

	// creating claims data to put inside token
	claims := jwt.MapClaims{
		"user_id": int64(42),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("test-secret"))

	if err != nil {
		t.Fatalf("Error generating JWT token %v", err)
	}

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := h.AuthMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Failed at authorization")
	}
}
