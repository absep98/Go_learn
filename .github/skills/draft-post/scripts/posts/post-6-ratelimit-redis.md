---
title: I Built Rate Limiting From Scratch in Go — Then Replaced It With Redis
subtitle: Part 6 of Learning Go in Public
tags: go, buildinpublic, backend, webdev
series: Learning Go in Public
---

In [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk), I added pagination and in-memory caching. Both worked fine — until I thought about what happens with more than one server.

Three days, two rewrites, one lesson: **in-memory is a local solution to a global problem**.

## Day 18: Rate Limiting From Scratch

I wanted to limit each IP to 100 requests per minute. The algorithm is called a **fixed window** — count requests in a time window, reset when the window expires.

I built it with a `map` and a `sync.RWMutex`:

```go
type RateLimitEntry struct {
    count       int
    WindowStart time.Time
}

type RateLimit struct {
    users map[string]RateLimitEntry
    mu    sync.RWMutex
}

func (rl *RateLimit) IsAllowed(key string, limit int, window time.Duration) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    entry, exists := rl.users[key]
    now := time.Now()

    // New IP or expired window → start fresh
    if !exists || now.After(entry.WindowStart.Add(window)) {
        rl.users[key] = RateLimitEntry{count: 1, WindowStart: now}
        return true
    }

    if entry.count >= limit {
        return false // blocked
    }

    entry.count++
    rl.users[key] = entry
    return true
}
```

Wiring it up as middleware was the satisfying part — one function wraps another:

```go
http.HandleFunc("/entries", handlers.RateLimitMiddleware(
    handlers.LoggingMiddleware(
        handlers.AuthMiddleware(handler),
    ),
))
```

The API now returns `429 Too Many Requests` if you hammer it. Done. Except...

## The Problem I Didn't Solve

I was writing system design notes alongside the code, and I drew this:

```
Load Balancer
    /        \
Server A    Server B
count: 50   count: 50   ← each has its OWN memory
```

User makes 100 requests. 50 go to A, 50 to B. Each server thinks "only 50 — under limit!". The user bypassed rate limiting entirely.

Same problem with caching: every server has its own in-memory cache. One server caches a count. Another server doesn't. They disagree. Your data is inconsistent.

In-memory solutions break the moment you run more than one instance.

## Days 19-20: Replacing Everything With Redis

Redis is a shared, external store. Both servers read from and write to the same place. Problem solved.

**Cache before (in-memory):**

```go
type Cache struct {
    data map[string]CacheEntry
    mu   sync.RWMutex
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    entry, exists := c.data[key]
    if !exists || time.Now().After(entry.Expiration) {
        return nil, false
    }
    return entry.value, true
}
```

**Cache after (Redis):**

```go
func Get(key string) (string, bool) {
    result, err := redis.Client.Get(context.Background(), key).Result()
    if err == goredis.Nil {
        return "", false // key doesn't exist
    }
    if err != nil {
        return "", false // other error
    }
    return result, true
}
```

Same function signature. The caller — `entries.go` — didn't change at all. That's separation of concerns working exactly as intended.

**Rate limiting before (in-memory):** the `IsAllowed` function with `sync.RWMutex` above.

**Rate limiting after (Redis):**

```go
func IsAllowed(key string, limit int, window time.Duration) bool {
    redisKey := "ratelimit:" + key

    // Atomic increment — no mutex needed
    count, err := redis.Client.Incr(context.Background(), redisKey).Result()
    if err != nil {
        return true // fail open if Redis is down
    }

    // First request in window → set expiry
    if count == 1 {
        redis.Client.Expire(context.Background(), redisKey, window)
    }

    return count <= int64(limit)
}
```

`INCR` is atomic in Redis. No mutex. No struct. No map. Redis handles the concurrency.

## What I Learned

**`sync.RWMutex` is for single-process concurrency.** Multiple goroutines in one server — yes. Multiple servers — no. When you need distributed state, you need a distributed store.

**The same API surface, different backend.** `cache.Get(key)` works whether the backend is a map or Redis. The handlers don't know or care. That's the value of keeping the interface clean.

**Fail open vs fail closed.** If Redis goes down, my rate limiter returns `true` (allow the request). This is a deliberate choice — I'd rather serve traffic than block everyone because of an infra failure. For some use cases (auth, payments) you'd fail closed. Know which you need.

**`goredis.Nil` is a sentinel error.** When a key doesn't exist, Redis returns a specific error value — not an empty string. This tripped me up the first time.

## The Architecture Now

```
Load Balancer
    /        \
Server A    Server B
    \        /
      Redis
   (shared state)
```

One Redis. Both servers read the same cache, increment the same counters. Scale horizontally without breaking anything.

## What's Next

In Part 7, I'll cover **graceful shutdown and health checks** — what happens when your server needs to stop without dropping requests in flight, and how to expose a `/health` endpoint that actually checks your dependencies.

*This is Part 6 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk)*
