package handlers

/*
Rate Limiting Middleware - Fixed Window Algorithm

How it works:
- Track request count per IP address in a time window
- If count exceeds limit → return 429 Too Many Requests
- Window resets after expiration (e.g., 1 minute)

=== WHY REDIS FOR RATE LIMITING? ===

Problem 1: Multiple Servers (In-Memory)
                    Load Balancer
                    /           \
                Server A         Server B
                count: 50        count: 50  (each has its OWN memory)

    User makes 100 requests → 50 go to A, 50 go to B
    Each server thinks: "Only 50 requests, under limit of 100!"
    Reality: User made 100 requests but wasn't blocked! ❌

With Redis:
                    Load Balancer
                    /           \
                Server A         Server B
                    \           /
                      Redis (shared)
                      count: 100  ← Both servers see this! ✅

Problem 2: Server Restart (In-Memory)
    count: 99 (almost blocked!)
    Server restarts...
    count: 0  ← Fresh start, user can abuse again! ❌

With Redis:
    count: 99 → Server restarts → count: 99 still there! ✅

=== REDIS COMMANDS USED ===
INCR key      - Increment by 1 (atomic, creates with value 1 if not exists)
EXPIRE key 60 - Auto-delete key after 60 seconds

See: System design Questions/Day18-Rate-Limiting.md for more details
*/

import (
	"context"
	"net"
	"net/http"
	"personal-analytics-backend/internal/redis"
	"time"
)

func IsAllowed(key string, limit int, window time.Duration) bool {
	// Prefix key to avoid conflicts with other Redis keys
	redisKey := "ratelimit:" + key

	// Step 1: Increment the counter in Redis
	// .Result() returns (newCount, error)
	count, err := redis.Client.Incr(context.Background(), redisKey).Result()
	if err != nil {
		// If Redis fails, allow the request (fail open)
		return true
	}

	// Step 2: If this is the FIRST request (count == 1), set expiration
	// Only set expiration once, not on every request
	if count == 1 {
		redis.Client.Expire(context.Background(), redisKey, window)
	}

	// Step 3: Check if count > limit
	if count > int64(limit) {
		return false // Rate limited!
	}
	return true // Allowed
}

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get identifier - use IP address
		key := r.RemoteAddr
		ip, _, err := net.SplitHostPort(key)

		if err != nil {
			ip = r.RemoteAddr
		}
		// 2. Check if allowed (100 requests per minute)
		if !IsAllowed(ip, 100, time.Minute) {
			http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
			return // IMPORTANT: Stop here, don't call next handler
		}

		// 3. If allowed, call next handler
		next(w, r)
	}
}
