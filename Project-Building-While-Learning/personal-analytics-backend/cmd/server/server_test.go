package main

/*
=== SERVER CONFIGURATION TESTS ===

These pin down operational properties that were missing (verified end to end
against the Docker container — see docs/ENGINEERING-NOTES.md):

- http.Server had no timeouts, so a client that sends headers byte-by-byte
  (Slowloris) holds a connection and goroutine open indefinitely.
- Graceful shutdown listened only for os.Interrupt (Ctrl+C). `docker stop` and
  Kubernetes send SIGTERM, which bypassed shutdown entirely: no draining of
  in-flight requests or background jobs.
*/

import (
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestNewHTTPServer_BoundsSlowClients(t *testing.T) {
	srv := newHTTPServer(":0", http.NewServeMux(), 10*time.Second)

	if srv.ReadHeaderTimeout <= 0 {
		t.Errorf("ReadHeaderTimeout must be set, or slow-header clients hold connections forever")
	}
	if srv.ReadTimeout <= 0 {
		t.Errorf("ReadTimeout must be set to bound slow request bodies")
	}
	if srv.IdleTimeout <= 0 {
		t.Errorf("IdleTimeout must be set so idle keep-alive connections are reclaimed")
	}
}

func TestNewHTTPServer_WriteTimeoutLeavesRoomForTimeoutResponse(t *testing.T) {
	const requestTimeout = 10 * time.Second
	srv := newHTTPServer(":0", http.NewServeMux(), requestTimeout)

	// If WriteTimeout <= the handler deadline, the connection is cut before
	// TimeoutMiddleware can send its 503 and the client sees a reset instead.
	if srv.WriteTimeout <= requestTimeout {
		t.Errorf("WriteTimeout (%s) must exceed the request timeout (%s)", srv.WriteTimeout, requestTimeout)
	}
}

func TestShutdownSignals_IncludeSIGTERM(t *testing.T) {
	for _, s := range shutdownSignals {
		if s == syscall.SIGTERM {
			return
		}
	}
	t.Fatalf("shutdownSignals %v does not include SIGTERM; docker stop / Kubernetes would skip graceful shutdown", shutdownSignals)
}
