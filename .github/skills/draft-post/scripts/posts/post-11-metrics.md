---
title: I Had No Idea Which Endpoint Was Slowest. So I Built My Own Metrics.
subtitle: Part 11 of Learning Go in Public
tags: go, metrics, backend, buildinpublic
series: Learning Go in Public
---

In [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7), I added structured logging and request IDs. I could now trace a single request through my entire backend — every log line tagged, searchable, debuggable in seconds.

But I still couldn't answer basic questions:

- How many requests did `/entries` handle today?
- What's the average response time for `/login`?
- Which endpoint is the slowest?
- Are errors increasing?

I had detailed logs for individual requests but zero visibility into the system as a whole. That's like having a microscope but no dashboard. I needed metrics.

## What Metrics Actually Are

Logs tell you *what happened* to one request. Metrics tell you *what's happening* across all requests.

Three types:

- **Counters** — only go up. Total requests, total errors. "We've served 50,000 requests since boot."
- **Gauges** — go up and down. In-flight requests right now. "There are 3 requests being processed."
- **Latency stats** — how fast. Average, min, max response time. "/entries averages 45ms but spiked to 2000ms."

I could have used Prometheus. But I wanted to understand what a metrics system actually does before adding a library. So I built one from scratch.

## The Metrics Struct

```go
type Metrics struct {
    mu sync.RWMutex

    // Counters (only increase)
    requestCount map[string]int64
    errorCount   map[string]int64

    // Gauges (go up and down)
    inFlight map[string]int64

    // Latency stats
    totalLatency map[string]float64
    minLatency   map[string]float64
    maxLatency   map[string]float64
}
```

Every field is a `map[string]` keyed by endpoint path. `/entries`, `/login`, `/health` — each gets its own counters.

`sync.RWMutex` is critical. Multiple goroutines (concurrent requests) read and write these maps simultaneously. Without the mutex, you get a race condition — two goroutines incrementing `requestCount["/entries"]` at the same time, one write gets lost.

`RWMutex` specifically lets multiple readers run concurrently (reading metrics is frequent and safe) but locks exclusively for writes. A regular `Mutex` would block all readers while any single request updates its counter.

## Two Functions: Start and Complete

```go
func (m *Metrics) RequestStarted(path string) {
    m.mu.Lock()
    m.inFlight[path]++
    defer m.mu.Unlock()
}

func (m *Metrics) RequestCompleted(path string, duration float64, statusCode int) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.inFlight[path]--
    m.requestCount[path]++

    if statusCode >= 400 {
        m.errorCount[path]++
    }

    m.totalLatency[path] += duration

    if _, exists := m.minLatency[path]; !exists {
        m.minLatency[path] = math.MaxFloat64
    }
    if duration < m.minLatency[path] {
        m.minLatency[path] = duration
    }
    if duration > m.maxLatency[path] {
        m.maxLatency[path] = duration
    }
}
```

`RequestStarted` bumps the in-flight gauge. `RequestCompleted` decrements it, increments the counter, tracks errors (any status >= 400), and updates latency stats. The min latency trick: initialize to `math.MaxFloat64` so the first real duration is always smaller.

## The Problem: How Do You Capture the Status Code?

Here's something I didn't expect. Go's `http.ResponseWriter` doesn't let you *read* the status code after it's been written. Once a handler calls `w.WriteHeader(200)`, that information is gone — it's sent to the client and the `ResponseWriter` doesn't store it.

So how do you know if a request returned 200 or 500?

You wrap it:

```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
    rw.statusCode = statusCode          // Save it
    rw.ResponseWriter.WriteHeader(statusCode) // Pass it through
}
```

This is an **embedding + interception** pattern. The custom `responseWriter` embeds the real one. Every method works normally — except `WriteHeader`, which we override to capture the status code before passing it through.

The handler doesn't know it's writing to a wrapper. It calls `w.WriteHeader(500)` like normal. But now we have the status code stored in `rw.statusCode`.

## The Metrics Middleware

