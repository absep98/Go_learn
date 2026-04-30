---
title: My Backend Crashed Every Time Redis Went Down. Three Patterns Fixed That.
subtitle: Part 12 of Learning Go in Public
tags: go, resilience, backend, buildinpublic
series: Learning Go in Public
---

In [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp), I built a metrics system to see the big picture — request counts, latencies, error rates per endpoint.

Then I looked at my `main.go` and realized something embarrassing. The JWT secret was hardcoded. The Redis address was a string literal. The rate limit was a magic number buried in middleware. Every config value was scattered across different files with no central source of truth.

And then Redis went down during a test, and my entire backend froze. Every request hung for 5 seconds waiting for a timeout. The server was technically running but completely useless.

Two problems. One week.

## Problem 1: Config Values Everywhere

My `main.go` looked like this:

```go
err = db.InitDB("./data.db")
err = redis.InitRedis("localhost:6379")

// Somewhere in a handler file:
secret := os.Getenv("JWT_SECRET")

// Somewhere in middleware:
maxRequests := 100
window := 60 * time.Second
```

Four files. Four different patterns for reading config. Some used `os.Getenv`, some were hardcoded, some had defaults and some didn't. If I wanted to change the rate limit, I had to know it lived inside `ratelimit.go`. If I forgot to set `JWT_SECRET`, the server would boot and silently produce invalid tokens.

### The Fix: One Struct, One Function

```go
type Config struct {
    Port             string
    DBPath           string
    RedisAddr        string
    JWTSecret        string
    LogLevel         string
    ShutdownTimeout  time.Duration
    RateLimitRequests int
    RateLimitWindow  time.Duration
    WorkerPoolSize   int
}
```

Every setting the app needs, in one place. Then a single `Load()` function that reads environment variables with sensible defaults:

```go
func Load() (*Config, error) {
    cfg := &Config{}

    cfg.Port = os.Getenv("PORT")
    if cfg.Port == "" {
        cfg.Port = "8080"
    }

    host := os.Getenv("REDIS_HOST")
    port := os.Getenv("REDIS_PORT")
    if host == "" { host = "localhost" }
    if port == "" { port = "6379" }
    cfg.RedisAddr = host + ":" + port

    cfg.JWTSecret = os.Getenv("JWT_SECRET")
    if cfg.JWTSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET environment variable is required")
    }

    // ... more fields with defaults ...
    return cfg, nil
}
```

The critical line: if `JWT_SECRET` is empty, `Load()` returns an error and the server refuses to start. Before, it would boot silently and produce tokens signed with an empty string. Now it fails fast at startup, not at runtime.

`main.go` becomes clean:

```go
cfg, err := config.Load()
if err != nil {
    slog.Error("Failed to load configuration", "error", err)
    os.Exit(1)
}

err = db.InitDB(cfg.DBPath)
err = redis.InitRedis(cfg.RedisAddr)
handlers.RateLimitRequests = cfg.RateLimitRequests
handlers.RateLimitWindow = cfg.RateLimitWindow
```

One source of truth. Every component reads from the same struct. Want to change the rate limit? Change the `RATE_LIMIT_REQUESTS` env var. Don't want to set anything? Defaults kick in. Want to know every config value the app uses? Read one file.

