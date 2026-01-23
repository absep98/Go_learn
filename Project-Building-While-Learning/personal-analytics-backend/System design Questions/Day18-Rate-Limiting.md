# Day 18: Rate Limiting - System Design

## What We Built

In-memory rate limiter using sliding window approach:

- Track request count per IP address
- Reset count when time window expires
- Return 429 Too Many Requests when limit exceeded

## The Algorithm (Fixed Window)

```
Request comes in:
1. Get user's entry from map
2. If no entry OR window expired → Create new (count=1) → Allow
3. If count < limit → Increment → Allow
4. If count >= limit → Block (429)
```

---

## Problem 1: Multiple Servers (Split Brain)

### The Issue

```
                    Load Balancer
                    /           \
                Server A         Server B
                count: 50        count: 50

User makes 100 requests → 50 go to A, 50 go to B
Each server thinks: "Only 50 requests, under limit!"
Reality: User made 100 requests but wasn't blocked! ❌
```

### Why It Happens

- Each server has its **own local memory**
- Rate limit data is **not shared** between servers
- User can bypass limit by: `limit × number_of_servers`

### Solution: Centralized Storage (Redis)

```go
// Instead of local map:
count := redis.INCR("ratelimit:user:123")
redis.EXPIRE("ratelimit:user:123", 60)
if count > 100 { block }
```

---

## Problem 2: Server Restart (Data Loss)

### The Issue

```
Before restart:  count: 99 (almost at limit!)
Server crashes or restarts...
After restart:   count: 0  (fresh start)

User continues without any limit! ❌
```

### Why It Happens

- In-memory data lives only while process runs
- Restart = all data gone
- No persistence

### Solution: External Storage

- Redis persists data outside the application
- Data survives application restarts
- Can even configure Redis persistence to disk

---

## Comparison Table

| Aspect | In-Memory (Our Implementation) | Redis (Production) |
|--------|-------------------------------|-------------------|
| Speed | Fastest (nanoseconds) | Fast (milliseconds) |
| Shared across servers | ❌ No | ✅ Yes |
| Survives restart | ❌ No | ✅ Yes |
| Complexity | Simple | Requires Redis setup |
| Use case | Single server, dev/test | Production, multi-server |

---

## Rate Limiting Algorithms

### 1. Fixed Window (What we built)

- Count requests in fixed time windows (e.g., 0:00-1:00, 1:00-2:00)
- Simple but has edge case: burst at window boundary

### 2. Sliding Window

- Counts requests in rolling window (last 60 seconds from NOW)
- Smoother, prevents boundary bursts
- Slightly more complex

### 3. Token Bucket

- Bucket holds tokens, refills at steady rate
- Each request consumes a token
- Allows bursts (if bucket is full)
- Used by AWS, Stripe

### 4. Leaky Bucket

- Requests queue up, processed at constant rate
- Smooths out traffic
- Good for consistent processing rate

---

## Interview Answer

**Q: "How would you implement rate limiting for a distributed system?"**

**A:** "For a single server, an in-memory solution with a map and mutex works well. But for distributed systems, we need centralized storage like Redis.

Key approach:

1. Use Redis INCR to atomically increment request count
2. Set TTL for automatic expiration (sliding window)
3. Check count before processing request
4. Return 429 if limit exceeded

For high-traffic APIs, I'd use Token Bucket algorithm - it allows controlled bursts while maintaining average rate limits. This is what AWS and Stripe use."

---

## Code Structure Reference

```go
type RateLimitEntry struct {
    count       int        // Request count in window
    WindowStart time.Time  // When window started
}

type RateLimit struct {
    users map[string]RateLimitEntry
    mu    sync.RWMutex  // Thread safety
}

func IsAllowed(key string, limit int, window time.Duration) bool
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc
```

---

# Day 19-20: Migration from In-Memory to Redis

## Why We Migrated

Both **Cache** and **Rate Limiting** had the same problems with in-memory maps:

