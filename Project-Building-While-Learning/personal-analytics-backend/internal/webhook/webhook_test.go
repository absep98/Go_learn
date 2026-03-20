package webhook

/*
=== WEBHOOK PACKAGE TESTS ===

WHAT WE ARE TESTING:
HTTPSender.Send() makes a real HTTP POST to a URL.
NoOpSender.Send() does nothing and returns nil.
We test: success (200), failure (500), and NoOp behavior.

KEY CONCEPTS USED:

1. httptest.NewServer — FAKE EXTERNAL SERVER
   server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       w.WriteHeader(200)
   }))
   defer server.Close()

   Unlike httptest.NewRecorder (which tests a handler function),
   NewServer starts a REAL local HTTP server on a random port.
   We use this because HTTPSender.Send() makes an OUTGOING HTTP request.
   The fake server receives that request and returns whatever status we set.

   Why not use a real URL like https://google.com?
   - Tests must work offline, on any machine, in CI
   - Real servers can go down, change behavior, or rate-limit
   - We need full control over the response (200 vs 500)

2. server.URL — the fake server's address
   sender.Send(server.URL, payload)
   This is the local address like http://127.0.0.1:PORT.
   HTTPSender sends its POST here instead of a real webhook endpoint.

3. TESTING NOOP SENDER
   NoOpSender doesn't make any HTTP call at all — it just returns nil.
   No server needed. Just create it and call Send() directly.
   This tests the "disabled webhook" code path.

Q: Why is httptest.NewServer used here but httptest.NewRecorder in handlers tests?
A: NewRecorder = testing a handler FUNCTION (handler is the code under test, writes response)
   NewServer = testing code that CALLS a server (HTTPSender is the caller, server receives)

Q: Why test the 500 case?
A: HTTPSender should return an error when the server responds with non-2xx.
   Without this test, a silent bug could cause failed webhooks to appear successful.
*/

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

func TestNoOpSender(t *testing.T) {

	sender := NoOpSender{}

	err := sender.Send("http://example.com", nil)

	if err != nil { // fail if there IS an error
		t.Errorf("expected nil, got %v", err)
	}
}
