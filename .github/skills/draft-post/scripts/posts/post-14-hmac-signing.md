---
title: My Webhooks Were Sending Data Anyone Could Fake. HMAC Signing Fixed That.
subtitle: Part 14 of Learning Go in Public
tags: go, security, backend, buildinpublic
series: Learning Go in Public
---

In [Part 13](https://dev.to/abhishek_sharma_a9792aee8/i-needed-to-send-an-http-request-without-slowing-down-my-api-goroutines-fixed-that), I built a worker pool that fires webhooks in the background when entries are created. The webhook delivery worked — but there was a security gap I'd already documented in the trade-offs section of my own code:

```
// TRADE-OFFS / WHAT PRODUCTION WOULD DO DIFFERENTLY:
// No request signing — production webhooks use HMAC-SHA256 signature in a header
// (X-Webhook-Signature) so the receiver can verify the payload wasn't tampered with
```

Time to close that gap.

## The Problem: Unsigned Webhooks Are Unauthenticated

When my server fires a webhook, the receiver gets a JSON payload at their URL. But they have no way to know:
- Did this actually come from my server?
- Was the payload modified in transit?
- Is someone replaying an old webhook?

An attacker who knows the webhook URL can send fake events. Without signing, the receiver has to either trust everything blindly or build their own authentication on top.

GitHub, Stripe, Twilio — every production webhook system solves this with HMAC signing.

## How HMAC Signing Works

HMAC (Hash-based Message Authentication Code) uses a shared secret to produce a signature:

```
signature = HMAC-SHA256(payload_bytes, shared_secret)
```

The sender:
1. Computes `signature = HMAC-SHA256(body, secret)`
2. Sends both the body AND the signature in a header: `X-Webhook-Signature: <hex>`

The receiver:
1. Gets the body and the header
2. Computes `expected = HMAC-SHA256(body, same_secret)`
3. Compares `expected == header_value`
4. If they match → payload is genuine and untampered

An attacker can see the payload in transit but can't produce a valid signature without the secret. If anything in the payload changes (even one byte), the HMAC changes completely.

## The Implementation

### Step 1: The Sign Function

A pure function — same inputs always produce the same output:

```go
// internal/webhook/sign.go
package webhook

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func Sign(payload []byte, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    return hex.EncodeToString(h.Sum(nil))
}
```

Why hex encoding? Raw HMAC output is bytes — some are non-printable characters. You can't put `byte(3)` or `byte(255)` in an HTTP header. Hex encodes every byte as two printable characters (0-9, a-f). The result is always a clean, safe string.

Why `[]byte(secret)`? `hmac.New` takes the key as bytes. `secret` is a string. Go doesn't auto-convert — explicit cast required.

### Step 2: The Test

Before wiring it anywhere, verify the function produces the expected output for known inputs:

```go
func Test_HexCode(t *testing.T) {
    hexString := Sign([]byte("hello"), "mysecret")
    hardCodedHexString := "f09399f0c446d84b31a080e57ec483392d41e6f512f3e7ada5027abbcd358c2a"

    if hexString != hardCodedHexString {
        t.Errorf("expected %s, got %s", hardCodedHexString, hexString)
    }
}
```

The expected value was computed from an online HMAC-SHA256 calculator. If the function produces the wrong value, the test catches it immediately — before touching any HTTP code.

### Step 3: Wire It Into HTTPSender

`HTTPSender` previously had no secret. Added a `Secret` field:

```go
// Before
type HTTPSender struct{}

// After
type HTTPSender struct {
    Secret string
}
```

Then updated `Send()` to sign the payload and set the header. The key change: switched from `http.Post()` (which doesn't let you set headers) to `http.NewRequest()` + `httpClient.Do()`:

```go
func (h HTTPSender) Send(url string, payload interface{}) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    // Sign BEFORE wrapping in a reader (data is still []byte here)
    signature := Sign(data, h.Secret)

    body := bytes.NewReader(data)

    // http.NewRequest gives us a request object we can modify
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Webhook-Signature", signature)

    resp, err := httpClient.Do(req)
    // ... error handling, status check
}
```

Why sign before `bytes.NewReader`? `data` is `[]byte` at that point — perfect for `Sign`. After `bytes.NewReader(data)`, `body` is an `io.Reader` stream. You could still access `data` since it's still in scope, but signing from the raw bytes is cleaner and more explicit.

### Step 4: Thread the Secret Through Config

The webhook secret is separate from `JWT_SECRET` — they protect different things:
- `JWT_SECRET` signs auth tokens (inbound requests prove who the user is)
- `WEBHOOK_SECRET` signs outbound webhooks (receivers verify the payload came from us)

Added to `.env`:
```
WEBHOOK_SECRET=your-webhook-secret-here
```

Added to `config.go`, read into `cfg.WebhookSecret`, and passed to `HTTPSender` at startup:

```go
// main.go
worker.StartWorkerPool(
    cfg.WorkerPoolSize,
    cfg.WebhookURL,
    webhook.HTTPSender{Secret: cfg.WebhookSecret},
)
```

This is the dependency injection pattern from earlier — secrets are passed in, not read from `os.Getenv()` inside the function. Makes testing straightforward: pass a test secret, no env vars needed.

## What I Had to Learn to Write This

**`http.Post` vs `http.NewRequest`**: `http.Post` is a one-liner convenience function — great for quick calls but returns no request object. To set custom headers, you need `http.NewRequest` to get a `*http.Request` you can modify, then `httpClient.Do(req)` to send it.

**`hmac.New` argument order**: `hmac.New(hashFunction, key)` — the key is the second argument. I initially passed them backwards. The test caught it immediately.

**Why hex, not base64?** Both would work for the header value. Hex is more common for webhook signatures (GitHub uses it), slightly more readable when debugging, and produces a fixed-length string that's easy to compare.

## The Receiver Side

I didn't implement the receiver in this project — the webhook goes to webhook.site for inspection. But the receiver pattern would be:

```go
func verifyWebhook(body []byte, signature, secret string) bool {
    expected := webhook.Sign(body, secret)
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

Note `hmac.Equal` instead of `==`. This does a constant-time comparison — it takes the same amount of time regardless of where the strings differ. A regular `==` comparison short-circuits on the first mismatch, which leaks timing information that an attacker could use to guess the signature byte by byte. For security comparisons, always use constant-time equality.

## What I Learned

**Test pure functions first, before wiring them anywhere.** `Sign` is a pure function — no HTTP, no state, no side effects. Writing a unit test for it before touching `Send()` meant I knew the core logic was correct before dealing with HTTP plumbing.

**The "strict mode" learning method works.** For this feature, I wrote all the code myself with zero reference to existing files — only Go's standard library docs when I forgot a function signature. First attempt had several mistakes. Second attempt had two. Third attempt was clean. Real retrieval practice is painful but it sticks.

**Security gaps you document will eventually come back.** I wrote "No request signing" in my trade-offs section knowing I'd close it later. Documenting trade-offs isn't just for interviews — it's a backlog for future improvements.

---

Up next: shipping this to a real VPS.

*This is Part 14 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7) | [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp) | [Part 12](https://dev.to/abhishek_sharma_a9792aee8/my-backend-crashed-every-time-redis-went-down-three-patterns-fixed-that-43j2) | [Part 13](https://dev.to/abhishek_sharma_a9792aee8/i-needed-to-send-an-http-request-without-slowing-down-my-api-goroutines-fixed-that)*
