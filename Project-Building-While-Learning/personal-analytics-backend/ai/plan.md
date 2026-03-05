# Implementation Plan — Webhook Notification System

> **Feature:** When a user creates, updates, or deletes an entry, fire an HTTP webhook to a configured URL.
> **Why this feature:** It connects THREE systems you already built (worker pool + retry + circuit breaker) with a new skill (outbound HTTP calls). This is the highest-value next feature because it turns your simulated `time.Sleep(2s)` in `processJob()` into real production work.

---

## Table of Contents

1. [What You'll Learn (The Real Goal)](#1-what-youll-learn-the-real-goal)
2. [Architecture Overview](#2-architecture-overview)
3. [New Package: `internal/webhook`](#3-new-package-internalwebhook)
4. [Files to Modify](#4-files-to-modify)
5. [Files to Create](#5-files-to-create)
6. [Interfaces — Why and Where](#6-interfaces--why-and-where)
7. [Structs — Complete Definitions](#7-structs--complete-definitions)
8. [Error Handling Strategy](#8-error-handling-strategy)
9. [Concurrency Model](#9-concurrency-model)
10. [Implementation Order (Step by Step)](#10-implementation-order-step-by-step)
11. [Testing Plan](#11-testing-plan)
12. [Configuration](#12-configuration)
13. [Go Concepts You'll Practice](#13-go-concepts-youll-practice)

---

## 1. What You'll Learn (The Real Goal)

This feature is disguised education. Here's what each piece teaches:

| Implementation Task | Go Concept Learned |
|--------------------|--------------------|
| Create `http.Client` with timeout | `http.Client` struct, `context` cancellation |
| POST JSON to external URL | `http.NewRequestWithContext`, `json.Marshal` |
| Retry failed webhooks | Your existing `retry.Do()` — now used for HTTP, not just Redis |
| Circuit breaker on webhook endpoint | Your existing `circuitbreaker.Execute()` — now protects HTTP, not just Redis |
| Define `Sender` interface | **Interfaces in Go** — the single most important concept for testability |
| Pass `context.Context` through worker | Context propagation beyond HTTP handlers |
| Write `_test.go` with mock sender | **Table-driven tests**, **mock via interface**, `httptest.NewServer` |
| Graceful worker shutdown | `sync.WaitGroup`, channel closing, draining |

---

## 2. Architecture Overview

### Current Flow (what happens now)

```
CreateEntry handler
    → worker.AddJob("entry_created", payload)
        → worker goroutine picks it up
            → processJob() does time.Sleep(2s)  ← FAKE WORK
```

### New Flow (what we're building)

```
CreateEntry handler
    → worker.AddJob("entry_created", payload)
        → worker goroutine picks it up
            → processJob() calls webhook.Send()  ← REAL WORK
                → webhook.Send() does:
                    1. Build JSON payload
                    2. Circuit breaker check (is endpoint healthy?)
                    3. HTTP POST with timeout context
                    4. If fails: retry with exponential backoff
                    5. Log result (success or final failure)
```

### Package Dependency Diagram

```
handlers/entries.go
    → worker.AddJob(...)               (existing, no change)

worker/worker.go
    → webhook.DefaultSender.Send(...)  (NEW: real work instead of Sleep)

webhook/webhook.go  (NEW PACKAGE)
    ├── circuitbreaker.Execute(...)    (EXISTING: reuse your circuit breaker)
    ├── retry.Do(...)                  (EXISTING: reuse your retry logic)
    └── http.Client.Do(req)            (NEW: outbound HTTP call)
```

**Key principle:** The new `webhook` package REUSES `circuitbreaker` and `retry`. You don't copy code, you import and compose. This is Go's composition model.

---

## 3. New Package: `internal/webhook`

### Why a Separate Package?

| Option | Problem |
|--------|---------|
| Put webhook logic in `worker/` | Worker becomes responsible for HTTP calls — violates single responsibility |
| Put it in `handlers/` | Handlers is for HTTP **inbound**, not outbound |
| New `webhook/` package | Clean boundary: "everything about sending webhooks lives here" |

### Package Structure

```
internal/webhook/
├── webhook.go       ← Sender interface + HTTPSender implementation
└── webhook_test.go  ← Tests with mock HTTP server
```

Only TWO files. Go packages should be small and focused.

---

## 4. Files to Modify

### 4.1 `internal/worker/worker.go`

**What changes:** Replace `time.Sleep(2s)` with real webhook calls.

**Before:**
```go
func processJob(job Job) {
    switch job.Type {
    case "entry_created":
        slog.Debug("Processing entry creation", "payload", job.Payload)
        time.Sleep(2 * time.Second) // Simulate slow task
    }
}
```

**After:**
```go
func processJob(job Job) {
    switch job.Type {
    case "entry_created":
        slog.Debug("Processing entry creation", "payload", job.Payload)
        err := webhookSender.Send(context.Background(), "entry.created", job.Payload)
        if err != nil {
            slog.Error("Webhook delivery failed",
                "job_type", job.Type,
                "error", err,
            )
        }

    case "entry_updated":
        slog.Debug("Processing entry update", "payload", job.Payload)
        err := webhookSender.Send(context.Background(), "entry.updated", job.Payload)
        if err != nil {
            slog.Error("Webhook delivery failed",
                "job_type", job.Type,
                "error", err,
            )
        }

    case "entry_deleted":
        slog.Debug("Processing entry deletion", "payload", job.Payload)
        err := webhookSender.Send(context.Background(), "entry.deleted", job.Payload)
        if err != nil {
            slog.Error("Webhook delivery failed",
                "job_type", job.Type,
                "error", err,
            )
        }
    }
}
```

**New code to add** (package-level variable for the sender):
```go
// webhookSender is the webhook delivery mechanism.
// Set via SetWebhookSender() from main.go during startup.
// Using an interface variable means tests can swap in a mock.
var webhookSender webhook.Sender

// SetWebhookSender configures the webhook sender for the worker pool.
// Called once from main.go after loading config.
func SetWebhookSender(s webhook.Sender) {
    webhookSender = s
}
```

**Why `SetWebhookSender()` instead of a global?**
- The worker package shouldn't import `config` (that's `main.go`'s job).
- `main.go` creates the sender with config values and injects it.
- Tests can inject a mock sender.
- This is **dependency injection without frameworks** — idiomatic Go.

### 4.2 `internal/config/config.go`

**What changes:** Add webhook configuration fields.

**New fields in `Config` struct:**
```go
type Config struct {
    // ... existing fields ...

    // Webhook
    WebhookURL     string        // Target URL for webhook notifications
    WebhookTimeout time.Duration // HTTP timeout for webhook calls
    WebhookSecret  string        // HMAC secret for signing payloads
}
```

**New loading logic in `Load()`:**
```go
// Load WebhookURL (optional - empty means webhooks disabled)
cfg.WebhookURL = os.Getenv("WEBHOOK_URL")

// Load WebhookTimeout (default: 5 seconds)
webhookTimeout, err := strconv.Atoi(os.Getenv("WEBHOOK_TIMEOUT"))
if err != nil || webhookTimeout == 0 {
    webhookTimeout = 5
}
cfg.WebhookTimeout = time.Duration(webhookTimeout) * time.Second

// Load WebhookSecret (optional - for HMAC signing)
cfg.WebhookSecret = os.Getenv("WEBHOOK_SECRET")
```

**Why optional?** If `WEBHOOK_URL` is empty, webhooks are disabled. The app still works without it. This is the **feature flag** pattern — no code change needed to toggle.

### 4.3 `cmd/server/main.go`

**What changes:** Create and inject the webhook sender.

**New code after worker pool start:**
```go
// Configure webhook sender (if URL is set)
if cfg.WebhookURL != "" {
    sender := webhook.NewHTTPSender(cfg.WebhookURL, cfg.WebhookTimeout, cfg.WebhookSecret)
    worker.SetWebhookSender(sender)
    slog.Info("Webhook configured", "url", cfg.WebhookURL, "timeout", cfg.WebhookTimeout)
} else {
    worker.SetWebhookSender(webhook.NoOpSender{})
    slog.Info("Webhook not configured (WEBHOOK_URL not set)")
}
```

### 4.4 `internal/handlers/entries.go`

**What changes:** Add webhook jobs for update and delete (currently only create fires a job).

**In `UpdateEntry()`** (after `cache.Delete(...)` line):
```go
worker.AddJob("entry_updated", map[string]interface{}{
    "entry_id": entryId,
    "user_id":  userID,
})
```

**In `DeleteEntry()`** (after `cache.Delete(...)` line):
```go
worker.AddJob("entry_deleted", map[string]interface{}{
    "entry_id": entryId,
    "user_id":  userID,
})
```

---

## 5. Files to Create

### 5.1 `internal/webhook/webhook.go`

This is the core new file. Full implementation below:

```go
package webhook

import (
    "bytes"
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "personal-analytics-backend/internal/circuitbreaker"
    "personal-analytics-backend/internal/retry"
    "time"
)

// =============================================================================
// INTERFACE — The most important part of this file
// =============================================================================

// Sender defines how webhooks are delivered.
//
// WHY AN INTERFACE?
// In production: HTTPSender sends real HTTP POST requests.
// In tests:      MockSender records calls without network.
// If disabled:   NoOpSender does nothing.
//
// The worker package depends on this INTERFACE, not on HTTPSender directly.
// This means the worker doesn't know or care HOW delivery happens.
// This is the Dependency Inversion Principle — depend on abstractions,
// not on concrete implementations.
//
// Go interface rule: Interfaces are SMALL. One method is ideal.
// The Go standard library's io.Reader has 1 method. io.Writer has 1 method.
// Our Sender has 1 method. This makes it trivial to implement.
type Sender interface {
    Send(ctx context.Context, event string, payload interface{}) error
}

// =============================================================================
// HTTP SENDER — The production implementation
// =============================================================================

// HTTPSender delivers webhooks via HTTP POST.
// It wraps each call with circuit breaker + retry for resilience.
type HTTPSender struct {
    url     string             // Where to POST (e.g., "https://example.com/webhook")
    client  *http.Client       // HTTP client with timeout
    secret  string             // HMAC-SHA256 signing secret (empty = no signing)
    breaker *circuitbreaker.CircuitBreaker // Protects against dead endpoints
}

// NewHTTPSender creates a sender configured for production use.
//
// Parameters:
//   - url: the webhook endpoint (e.g., from WEBHOOK_URL env var)
//   - timeout: how long to wait for a response (e.g., 5s)
//   - secret: HMAC signing key (empty string = skip signing)
//
// The circuit breaker trips after 3 consecutive failures and
// waits 60 seconds before testing recovery. These are tuned for
// webhook endpoints which may have temporary outages.
func NewHTTPSender(url string, timeout time.Duration, secret string) *HTTPSender {
    return &HTTPSender{
        url: url,
        client: &http.Client{
            Timeout: timeout,
        },
        secret:  secret,
        breaker: circuitbreaker.NewCircuitBreaker(3, 60*time.Second),
    }
}

// WebhookPayload is the JSON body sent to the webhook URL.
// The receiver sees a consistent, predictable format.
type WebhookPayload struct {
    Event     string      `json:"event"`      // "entry.created", "entry.updated", "entry.deleted"
    Timestamp string      `json:"timestamp"`  // ISO 8601 format
    Data      interface{} `json:"data"`       // The job payload (entry_id, user_id, etc.)
}

// Send delivers a webhook notification.
//
// Flow:
//   1. Marshal payload to JSON
//   2. Circuit breaker: is the endpoint healthy?
//      - Yes → continue
//      - No (open) → return error immediately (no HTTP call)
//   3. Inside circuit breaker: retry up to 3 times with backoff
//      - Each attempt: HTTP POST with timeout context
//      - Success (2xx) → return nil
//      - Failure → backoff, try again
//   4. All retries exhausted → circuit breaker counts this as a failure
//
// WHY RETRY INSIDE CIRCUIT BREAKER?
// Circuit breaker answers: "Is this service alive at all?"
// Retry answers: "That one call failed, maybe try again?"
//
// If the endpoint is DOWN (circuit open) → skip retry entirely (fast fail).
// If the endpoint is UP but one request glitched → retry it.
// This is the correct composition order.
func (s *HTTPSender) Send(ctx context.Context, event string, payload interface{}) error {
    // Step 1: Build the JSON payload
    body := WebhookPayload{
        Event:     event,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Data:      payload,
    }

    jsonBytes, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("webhook: failed to marshal payload: %w", err)
    }

    // Step 2: Circuit breaker wraps the entire retry loop
    return s.breaker.Execute(func() error {
        // Step 3: Retry with exponential backoff (3 attempts, starting at 500ms)
        return retry.Do(3, 500*time.Millisecond, func() error {
            return s.doPost(ctx, jsonBytes)
        })
    })
}

// doPost performs a single HTTP POST request.
// Separated from Send() so retry.Do() can call it repeatedly.
//
// WHY http.NewRequestWithContext?
// - Creates a request that is cancellable via context
// - If the parent context (from timeout middleware or shutdown) is cancelled,
//   this request is aborted immediately — no orphaned HTTP calls
// - Always prefer NewRequestWithContext over NewRequest in production Go
func (s *HTTPSender) doPost(ctx context.Context, jsonBytes []byte) error {
    // Create request with context (cancellable!)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(jsonBytes))
    if err != nil {
        return fmt.Errorf("webhook: failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "PersonalAnalytics-Webhook/1.0")

    // HMAC signature (if secret is configured)
    // This lets the receiver verify the payload came from us
    // and wasn't tampered with. GitHub webhooks use this exact pattern.
    if s.secret != "" {
        signature := computeHMAC(jsonBytes, s.secret)
        req.Header.Set("X-Webhook-Signature", signature)
    }

    // Execute HTTP call
    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("webhook: HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    // Check response status
    // 2xx = success, anything else = failure
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("webhook: endpoint returned status %d", resp.StatusCode)
    }

    slog.Debug("Webhook delivered", "url", s.url, "status", resp.StatusCode)
    return nil
}

// computeHMAC creates an HMAC-SHA256 signature of the payload.
//
// HMAC = Hash-based Message Authentication Code
// It takes a message + a secret key → produces a unique signature.
// The receiver computes the same HMAC with the same secret.
// If signatures match → message is authentic and unmodified.
//
// This is HOW GitHub, Stripe, and Slack verify webhook payloads.
func computeHMAC(message []byte, secret string) string {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(message)
    return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// =============================================================================
// NO-OP SENDER — Used when webhooks are disabled
// =============================================================================

// NoOpSender is a Sender that does nothing.
// Used when WEBHOOK_URL is not configured.
//
// WHY NOT JUST USE nil?
// If webhookSender is nil, every call site needs: if sender != nil { sender.Send() }
// With NoOpSender, the call site is always: sender.Send() — no nil checks.
// This is the Null Object Pattern — a real object that safely does nothing.
type NoOpSender struct{}

// Send on NoOpSender does nothing and always succeeds.
func (n NoOpSender) Send(ctx context.Context, event string, payload interface{}) error {
    return nil
}
```

### 5.2 `internal/webhook/webhook_test.go`

```go
package webhook

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

// =============================================================================
// TEST WITH MOCK HTTP SERVER
// =============================================================================

// httptest.NewServer creates a REAL HTTP server on localhost with a random port.
// It's not a "mock" — it actually listens on TCP. But you control the responses.
// This is Go's standard library approach to testing HTTP clients.

func TestHTTPSender_Success(t *testing.T) {
    // Create a test server that returns 200 OK
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify the request is what we expect
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        if r.Header.Get("Content-Type") != "application/json" {
            t.Errorf("expected application/json content type")
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close() // Shut down test server when test ends

    // Create sender pointing to test server
    sender := NewHTTPSender(server.URL, 5*time.Second, "")

    // Send a webhook
    err := sender.Send(context.Background(), "entry.created", map[string]interface{}{
        "entry_id": 1,
        "user_id":  42,
    })

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}

func TestHTTPSender_ServerError_ReturnsError(t *testing.T) {
    // Server that always returns 500
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()

    sender := NewHTTPSender(server.URL, 5*time.Second, "")

    err := sender.Send(context.Background(), "entry.created", map[string]interface{}{
        "entry_id": 1,
    })

    // Should fail after retries
    if err == nil {
        t.Fatal("expected error for 500 response, got nil")
    }
}

func TestHTTPSender_HMAC_Signature(t *testing.T) {
    secret := "my-webhook-secret"
    var receivedSignature string

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        receivedSignature = r.Header.Get("X-Webhook-Signature")
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    sender := NewHTTPSender(server.URL, 5*time.Second, secret)
    sender.Send(context.Background(), "entry.created", map[string]interface{}{"id": 1})

    if receivedSignature == "" {
        t.Fatal("expected X-Webhook-Signature header, got empty")
    }
    if len(receivedSignature) < 10 {
        t.Fatalf("signature looks too short: %s", receivedSignature)
    }
}

func TestHTTPSender_ContextCancelled(t *testing.T) {
    // Slow server that takes 10 seconds
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(10 * time.Second)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    sender := NewHTTPSender(server.URL, 1*time.Second, "")

    // Cancel context after creation
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    err := sender.Send(ctx, "entry.created", map[string]interface{}{"id": 1})

    if err == nil {
        t.Fatal("expected error when context is cancelled, got nil")
    }
}

func TestNoOpSender_AlwaysSucceeds(t *testing.T) {
    sender := NoOpSender{}
    err := sender.Send(context.Background(), "anything", nil)
    if err != nil {
        t.Fatalf("NoOpSender should never fail, got: %v", err)
    }
}

// =============================================================================
// TABLE-DRIVEN TESTS — The Go way to test multiple scenarios
// =============================================================================

// Instead of writing 5 separate test functions for different status codes,
// define a table of inputs and expected outputs, loop through them.
// This is the most common test pattern in Go codebases.

func TestHTTPSender_StatusCodes(t *testing.T) {
    tests := []struct {
        name       string // Describes the test case
        statusCode int    // What the server returns
        wantErr    bool   // Do we expect an error?
    }{
        {name: "200 OK", statusCode: 200, wantErr: false},
        {name: "201 Created", statusCode: 201, wantErr: false},
        {name: "204 No Content", statusCode: 204, wantErr: false},
        {name: "400 Bad Request", statusCode: 400, wantErr: true},
        {name: "404 Not Found", statusCode: 404, wantErr: true},
        {name: "500 Internal Server Error", statusCode: 500, wantErr: true},
    }

    for _, tt := range tests {
        // t.Run creates a sub-test — each row runs independently
        // If "400 Bad Request" fails, you see that specific name in output
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(tt.statusCode)
            }))
            defer server.Close()

            sender := NewHTTPSender(server.URL, 5*time.Second, "")
            err := sender.Send(context.Background(), "test.event", nil)

            if tt.wantErr && err == nil {
                t.Errorf("expected error for status %d, got nil", tt.statusCode)
            }
            if !tt.wantErr && err != nil {
                t.Errorf("expected no error for status %d, got: %v", tt.statusCode, err)
            }
        })
    }
}
```

### 5.3 `test-endpoints/test-webhook.ps1`

```powershell
# Test Webhook Integration
# Prerequisites:
#   1. Start a webhook receiver: npx webhook-test (or use webhook.site)
#   2. Set WEBHOOK_URL in .env to the receiver URL
#   3. Restart the server

$baseUrl = "http://localhost:8080"

# Step 1: Login (get token)
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method POST -Body '{"email":"test@test.com","password":"password123"}' -ContentType "application/json"
$token = $loginResponse.token
Write-Host "Token: $($token.Substring(0, 20))..."

# Step 2: Create entry (should trigger webhook)
Write-Host "`nCreating entry (should fire webhook)..."
$headers = @{ Authorization = "Bearer $token" }
$body = '{"text":"Webhook test entry","mood":8,"category":"testing"}'
$response = Invoke-RestMethod -Uri "$baseUrl/entries" -Method POST -Headers $headers -Body $body -ContentType "application/json"
Write-Host "Create response: $($response | ConvertTo-Json -Compress)"

# Step 3: Update entry (should trigger webhook)
$entryId = $response.id
Write-Host "`nUpdating entry $entryId (should fire webhook)..."
$updateBody = '{"text":"Updated via webhook test","mood":9,"category":"testing"}'
$updateResponse = Invoke-RestMethod -Uri "$baseUrl/entries?id=$entryId" -Method PATCH -Headers $headers -Body $updateBody -ContentType "application/json"
Write-Host "Update response: $($updateResponse | ConvertTo-Json -Compress)"

# Step 4: Delete entry (should trigger webhook)
Write-Host "`nDeleting entry $entryId (should fire webhook)..."
$deleteResponse = Invoke-RestMethod -Uri "$baseUrl/entries?id=$entryId" -Method DELETE -Headers $headers
Write-Host "Delete response: $($deleteResponse | ConvertTo-Json -Compress)"

Write-Host "`nDone! Check your webhook receiver for 3 events:"
Write-Host "  1. entry.created"
Write-Host "  2. entry.updated"
Write-Host "  3. entry.deleted"
```

---

## 6. Interfaces — Why and Where

### The `Sender` Interface

```go
type Sender interface {
    Send(ctx context.Context, event string, payload interface{}) error
}
```

**Why this is the most important line in the entire feature:**

```
                      Sender (interface)
                     /        |         \
              HTTPSender   NoOpSender   MockSender (in tests)
              (real HTTP)  (disabled)   (records calls)
```

The `worker` package depends on `Sender`, NOT on `HTTPSender`. This means:

| Situation | What's injected | Effect |
|-----------|----------------|--------|
| Production (`WEBHOOK_URL` set) | `HTTPSender` | Real HTTP POST calls |
| Production (`WEBHOOK_URL` empty) | `NoOpSender` | Nothing happens (no `if` checks needed) |
| Unit tests | `MockSender` | Records calls for assertions, no network |

**Go interface rule of thumb:**
> Define the interface WHERE IT'S USED, not where it's implemented.

The `Sender` interface lives in the `webhook` package because that's where the implementations are. But if we later needed the worker package to define its own dependency, it would define its own 1-method interface:

```go
// In worker package (if we wanted even looser coupling)
type notifier interface {
    Send(ctx context.Context, event string, payload interface{}) error
}
```

Go interfaces are **implicit** — `HTTPSender` satisfies `Sender` without declaring `implements`. The compiler checks it automatically. This is fundamentally different from Java/C# and is one of Go's most powerful features.

### Why Not Use an Interface for Everything?

You might ask: "Should I also make `db` an interface? And `cache`?"

**Answer:** Not yet. Add an interface **when you have a second implementation** (usually a mock for tests). Premature interfaces add complexity without benefit.

When to add:
- `db` → when you write `db_test.go` and need to mock SQL calls
- `cache` → when you want a "test mode" that skips Redis entirely

This is called **"Accept interfaces, return structs"** — a core Go proverb.

---

## 7. Structs — Complete Definitions

### `WebhookPayload` (what gets sent)

```go
type WebhookPayload struct {
    Event     string      `json:"event"`      // "entry.created"
    Timestamp string      `json:"timestamp"`  // "2026-02-24T10:30:00Z"
    Data      interface{} `json:"data"`       // {entry_id: 5, user_id: 3}
}
```

**Example JSON sent to webhook URL:**
```json
{
    "event": "entry.created",
    "timestamp": "2026-02-24T10:30:00Z",
    "data": {
        "entry_id": 42,
        "user_id": 3
    }
}
```

### `HTTPSender` (production implementation)

```go
type HTTPSender struct {
    url     string
    client  *http.Client
    secret  string
    breaker *circuitbreaker.CircuitBreaker
}
```

**Why each field:**

| Field | Type | Purpose |
|-------|------|---------|
| `url` | `string` | Where to POST. Set once from config. |
| `client` | `*http.Client` | Reusable HTTP client with timeout. **Never use `http.DefaultClient` in production** — it has no timeout. |
| `secret` | `string` | HMAC key for payload signing. Empty = skip signing. |
| `breaker` | `*CircuitBreaker` | Separate breaker from Redis's. Each external dependency gets its own breaker. |

### `NoOpSender` (null object)

```go
type NoOpSender struct{}
```

Empty struct = zero bytes of memory. This is how Go represents "a thing that exists but holds no data." Used in channels (`chan struct{}`), and here for the null object pattern.

---

## 8. Error Handling Strategy

### Layer-by-Layer Error Behavior

```
Layer 1: doPost()
    HTTP call fails → return wrapped error: "webhook: HTTP request failed: <original>"
    Non-2xx status  → return specific error: "webhook: endpoint returned status 500"

Layer 2: retry.Do() (wraps doPost)
    Attempt 1 fails → wait 500ms → retry
    Attempt 2 fails → wait 1s → retry
    Attempt 3 fails → return: "operation failed after 3 attempts: webhook: ..."

Layer 3: circuitbreaker.Execute() (wraps retry)
    If closed → run retry loop → count success/failure
    If open   → return immediately: "circuit breaker is open"

Layer 4: processJob() (in worker)
    Error → log it with slog.Error (worker doesn't return errors to anyone)
    No error → log success

Layer 5: The user
    Never sees webhook errors. Webhooks are FIRE AND FORGET.
    The HTTP response was already sent before the worker even starts.
```

### Why Fire and Forget?

```
                  Timeline
                     │
Handler responds ────┤ ← User gets 201 Created HERE
                     │
Worker picks job ────┤ ← 0-100ms later
                     │
Webhook sent ────────┤ ← 500ms-5s later
                     │
Retry if failed ─────┤ ← Up to 3.5s more
```

The user already has their response. Webhook delivery is a **background concern**. If it fails, we log it. We don't retry forever — 3 attempts is enough.

### Error Types and Wrapping

Every error in the chain uses `%w` (wrapping):

```go
fmt.Errorf("webhook: HTTP request failed: %w", err)
fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
```

This creates an error chain. In logs, you can see the full history:
```
circuit breaker failure → retry exhausted → HTTP timeout → context deadline exceeded
```

---

## 9. Concurrency Model

### What Runs Where

```
Main goroutine
    → starts HTTP server (in its own goroutine)
    → starts 3 worker goroutines
    → blocks on signal.Notify

HTTP request goroutine (per request)
    → handler runs → AddJob() → returns to client

Worker goroutine (1 of 3)
    → picks job from channel
    → calls webhook.Send()
        → retry.Do() runs synchronously WITHIN the worker goroutine
        → http.Client.Do() blocks until response or timeout
    → job done, wait for next
```

### No New Goroutines Needed

The worker already runs in a goroutine. The HTTP call blocks that goroutine (which is fine — that's what workers are for). Retry sleeps block that goroutine too. **Three worker goroutines handle three concurrent webhook deliveries.** If you need more throughput, increase `WORKERPOOL_SIZE`.

### Graceful Shutdown Improvement

**Current problem:** When server shuts down, workers are killed mid-job.

**Fix:** Add `sync.WaitGroup` to track in-flight jobs.

```go
var wg sync.WaitGroup

func worker(id int) {
    for job := range JobQueue {
        wg.Add(1)
        processJob(job)
        wg.Done()
    }
}

// Called from main.go during shutdown
func DrainWorkers() {
    close(JobQueue) // Stop accepting new jobs; worker range loops exit
    wg.Wait()       // Wait for in-flight jobs to finish
}
```

**In `main.go` shutdown sequence:**
```go
// STEP 8: Drain workers before closing connections
worker.DrainWorkers()
slog.Info("Workers drained")
```

**Why this matters:** Without draining, a webhook HTTP call in progress could be interrupted when Redis/DB close. With draining, the server waits for ongoing deliveries to complete.

---

## 10. Implementation Order (Step by Step)

Do these in order. Each step compiles and works before the next.

### Step 1: Config (5 min)
Add `WebhookURL`, `WebhookTimeout`, `WebhookSecret` to `config.go`.
Add `.env` entries. **Test:** Server starts without errors.

### Step 2: Webhook package — types only (10 min)
Create `internal/webhook/webhook.go` with:
- `Sender` interface
- `WebhookPayload` struct
- `NoOpSender` struct + `Send()` method
- `HTTPSender` struct + `NewHTTPSender()` constructor
- Leave `Send()` as a stub that just logs

**Test:** `go build ./...` compiles.

### Step 3: Wire it up (10 min)
- Add `SetWebhookSender()` to `worker.go`
- Add sender creation to `main.go`
- Change `processJob()` to call `webhookSender.Send()`

**Test:** Start server, create entry, see log: "Webhook delivered" (stub).

### Step 4: Real HTTP implementation (20 min)
Implement `HTTPSender.Send()` and `doPost()`:
- JSON marshaling
- `http.NewRequestWithContext`
- Status code check
- No retry/breaker yet — just a raw HTTP call

**Test:** Point `WEBHOOK_URL` at `https://webhook.site/<your-id>`. Create an entry. See the webhook arrive.

### Step 5: Add circuit breaker + retry (10 min)
Wrap `doPost()` with `retry.Do()`, wrap that with `breaker.Execute()`.
These are your EXISTING packages — just import and compose.

**Test:** Point `WEBHOOK_URL` at a dead URL. See retry logs. See circuit breaker open after 3 failures.

### Step 6: HMAC signing (10 min)
Implement `computeHMAC()`. Add `X-Webhook-Signature` header.

**Test:** See signature in webhook.site request headers.

### Step 7: Update/Delete webhooks (5 min)
Add `worker.AddJob()` calls to `UpdateEntry()` and `DeleteEntry()`.

**Test:** Update and delete entries, see webhook events.

### Step 8: Graceful worker shutdown (10 min)
Add `sync.WaitGroup` and `DrainWorkers()`.

**Test:** Start server, create entry (triggers webhook), immediately Ctrl+C. Should see "Workers drained" before "Server stopped."

### Step 9: Write tests (20 min)
Create `webhook_test.go` with:
- `TestHTTPSender_Success`
- `TestHTTPSender_ServerError_ReturnsError`
- `TestHTTPSender_HMAC_Signature`
- `TestHTTPSender_ContextCancelled`
- `TestNoOpSender_AlwaysSucceeds`
- `TestHTTPSender_StatusCodes` (table-driven)

**Test:** `go test ./internal/webhook/ -v`

### Step 10: PowerShell integration test (5 min)
Create `test-endpoints/test-webhook.ps1`.

**Total estimated time: ~1.5–2 hours**

---

## 11. Testing Plan

### Unit Tests (Go's `testing` package)

| Test | What It Validates | Go Concept |
|------|-------------------|------------|
| `TestHTTPSender_Success` | 200 response → no error | `httptest.NewServer`, basic assertion |
| `TestHTTPSender_ServerError` | 500 → error after retries | Error checking, retry behavior |
| `TestHTTPSender_HMAC` | Signature header present & non-empty | Header inspection |
| `TestHTTPSender_ContextCancelled` | Cancelled context → fast error | `context.WithTimeout`, cancellation propagation |
| `TestNoOpSender` | Always returns nil | Null object pattern verification |
| `TestHTTPSender_StatusCodes` | 2xx=ok, 4xx/5xx=error | **Table-driven tests** (the Go testing pattern) |

### How to Run

```bash
# Run webhook tests only
go test ./internal/webhook/ -v

# Run with race detector (finds concurrency bugs)
go test ./internal/webhook/ -race -v

# Run ALL tests in the project
go test ./... -v
```

### Integration Tests (PowerShell)

`test-endpoints/test-webhook.ps1` — creates/updates/deletes entries and verifies webhook delivery by checking the receiver.

### Manual Testing

Use [webhook.site](https://webhook.site) — it gives you a unique URL that shows all incoming requests. Set `WEBHOOK_URL` to it, create entries, and watch webhooks arrive in real-time in the browser.

---

## 12. Configuration

### New `.env` entries

```bash
# Webhook Configuration
WEBHOOK_URL=https://webhook.site/your-unique-id   # Empty = webhooks disabled
WEBHOOK_TIMEOUT=5                                   # Seconds (default: 5)
WEBHOOK_SECRET=your-hmac-secret-key                 # Empty = no signing
```

### Default Behavior

| Env Var | Default | Effect when empty |
|---------|---------|-------------------|
| `WEBHOOK_URL` | `""` | Webhooks disabled, `NoOpSender` used |
| `WEBHOOK_TIMEOUT` | `5` seconds | — |
| `WEBHOOK_SECRET` | `""` | No HMAC signature on payloads |

---

## 13. Go Concepts You'll Practice

Ordered by learning value:

### Tier S: Career-Defining Concepts
1. **Interfaces** — Defining `Sender`, implementing it three ways, injecting via function
2. **`context.Context` propagation** — From handler → worker → HTTP client, with timeout
3. **Table-driven tests** — The pattern used in 90% of Go test files
4. **`httptest.NewServer`** — Testing HTTP clients without real networks

### Tier A: Production Must-Knows
5. **`http.Client` with timeout** — Why `http.DefaultClient` is dangerous
6. **`http.NewRequestWithContext`** — Context-aware outbound requests
7. **Error wrapping chain** — `%w` through multiple layers
8. **`sync.WaitGroup`** — Waiting for goroutines to finish (shutdown)

### Tier B: Security & Design
9. **HMAC-SHA256** — How webhooks are authenticated in the real world
10. **Null Object Pattern** — `NoOpSender` eliminates nil checks
11. **Composition over inheritance** — Circuit breaker wraps retry wraps HTTP
12. **Dependency injection without frameworks** — `SetWebhookSender(s Sender)`

---

## Summary: Files Changed

| File | Action | Lines Changed (approx) |
|------|--------|----------------------|
| `internal/webhook/webhook.go` | **CREATE** | ~130 |
| `internal/webhook/webhook_test.go` | **CREATE** | ~120 |
| `internal/worker/worker.go` | MODIFY | ~30 |
| `internal/config/config.go` | MODIFY | ~15 |
| `cmd/server/main.go` | MODIFY | ~10 |
| `internal/handlers/entries.go` | MODIFY | ~8 |
| `test-endpoints/test-webhook.ps1` | **CREATE** | ~30 |
| `.env` | MODIFY | ~3 |

**Total new code: ~220 lines**
**Total modified code: ~65 lines**
**New packages: 1 (`internal/webhook/`)**
**New dependencies: 0** (everything uses stdlib + your existing packages)

---

*Read this plan slowly. For each section, ask yourself: "Why this choice? What would break if I did it differently?" That's where the real learning is.*
