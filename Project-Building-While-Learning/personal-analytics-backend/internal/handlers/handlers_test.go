package handlers

/*
=== HANDLERS PACKAGE TESTS ===

WHAT WE ARE TESTING:
HTTP handlers and middleware — without starting a real server.

KEY CONCEPTS USED:

1. httptest.NewRequest — FAKE INCOMING REQUEST
   req := httptest.NewRequest("GET", "/ping", nil)
   Creates a fake *http.Request. Arguments: method, URL path, body (nil = no body).
   Passed to the handler as the `r` parameter.

2. httptest.NewRecorder — FAKE RESPONSE WRITER
   w := httptest.NewRecorder()
   In production, the handler writes to a real network connection.
   NewRecorder captures everything the handler writes (status, headers, body) in memory.
   Think of it as a notepad that records the response.

3. CALLING HANDLER DIRECTLY — NO SERVER NEEDED
   PingHandler(w, req)
   The handler runs exactly as in production — just writes to the recorder instead of network.

4. w.Result() — READ WHAT WAS RECORDED
   result := w.Result()         → converts recorder to *http.Response
   result.StatusCode            → what status code the handler set
   io.ReadAll(result.Body)      → read the response body as []byte
   string(body)                 → convert bytes to readable string

5. TESTING ENV VARIABLES
   os.Setenv("KEY", "value")          → set before calling handler
   defer os.Unsetenv("KEY")           → clean up AFTER test finishes
   defer runs when the function returns — prevents env leaking into other tests.

6. TESTING MIDDLEWARE
   Middleware wraps a handler: func(next http.HandlerFunc) http.HandlerFunc
   To test it: create an inner handler that captures what you need, wrap it, call wrapped.
   For RequestIDMiddleware: inner handler reads request_id from context.
   For LoggingMiddleware: just verify the inner handler was called (status 200).

Q: Why httptest.NewServer for webhook but httptest.NewRecorder for handlers?
A: NewRecorder tests a handler FUNCTION directly (no server).
   NewServer starts a real local server — needed when the CODE UNDER TEST
   makes an outgoing HTTP request (like HTTPSender.Send does).

Q: Why use defer for os.Unsetenv?
A: If the test fails mid-way and panics, defer still runs — guaranteed cleanup.
   Without defer, a failed test could leave the env var set and break other tests.
*/

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPingHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)

	w := httptest.NewRecorder()
	PingHandler(w, req)
	result := w.Result()
	if result.StatusCode != 200 {
		t.Errorf("Request isn't successfully")
	}

	body, _ := io.ReadAll(result.Body)
	response := string(body)
	if !strings.Contains(response, "pong") {
		t.Errorf("Request isn't successfully")
	}
}

func TestPingHandler_CustomMessage(t *testing.T) {

	os.Setenv("PING_MESSAGE", "Hola")
	defer os.Unsetenv("PING_MESSAGE") // clean up after test

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	PingHandler(w, req)

	result := w.Result()
	body, _ := io.ReadAll(result.Body)
	response := string(body)

	if !strings.Contains(response, "Hola") {
		t.Errorf("Request isn't successfull")
	}
}

func TestRequestIDMiddleware(t *testing.T) {

	var captureID string

	// inner handler - captures whatever is in context
	innerHandler := func(w http.ResponseWriter, r *http.Request) {
		captureID = r.Context().Value(requestIDKey).(string)
	}

	// wrap it with the middleware
	wrapped := RequestIDMiddleware(innerHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	wrapped(w, req)

	if captureID == "" {
		t.Errorf("expected request_id in context, got empty string")
	}
}

func TestLogginMiddleware(t *testing.T) {

	innerHanlder := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}

	wrapped := LoggingMiddleware(innerHanlder)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	wrapped(w, req)

	result := w.Result()
	if result.StatusCode != 200 {
		t.Errorf("Test logging failed")
	}

}

func TestGenerateRequestID(t *testing.T) {

	id := GenerateRequestID()

	length := len(id)
	if length == 0 {
		t.Errorf("request id isn't generated successfully")
	}
	if length != 16 {
		t.Errorf("request id isn't generated successfully")
	}

}
