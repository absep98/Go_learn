package main

import (
	"net/http"
	"os"
	"syscall"
	"time"
)

// shutdownSignals trigger graceful shutdown.
// os.Interrupt = Ctrl+C locally; SIGTERM = what `docker stop` and Kubernetes send.
// With only os.Interrupt, SIGTERM fell through to Go's default handler and the
// process exited immediately — no request draining, no worker drain.
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// newHTTPServer builds the server with connection-level timeouts.
//
//   - ReadHeaderTimeout: a client must finish sending headers within 5s
//     (defends against Slowloris-style connection exhaustion).
//   - ReadTimeout: bounds slow request bodies.
//   - WriteTimeout: requestTimeout + 5s, so TimeoutMiddleware's 503 can still be
//     written after the handler deadline fires.
//   - IdleTimeout: reclaims idle keep-alive connections.
func newHTTPServer(addr string, handler http.Handler, requestTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      requestTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
