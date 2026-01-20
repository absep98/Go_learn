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

### The Issue:
```
                    Load Balancer
                    /           \
                Server A         Server B
                count: 50        count: 50
                
User makes 100 requests → 50 go to A, 50 go to B
Each server thinks: "Only 50 requests, under limit!"
Reality: User made 100 requests but wasn't blocked! ❌
```

### Why It Happens:
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

### The Issue:
```
Before restart:  count: 99 (almost at limit!)
Server crashes or restarts...
After restart:   count: 0  (fresh start)

User continues without any limit! ❌
```

### Why It Happens:
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
