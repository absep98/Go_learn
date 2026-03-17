package handlers

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
		captureID = r.Context().Value("request_id").(string)
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
