package cache

import (
	"context"
	"personal-analytics-backend/internal/circuitbreaker"
	"personal-analytics-backend/internal/redis" // Your redis package with Client
	"time"
	// For goredis.Nil error check
)

// Ciricut breaker for Redis operations
// 5 failures -> open ciricut -> 30 second cooldown.
var RedisBreaker = circuitbreaker.NewCircuitBreaker(5, 30*time.Second)

// Get retrieves a value from Redis
// Returns (value, true) if found, ("", false) if not found or error
func Get(key string) (string, bool) {
	// redis.Client = your connection to Redis (from internal/redis/redis.go)
	// context.Background() = empty context (no timeout/cancellation)
	// .Result() = execute command and get (value, error)
	var result string

	err := RedisBreaker.Execute(func() error {
		var err error
		result, err = redis.Client.Get(context.Background(), key).Result()
		return err
	})

	// goredis.Nil = special error meaning "key doesn't exist"
	if err != nil {
		return "", false // Some other error
	}
	return result, true // Found!
}

// Set stores a value in Redis with TTL (auto-expires)
func Set(key string, value interface{}, ttl time.Duration) {
	// Redis Set automatically handles expiration via ttl parameter
	RedisBreaker.Execute(func() error {
		return redis.Client.Set(context.Background(), key, value, ttl).Err()
	})

}

// Delete removes a key from Redis
func Delete(key string) {
	// Del = delete command in Redis
	RedisBreaker.Execute(func() error {
		return redis.Client.Del(context.Background(), key).Err()
	})
}

/*
=== INTERVIEW ANSWER: CACHING STRATEGY ===

WHAT WE CACHE:
The result of GET /entries — the list of entries per user.
Key format: "entries:<userID>". Not every endpoint — only expensive read-heavy ones.

WHY CACHE AT ALL:
- DB queries are slow: disk I/O, query planning, network round-trip.
- Redis is in-memory: ~10x faster than DB.
- If 1000 users request the same data, hit DB once, serve 999 from cache.
- Reduces DB load — fewer connections, less CPU on DB server.

CANCHE-ASIDE PATTERN (what we use):
1. GET request arrives
2. Check Redis first (cache.Get)
3. Hit → return cached value immediately, skip DB
4. Miss → query DB → store result in Redis (cache.Set with TTL) → return

CACHE INVALIDATION:
Any write operation (create/update/delete entry) deletes the cached key.
Reason: cached list is now stale — it doesn't reflect the change.
Next GET will miss cache, fetch fresh data from DB, re-cache updated list.
TTL is a safety net: even if invalidation has a bug, stale data expires automatically.

WHY REDIS OVER IN-MEMORY MAP:
1. Survives server restarts — Redis is a separate process. Go map dies with the process.
2. Shared across multiple servers — in-memory maps are per-process.
   With 3 servers behind a load balancer, each has its own map = inconsistent cache.
   Redis is a distributed cache — all servers share one view of the data.

TTL PURPOSE:
- Primary: auto-expire data that's no longer relevant (saves memory).
- Safety net: bounds how stale data can get if invalidation fails.
- In your system: entries cached for duration of request (or explicit TTL set).

CIRCUIT BREAKER INTEGRATION:
All three cache operations (Get/Set/Delete) go through RedisBreaker.
If Redis goes down: after 5 failures, breaker opens → instant rejection.
System degrades gracefully: cache misses → all requests hit DB directly.
DB handles more load but server stays functional — Redis failure is not fatal.

TRADE-OFFS:
- Cache invalidation is hard to get right — missed Delete = stale data served
- No cache warming on startup — first requests after restart always hit DB
- Production: consider read-through cache, write-through, or event-driven invalidation
*/
