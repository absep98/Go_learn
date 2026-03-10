package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func Send(url string, payload interface{}) error {
	data, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	body := bytes.NewReader(data)

	resp, err := httpClient.Post(url, "application/json", body)

	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

/*
=== INTERVIEW ANSWER: WEBHOOK SENDER ===

WHAT:
This package sends outbound HTTP POST requests (webhooks) to an external URL,
delivering a JSON payload. It is the "notifier" — when something happens in our
system (entry created), we push data to whoever is listening.

A webhook is the opposite of polling:
  Polling:  "Hey, anything new?" asked every N seconds by the receiver
  Webhook:  "Something happened!" pushed immediately by the sender
Webhooks are faster, cheaper, and simpler for real-time integrations.

WHY http.Client at package level (not inside the function):
  BAD — client created inside Send() on every call:
    func Send(...) {
        client := &http.Client{Timeout: 5*time.Second}  // new client each time
        client.Post(...)
    }
  Each new http.Client starts with an empty connection pool. This means a fresh
  TCP handshake + TLS negotiation on EVERY webhook send. Expensive and slow.

  GOOD — client at package level:
    var httpClient = &http.Client{Timeout: 5*time.Second}  // created once
    func Send(...) {
        httpClient.Post(...)  // reuses existing connections
    }
  http.Client internally manages an http.Transport with a connection pool.
  Reusing the client = TCP connections to the same host are kept alive and reused.
  This is the same reason database connection pools exist.

WHY explicit status code check:
  http.Client returns an error ONLY for network-level failures:
    - DNS lookup failed
    - Connection refused
    - Timeout exceeded
  It does NOT return an error for HTTP errors like 404, 500, etc.
  A 500 from the webhook endpoint means "I received your request but failed" —
  the HTTP layer succeeded, so err == nil. Without the status check, webhook
  failures would be silently swallowed and never retried.

WHY 5 second timeout:
  No timeout = DefaultClient = waits forever.
  If webhook endpoint hangs (accepts connection, never responds), the worker
  goroutine blocks forever. With 3 workers, all 3 can get stuck on a hanging
  endpoint → entire worker pool frozen → no jobs processed.
  5 seconds is generous enough for a real webhook response but short enough
  to protect workers. Combined with retry (3 attempts × 5s max = 15s worst case
  per job before circuit breaker counts it as a failure).

WHY defer resp.Body.Close():
  Go's HTTP client streams the response body — the connection stays open until
  the body is fully read AND closed. If you skip Close(), the connection is never
  returned to the pool. After enough requests, you exhaust the connection pool
  and all future requests hang. defer ensures cleanup even if we return early
  on a status code error.

HOW it fits in the pipeline:
  entries.go → AddJob() → worker.go → WebhookBreaker.Execute() → retry.Do() → Send()
  Each layer has a role:
    Send()            — makes one HTTP attempt, returns error or nil
    retry.Do()        — retries transient failures (3 attempts, exponential backoff)
    WebhookBreaker    — stops all attempts if endpoint is sustainedly down (5 failures → open 30s)
    AddJob()          — non-blocking enqueue, HTTP handler returns immediately

TRADE-OFFS / WHAT PRODUCTION WOULD DO DIFFERENTLY:
  1. URL hardcoded → should come from config/env var (WEBHOOK_URL)
  2. No request signing — production webhooks use HMAC-SHA256 signature in a header
     (X-Webhook-Signature) so the receiver can verify the payload wasn't tampered with
  3. No dead letter queue — failed webhooks are logged and dropped. Production would
     store failed events in DB and replay them later
  4. Single URL — production would support per-subscriber URLs (webhook fan-out)
  5. No context propagation — should accept context.Context to respect upstream timeouts
*/