```go
var AppMetrics = metrics.NewMetrics()

func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        AppMetrics.RequestStarted(path)
        start := time.Now()

        wrapped := &responseWriter{
            ResponseWriter: w,
            statusCode:     200, // Default if WriteHeader not called
        }
        next(wrapped, r)

        duration := time.Since(start).Milliseconds()
        AppMetrics.RequestCompleted(path, float64(duration), wrapped.statusCode)
    })
}
```

Before the handler runs: mark the request as in-flight, start the clock.

After the handler runs: stop the clock, record the duration and status code.

Default `statusCode: 200` handles the case where a handler writes a body without explicitly calling `WriteHeader` — Go's `http` package implicitly sends 200 in that case.

## The /metrics Endpoint

```go
func GetMetrics(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    snapshot := AppMetrics.GetSnapshot()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(snapshot)
}
```

`GetSnapshot()` takes a read lock, loops through all tracked endpoints, and builds a response:

```json
{
  "/entries": {
    "total_requests": 1247,
    "errors": 23,
    "in_flight": 2,
    "avg_latency_ms": 45.3,
    "min_latency_ms": 8,
    "max_latency_ms": 2100
  },
  "/login": {
    "total_requests": 89,
    "errors": 12,
    "in_flight": 0,
    "avg_latency_ms": 120.5,
    "min_latency_ms": 95,
    "max_latency_ms": 340
  }
}
```

One GET request and I can see: `/entries` has handled 1,247 requests with a 1.8% error rate, averaging 45ms but occasionally spiking to 2.1 seconds. `/login` is slower on average (bcrypt hashing isn't cheap) but consistent. Two requests are being processed right now.

## The Middleware Chain

MetricsMiddleware wraps every route, sitting between RequestID and RateLimit:

```go
http.HandleFunc("/entries", handlers.RequestIDMiddleware(
    handlers.MetricsMiddleware(handlers.RateLimitMiddleware(
        handlers.LoggingMiddleware(
            handlers.AuthMiddleware(entryHandler))))))
```

The order: RequestID (outermost) assigns the ID, MetricsMiddleware starts the clock, RateLimit checks if the request is allowed, Logging records the request, Auth verifies the token, then the handler runs. When the handler returns, the stack unwinds — MetricsMiddleware stops the clock and records the result.

Five middleware deep. Each one does exactly one thing. The handler doesn't know any of them exist.

## Why Not Just Use Prometheus?

I will, eventually. But building it from scratch taught me things the Prometheus docs don't:

- **Why you need a mutex** — not obvious until two goroutines corrupt your counter
- **The ResponseWriter wrapper pattern** — you can't observe what you can't intercept, and Go's interface doesn't expose the status code
- **The difference between counters and gauges** — sounds trivial until you need to track in-flight requests (which go down, so a counter won't work)
- **How a snapshot works** — read lock, copy the maps, release, encode. Don't hold the lock while serializing JSON

The jump from "hand-built metrics" to Prometheus is just replacing my `Metrics` struct with `prometheus.Counter` and `prometheus.Histogram`. The middleware pattern stays identical.

## What I Learned

**You can't improve what you can't measure.** Logs tell you about failures. Metrics tell you about trends. "Error rate went from 1% to 5% over the last hour" is something logs alone will never show you.

**`sync.RWMutex` exists for a reason.** A regular `Mutex` locks everyone out. `RWMutex` lets concurrent readers through. When your metrics endpoint is being polled every 10 seconds while hundreds of requests update counters, the distinction matters.

**The ResponseWriter wrapper is a pattern you'll use everywhere.** Capturing status codes, response sizes, headers — Go's `http.ResponseWriter` is intentionally minimal. Wrapping it with embedding is how the ecosystem extends it.

**Start with what you understand, replace with tools later.** I'll switch to Prometheus when I need histograms, percentiles, and Grafana dashboards. But I'll know exactly what those tools are doing under the hood because I built the primitive version first.

---

Up next: my backend had hardcoded config values scattered across `main.go`. JWT secrets, Redis hosts, database credentials — all inline. Then I started adding resilience patterns: retry with exponential backoff, request timeouts, and a circuit breaker for Redis. The week my backend learned to survive failures.

*This is Part 11 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7)*