This is the [12-Factor App](https://12factor.net/config) pattern: configuration lives in environment variables, not in code.

## Problem 2: My Backend Can't Handle Failure

With config sorted, I moved to a harder problem. My backend assumed every external service was always available. Redis up? Great. Redis down? Total freeze.

Here's what happened when Redis died:

```
Request 1:  user hits /entries → tries Redis cache → waits 5s timeout → fails → 5s wasted
Request 2:  another user → same thing → 5s wasted
Request 50: still hammering dead Redis → 50 goroutines stuck waiting
Server: technically running, practically dead
```

I needed three things: retry failed connections, time out hung requests, and stop calling dead services.

### Pattern 1: Retry with Exponential Backoff

Sometimes Redis isn't dead — it's just restarting. My Docker setup starts Redis and my app simultaneously. The app tries to connect, Redis isn't ready yet, app crashes. Five seconds later Redis is fine. My app is dead.

```go
func Do(maxAttempts int, initialDelay time.Duration, operation func() error) error {
    var lastErr error
    delay := initialDelay

    for attempt := 1; attempt <= maxAttempts; attempt++ {
        lastErr = operation()
        if lastErr == nil {
            if attempt > 1 {
                slog.Info("Operation succeeded after retry",
                    "attempt", attempt)
            }
            return nil
        }

        if attempt == maxAttempts {
            break
        }

        slog.Warn("Operation failed, retrying",
            "attempt", attempt,
            "next_delay", delay.String(),
            "error", lastErr)

        time.Sleep(delay)
        delay *= 2 // Exponential backoff
    }

    return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}
```

The key insight is `delay *= 2`. Not immediate retries — that hammers a struggling service. Wait 500ms, then 1s, then 2s. Give the service time to recover.

Applied to Redis initialization:

```go
err := retry.Do(3, 500*time.Millisecond, func() error {
    _, err := Client.Ping(context.Background()).Result()
    return err
})
```

Three attempts, starting at 500ms. If Redis needs 2 seconds to boot, my app survives. Before, it crashed on the first failed ping.

The function signature — `operation func() error` — is a higher-order function. `Do()` doesn't know or care *what* it's retrying. Redis ping, database query, HTTP call. Write retry logic once, use it for anything.

### Pattern 2: Request Timeout

Retries handle startup failures. But what about a request that hangs *during* normal operation? A slow database query. A Redis call that never returns. Without a timeout, that goroutine blocks forever — a goroutine leak. Enough leaked goroutines and the server runs out of memory.

```go
func TimeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
        defer cancel()

        r = r.WithContext(ctx)

        done := make(chan struct{})

        go func() {
            next(w, r)
            close(done)
        }()

        select {
        case <-done:
            // Handler finished normally
            return
        case <-ctx.Done():
            // Timeout expired
            slog.Warn("Request timeout",
                "method", r.Method,
                "path", r.URL.Path,
                "timeout", RequestTimeout.String())
            http.Error(w, "Request timed out", http.StatusGatewayTimeout)
            return
        }
    }
}
```

`context.WithTimeout` creates a context that automatically cancels after a deadline. The handler runs in a goroutine. A `select` statement waits for whichever happens first: the handler finishes, or the timeout fires.

If the timeout wins, the client gets a 504 immediately instead of waiting forever. The handler goroutine keeps running in the background (you can't forcefully kill goroutines in Go), but because we passed the timeout context to `r.WithContext(ctx)`, any downstream `db.QueryContext(ctx, ...)` or `redis.Client.Get(ctx, key)` will also abort when the context cancels.

The `select` pattern is one of those Go concurrency primitives that seems simple but solves a hard problem: "wait for A or B, whichever comes first."

### Pattern 3: Circuit Breaker

Retries handle temporary failures. Timeouts prevent hung requests. But neither solves this: Redis is *down*. Not slow, not restarting. Down. Every request still tries Redis, waits for the timeout, and fails. You're burning 5 seconds per request on a service you already know is dead.

The circuit breaker is named after the electrical MCB in your home's fuse box. Overload trips the breaker, cuts the circuit, protects the wiring. Same idea:

```go
type CircuitBreaker struct {
    state           string        // "closed", "open", "half-open"
    failureCount    int
    lastFailureTime time.Time
    threshold       int           // failures before tripping
    cooldown        time.Duration // how long to stay open
    mu              sync.Mutex
}
```

Three states:

**Closed** (normal) — requests flow through. Failures are counted. Hit the threshold? Trip to open.

**Open** (tripped) — all requests rejected *immediately* without calling the service. No 5-second waits. No goroutine leaks. After a cooldown period, move to half-open.

**Half-open** (testing) — allow one request through. If it succeeds, the service recovered — back to closed. If it fails, back to open for another cooldown.

```go
func (cb *CircuitBreaker) Execute(operation func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case "closed":
        err := operation()
        if err != nil {
            cb.failureCount++
            cb.lastFailureTime = time.Now()
            if cb.failureCount >= cb.threshold {
                cb.state = "open"
            }
            return err
        }
        cb.failureCount = 0
        return nil

    case "open":
        if time.Since(cb.lastFailureTime) > cb.cooldown {
            cb.state = "half-open"
            err := operation()
            if err != nil {
                cb.state = "open"
                cb.lastFailureTime = time.Now()
                return err
            }
            cb.state = "closed"
            cb.failureCount = 0
            return nil
        }
        return fmt.Errorf("circuit breaker is open")

    case "half-open":
        // ... same test-and-recover logic
    }
    return nil
}
```

Applied to every Redis operation:

```go
var RedisBreaker = circuitbreaker.NewCircuitBreaker(5, 30*time.Second)

func Get(key string) (string, bool) {
    var result string
    err := RedisBreaker.Execute(func() error {
        var err error
        result, err = redis.Client.Get(context.Background(), key).Result()
        return err
    })
    if err != nil {
        return "", false
    }
    return result, true
}
```

Five failures and the breaker trips. For the next 30 seconds, every Redis call returns instantly with an error instead of waiting for a timeout. After 30 seconds, one test request checks if Redis is back. If yes, normal operation resumes. If no, another 30-second cooldown.

**Without circuit breaker** (Redis down):
```
Request 1-100: each waits 5s timeout → 500s total wasted
```

**With circuit breaker** (Redis down):
```
Request 1-5:   each waits 5s timeout → 25s wasted, breaker trips
Request 6-100: instant reject <1ms each → ~0s wasted
```

The server stays responsive. Users get fast errors instead of frozen pages. And when Redis recovers, traffic flows again automatically.

## How They Work Together

The three patterns complement each other:

**Retry** handles transient failures — a service that's temporarily unavailable during startup or a brief blip. Try again with backoff.

**Timeout** prevents indefinite waits — if a request can't complete in 10 seconds, kill it. Don't let one slow service freeze the whole server.

**Circuit breaker** prevents repeated failures — if a service is genuinely down, stop calling it entirely. Fail fast, protect the server, auto-recover when the service comes back.

## What I Learned

**Config should be boring.** One struct, one loader, fail at startup not runtime. The moment I centralized config, I stopped having "it works on my machine" bugs. Every deployment reads from the same env vars.

**Fail fast beats fail slow.** A 504 in 10 seconds is bad. A 504 in 0ms is useful — the client can retry, show a cached version, or degrade gracefully. The circuit breaker turns slow failures into fast failures.

**Higher-order functions unlock reusable patterns.** `retry.Do(attempts, delay, func() error { ... })` and `breaker.Execute(func() error { ... })` both take functions as arguments. Write the pattern once, wrap any operation. This is how Go avoids needing generics for most resilience code.

**`select` is the most powerful statement in Go.** "Wait for A or B, whichever comes first" — that's the timeout middleware in one sentence. Channels + select is how Go makes concurrency manageable.

---

Up next: I needed my backend to notify external services when things happened — new entries, deleted data. But HTTP calls inside request handlers block the response. So I built a webhook system with fire-and-forget delivery, backed by a worker pool using goroutines and channels. Concurrency in Go, for real this time.

*This is Part 12 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7) | [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp)*