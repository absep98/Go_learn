package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSender_Send(t *testing.T) {

	// create a real http server on a random port
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200) // fake server always returns 200
	}))

	defer server.Close()

	sender := HTTPSender{}
	err := sender.Send(server.URL, map[string]string{"test": "data"})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHTTPSender_Send_NonOKStatus(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))

	defer server.Close()

	sender := HTTPSender{}
	err := sender.Send(server.URL, map[string]string{"test": "data"})

	if err == nil {
		t.Errorf("expected error for 500 status, got nil")
	}
}
