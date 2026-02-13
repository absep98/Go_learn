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

// Global rate limit configuration (set from main)
// Exported (capitalized) so main package can modify them
//
// === PACKAGE VISIBILITY & VARIABLE INITIALIZATION PATTERN ===
//
// Q: Why declare variables here with defaults instead of in main.go?
// A: Variables must live in the package where they're USED.
//
// EXECUTION ORDER:
// 1. Package Init: These variables are initialized with defaults (100, 1min)
// 2. main() runs:  main.go loads config and OVERWRITES these values
// 3. Middleware:   Uses the UPDATED values from config, not defaults
//
// Example:
//   ratelimit.go:  var RateLimitRequests = 100  (initial default)
//   main.go:       handlers.RateLimitRequests = 200  (overwrites with config)
//   middleware:    IsAllowed(ip, RateLimitRequests, ...)  (uses 200, not 100!)
//
// KEY RULES:
// 1. Variable Lives Where Used: RateLimitRequests declared in handlers package
// 2. Same Package = Direct Access: Code here uses RateLimitRequests directly
// 3. Cross-Package = Prefix: main.go uses handlers.RateLimitRequests
// 4. Capitalization = Export: Capital letter makes it visible to other packages
//
// Why not declare in main.go?
//   main package variables are invisible to handlers package (scope isolation)
//
var (
	RateLimitRequests = 100        // Default: 100 requests
	RateLimitWindow   = time.Minute // Default: 1 minute
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
		// 2. Check if allowed (configured rate limit)
		if !IsAllowed(ip, RateLimitRequests, RateLimitWindow) {
			http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
			return // IMPORTANT: Stop here, don't call next handler
		}

		// 3. If allowed, call next handler
		next(w, r)
	}
}