1. ❌ Data lost on server restart
2. ❌ Each server has its own copy (can't scale horizontally)
3. ❌ Memory fills up over time (no automatic cleanup)

---

## Cache: Before vs After

### BEFORE (In-Memory Map)

```go
// internal/cache/cache.go

type CacheEntry struct {
    Value      interface{}
    Expiration time.Time
}

type Cache struct {
    data map[string]CacheEntry
    mu   sync.RWMutex  // Mutex for thread safety
}

var AppCache *Cache  // Global variable

func NewCache() *Cache {
    return &Cache{
        data: make(map[string]CacheEntry),
    }
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()         // Lock for reading
    defer c.mu.RUnlock()

    entry, exists := c.data[key]
    if !exists || time.Now().After(entry.Expiration) {
        return nil, false
    }
    return entry.Value, true
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()          // Lock for writing
    defer c.mu.Unlock()

    c.data[key] = CacheEntry{
        Value:      value,
        Expiration: time.Now().Add(ttl),
    }
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.data, key)
}
```

**Problems:**

- Manual expiration check (CPU waste)
- Memory grows indefinitely (need cleanup goroutine)
- Lost on restart
- Not shared across servers

### AFTER (Redis)

```go
// internal/cache/cache.go

import (
    "context"
    "personal-analytics-backend/internal/redis"
    "time"
)

func Get(key string) (string, bool) {
    val, err := redis.Client.Get(context.Background(), key).Result()
    if err != nil {
        return "", false
    }
    return val, true
}

func Set(key string, value interface{}, ttl time.Duration) {
    redis.Client.Set(context.Background(), key, value, ttl)
}

func Delete(key string) {
    redis.Client.Del(context.Background(), key)
}
```

**Improvements:**

- ✅ Automatic TTL (Redis handles expiration)
- ✅ No memory management needed
- ✅ Survives restart
- ✅ Shared across all servers

---

## Rate Limiting: Before vs After

### BEFORE (In-Memory Map)

```go
// internal/handlers/ratelimit.go

type RateLimitEntry struct {
    count       int
    WindowStart time.Time
}

type RateLimit struct {
    users map[string]RateLimitEntry
    mu    sync.RWMutex
}

var limiter = &RateLimit{
    users: make(map[string]RateLimitEntry),
}

func (rl *RateLimit) IsAllowed(key string, limit int, window time.Duration) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    entry, exists := rl.users[key]
    now := time.Now()

    // Check if window expired or new user
    if !exists || now.Sub(entry.WindowStart) > window {
        rl.users[key] = RateLimitEntry{count: 1, WindowStart: now}
        return true
    }

    // Increment and check
    entry.count++
    rl.users[key] = entry
    return entry.count <= limit
}
```

**Problems:**

- Each server tracks separately
- Attacker can bypass by hitting different servers
- Reset on restart = abuse window

### AFTER (Redis)

```go
// internal/handlers/ratelimit.go

func IsAllowed(key string, limit int, window time.Duration) bool {
    redisKey := "ratelimit:" + key

    // INCR is atomic! No race conditions
    count, err := redis.Client.Incr(context.Background(), redisKey).Result()
    if err != nil {
        return true  // Fail open
    }

    // Set expiration only on first request
    if count == 1 {
        redis.Client.Expire(context.Background(), redisKey, window)
    }

    return count <= int64(limit)
}
```

**Improvements:**

- ✅ Atomic INCR (no race conditions, no mutex needed)
- ✅ All servers share the same counter
- ✅ Automatic TTL cleanup
- ✅ Survives restart

---

## Key Concepts Learned

### 1. Mutex vs Redis Atomic Operations

| In-Memory | Redis |
|-----------|-------|
| `mu.Lock()` | Not needed! |
| Manual increment | `INCR` is atomic |
| Race conditions possible | Built-in thread safety |

### 2. Manual TTL vs Automatic TTL

```go
// In-Memory: Check expiration every time
if time.Now().After(entry.Expiration) {
    // Expired, need to handle
}

// Redis: Set once, forget
redis.Client.Set(ctx, key, value, 5*time.Minute)
// Redis auto-deletes after 5 minutes
```

### 3. Global Variable vs Shared Storage

```go
// In-Memory: Each server has its own
var AppCache *Cache  // Server A has different data than Server B

// Redis: All servers connect to same Redis
redis.Client.Get(ctx, "key")  // Same data everywhere
```

---

## Redis Commands Cheat Sheet

| Command | What it does | Example |
|---------|--------------|---------|
| `SET key value` | Store a value | `SET user:1 "John"` |
| `GET key` | Retrieve a value | `GET user:1` |
| `SETEX key ttl value` | Set with expiration | `SETEX session 3600 "abc"` |
| `INCR key` | Increment (creates with 1 if not exists) | `INCR counter` |
| `EXPIRE key seconds` | Set TTL on existing key | `EXPIRE mykey 60` |
| `TTL key` | Check remaining TTL | `TTL mykey` |
| `DEL key` | Delete a key | `DEL user:1` |
| `KEYS pattern` | Find keys matching pattern | `KEYS ratelimit:*` |
| `FLUSHALL` | Delete ALL keys (careful!) | `FLUSHALL` |

---

## File Changes Summary

| File | Change |
|------|--------|
| `internal/redis/redis.go` | NEW - Redis connection setup |
| `internal/cache/cache.go` | REWRITTEN - Map → Redis |
| `internal/handlers/ratelimit.go` | REWRITTEN - Map → Redis |
| `cmd/server/main.go` | Added Redis init/close |
| `internal/db/db.go` | Updated to use new cache API |
| `internal/handlers/entries.go` | Updated cache.Delete() call |

---

## Gotcha: Extracting IP from RemoteAddr

```go
// r.RemoteAddr returns IP:Port like "[::1]:54321"
// Need to extract just the IP for rate limiting

ip, _, err := net.SplitHostPort(r.RemoteAddr)
if err != nil {
    ip = r.RemoteAddr  // Fallback
}

// Now ip = "::1" (without port)
```

**Why this matters:**

- Without this fix: Each request from same IP gets different key (different port)
- With fix: All requests from same IP share one counter
