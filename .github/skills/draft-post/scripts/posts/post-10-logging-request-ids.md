---
title: Every Request Looked the Same in My Logs. Then I Added One Line.
subtitle: Part 10 of Learning Go in Public
tags: go, logging, backend, buildinpublic
series: Learning Go in Public
---

In [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93), I Dockerized my backend and added database migrations. One command (`docker compose up`) now spins up Go + Postgres + Redis.

Then I hit a bug. I sent a POST to `/entries` and got a 500 back. No idea why.

I opened the logs:

```
2026/02/01 10:15:23 POST /entries - Request received
2026/02/01 10:15:23 Creating entry for user ID: 3
2026/02/01 10:15:23 POST /entries - Request received
2026/02/01 10:15:23 Creating entry for user ID: 7
2026/02/01 10:15:23 Failed to extract user_id from context
2026/02/01 10:15:23 Entry created successfully with ID: 42
```

Five lines. Three users. Which line belongs to which request? I had no idea. The logs were flat strings — timestamps and messages, nothing else. When two requests hit at the same millisecond, the lines interleaved and became useless.

I couldn't debug my own backend.

## Step 1: Replace Every log.Printf in the Codebase

Go ships a structured logging package in the standard library: `log/slog`. No external dependencies. I created a logger package:

```go
package logger

import (
    "log/slog"
    "os"
)

var Log *slog.Logger

func InitLogger() {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    Log = slog.New(handler)
    slog.SetDefault(Log)
}
```

Then I went through every file and replaced `log.Printf` with `slog.Info`, `slog.Error`, `slog.Debug`.

**Before:**
```go
log.Printf("Entry created successfully with ID: %d", id)
```

**After:**
```go
slog.Info("Entry created", "entry_id", id, "user_id", userID)
```

The difference isn't cosmetic. Here's what the output looks like:

**Before** (flat string):
```
2026/02/01 10:15:23 Entry created successfully with ID: 42
```

**After** (structured JSON):
```json
{"time":"2026-02-01T10:15:23Z","level":"INFO","msg":"Entry created","entry_id":42,"user_id":3}
```

Every field is a key-value pair. You can search for `user_id:3` to find all logs for user 3. You can filter by `level:ERROR` to find only failures. You can pipe everything into Datadog, ELK, or just `jq` on the command line:

```bash
cat logs.json | jq 'select(.user_id == 3)'
```

Try doing that with `log.Printf("User %d created entry %d", userID, entryID)`. You'd need regex. For every query. Forever.

## The Logging Middleware Upgrade

My logging middleware was the worst offender. Before:

```go
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        current_time := time.Now().Format("2006-01-02 15:04:05")
        log.Printf("%s %s %s", current_time, r.Method, r.URL.Path)
        next(w, r)
    }
}
```

No duration. No way to know how long a request took. After:

```go
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next(w, r)
        duration := time.Since(start)

        slog.Info("Request",
            "method", r.Method,
            "path", r.URL.Path,
            "duration_ms", duration.Milliseconds(),
        )
    }
}
```

Now every request automatically logs how long it took. If a request takes 2000ms instead of 50ms, I see it instantly. Before, I would never have known.

## Step 2: The Problem That Structured Logging Didn't Solve

The logs were clean now. Structured. Searchable. But go back to the original problem:

```json
{"time":"...","level":"INFO","msg":"Request received","method":"POST","path":"/entries"}
{"time":"...","level":"DEBUG","msg":"Creating entry","user_id":3}
{"time":"...","level":"INFO","msg":"Request received","method":"POST","path":"/entries"}
{"time":"...","level":"DEBUG","msg":"Creating entry","user_id":7}
{"time":"...","level":"ERROR","msg":"Failed to extract user_id from context"}
{"time":"...","level":"INFO","msg":"Entry created","entry_id":42,"user_id":3}
```

Which request failed? User 3's or user 7's? The error line has no user ID — the extraction *failed*, so there's nothing to log. And even if it did, how do you correlate "Request received" with "Entry created" when 10 requests are running concurrently?

You need a thread that ties every log line from the same request together. That thread is a **request ID**.

## Request IDs: One String That Changes Everything

