---
title: My Server Was Silently Dropping Requests — Here's the One-Line Fix That Changed Everything
subtitle: Part 7 of Learning Go in Public
tags: go, buildinpublic, backend, webdev
series: Learning Go in Public
---

In [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i), I added rate limiting and Redis. The API was getting real — pagination, caching, authentication, rate limiting. It handled traffic. It scaled horizontally.

Then I pressed `Ctrl+C` and realized something: **every request in flight just vanished.**

## The Problem Nobody Warns You About

Here's what my server startup looked like before:

```go
log.Printf("Server starting on port %s", port)
http.ListenAndServe(":"+port, nil)
```

Two lines. Works fine. Except when you stop it.

`http.ListenAndServe` blocks forever. When you press `Ctrl+C`, the process gets killed instantly. Any request that was mid-response? Gone. Any database transaction that was half-committed? Corrupted. Any user waiting for a response? Timeout.

In development this doesn't matter. In production, this is how you lose data.

## Day 21: Teaching the Server How to Die

The fix required rethinking how the server starts. Instead of `ListenAndServe` (which blocks), I needed to:

1. Start the server in a **goroutine** (background)
2. Listen for shutdown signals on the **main thread**
3. When Ctrl+C hits, **stop accepting new requests** but **finish current ones**

```go
// Step 1: Create a server (so we can call Shutdown later)
server := &http.Server{Addr: ":" + port}

// Step 2: Start in background
go func() {
    log.Printf("Server starting on port %s", port)
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}()

// Step 3: Wait for Ctrl+C
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit  // blocks here until signal received
```

That `<-quit` line is where I had my "aha" moment with Go channels. It's not a sleep. It's not polling. The goroutine just *parks* itself until a signal arrives. Zero CPU usage while waiting.

## The Graceful Part

When `Ctrl+C` hits, the signal flows through the channel, and shutdown begins:

```go
log.Println("Shutdown signal received...")

// Give in-flight requests 5 seconds to finish
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// This is the magic line:
server.Shutdown(ctx)
```

`server.Shutdown(ctx)` does two things simultaneously:
- **Stops accepting** new connections immediately
- **Waits** for existing requests to complete (up to 5 seconds)

If a request is mid-response, it gets those 5 seconds to finish. If it takes longer — force close. No data loss in the normal case. No hanging forever in the worst case.

After that, close Redis, close the database, log "Server stopped gracefully", exit clean.

## Day 22: The Health Check That Actually Checks

While I was thinking about server lifecycle, I looked at my health endpoint:

```go
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "OK")
}
```

This is lying. It says "OK" even if the database is down. Even if Redis is dead. It's not a health check — it's a health *lie*.

I replaced it with one that actually pings dependencies:

```go
type HealthResponse struct {
    Status   string `json:"status"`
    Database string `json:"database"`
    Redis    string `json:"redis"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{
        Status:   "healthy",
        Database: "connected",
        Redis:    "connected",
    }

    if err := redis.Client.Ping(context.Background()).Err(); err != nil {
        response.Redis = "disconnected"
        response.Status = "unhealthy"
    }

    if err := db.DB.Ping(); err != nil {
        response.Database = "disconnected"
        response.Status = "unhealthy"
    }

    if response.Status == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable) // 503
    }
    json.NewEncoder(w).Encode(response)
}
```

Now when a load balancer hits `/health`, it gets a real answer:

```json
{"status": "healthy", "database": "connected", "redis": "connected"}
```

Or:

```json
{"status": "unhealthy", "database": "connected", "redis": "disconnected"}
```

The 503 status code tells the load balancer: "stop sending me traffic." That's how production systems route around failure — not by guessing, but by asking.

## What I Learned

**`http.ListenAndServe` is a learning tool, not a production tool.** The moment your server handles real traffic, you need `http.Server` + `Shutdown()`. It's 15 extra lines of code. There's no excuse to skip it.

**Channels aren't just for concurrency.** `<-quit` is a channel used for *coordination* — the main goroutine waits for an OS signal without burning CPU. Channels are Go's answer to "how do goroutines talk to each other?"

**Health checks should fail loudly.** A health endpoint that always returns 200 is worse than no health endpoint at all. It gives you false confidence. If it can't verify dependencies, it should say so.

## The Server Now

```
Ctrl+C pressed
  → Signal hits quit channel
  → Stop accepting new connections
  → Wait up to 5 seconds for in-flight requests
  → Close Redis
  → Close database
  → Exit 0
```

Clean lifecycle. No dropped requests. No corrupted transactions. And a health check that actually tells the truth.

## What's Next

In Part 8, I'll cover **Docker and Postgres** — containerizing the entire stack and migrating from SQLite to a real database. The moment this thing stopped being a toy and started being deployable.

*This is Part 7 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i)*