The idea is simple: generate a unique ID when a request arrives, attach it to every log line for that request, and return it in the response header.

```go
func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        requestID := GenerateRequestID()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next(w, r.WithContext(ctx))
    }
}

func GenerateRequestID() string {
    b := make([]byte, 8)
    _, err := rand.Read(b)
    if err != nil {
        panic("failed to generate request ID: " + err.Error())
    }
    return hex.EncodeToString(b)
}
```

`crypto/rand` generates 8 random bytes. `hex.EncodeToString` turns them into a 16-character hex string like `a3f8b2c1e9d04567`. Unique enough for request tracing. Stored in the request context so every handler downstream can access it.

## The Glue: Logger With Request ID

The key function that ties it all together:

```go
func GetLoggerWithRequestID(r *http.Request) *slog.Logger {
    value := r.Context().Value("request_id")
    requestID, ok := value.(string)
    if !ok {
        requestID = ""
    }
    return slog.Default().With("request_id", requestID)
}
```

`slog.Default().With("request_id", requestID)` creates a new logger where *every* log line automatically includes the request ID. You don't pass it manually. You call `GetLoggerWithRequestID(r)` once and every `.Info()`, `.Error()`, `.Debug()` on that logger includes it.

The logging middleware now uses this:

```go
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next(w, r)
        duration := time.Since(start)

        logger := GetLoggerWithRequestID(r)
        logger.Info("Request",
            "method", r.Method,
            "path", r.URL.Path,
            "duration_ms", duration.Milliseconds(),
        )
    }
}
```

## The Same Logs, Now Debuggable

```json
{"time":"...","level":"INFO","msg":"Request received","method":"POST","path":"/entries","request_id":"a3f8b2c1"}
{"time":"...","level":"DEBUG","msg":"Creating entry","user_id":3,"request_id":"a3f8b2c1"}
{"time":"...","level":"INFO","msg":"Request received","method":"POST","path":"/entries","request_id":"d7e9f0a2"}
{"time":"...","level":"DEBUG","msg":"Creating entry","user_id":7,"request_id":"d7e9f0a2"}
{"time":"...","level":"ERROR","msg":"Failed to extract user_id","request_id":"c4b5e6f8"}
{"time":"...","level":"INFO","msg":"Entry created","entry_id":42,"user_id":3,"request_id":"a3f8b2c1"}
```

Now I can see:
- `a3f8b2c1` = User 3's request. Succeeded.
- `d7e9f0a2` = User 7's request. Log lines stop after "Creating entry" — something failed silently.
- `c4b5e6f8` = A third request entirely. The one that couldn't extract a user ID.

One `jq` command pulls out every line for a single request:

```bash
cat logs.json | jq 'select(.request_id == "c4b5e6f8")'
```

The bug that took 30 minutes to find with flat logs now takes 10 seconds.

## The Middleware Chain

Request ID middleware wraps everything else — it's the outermost layer:

```go
http.HandleFunc("/entries", handlers.RequestIDMiddleware(
    handlers.RateLimitMiddleware(handlers.LoggingMiddleware(
        handlers.AuthMiddleware(entryHandler)))))
```

The order matters: RequestID runs first, so the ID exists in context before the logging middleware tries to read it. If you put logging outside request ID, the first log line would have no ID.

## What I Learned

**`log.Printf` is a dead end.** The moment you have more than one concurrent request, flat strings become useless. `slog` is in the standard library. There's no reason to use anything else for most Go backends.

**Structured logging is about searchability, not readability.** JSON logs look ugly in a terminal. They look beautiful in a log aggregator. You're not reading logs for fun — you're filtering them during an outage at 2 AM.

**Request IDs cost almost nothing and save hours.** 16 bytes of randomness per request. Stored in context. Attached to every log line. The debugging ROI is absurd.

**`slog.With()` is the secret weapon.** Instead of manually passing the request ID to every log call, `slog.Default().With("request_id", id)` creates a child logger that includes it automatically. One call per request, not one per log line.

---

Up next: I can see individual requests now, but I still can't see the *big picture*. How many requests per minute? What's the average response time? Which endpoints are slowest? Time to build a metrics endpoint.

*This is Part 10 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93)*