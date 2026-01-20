package handlers

import (
	"net/http"
	"sync"
	"time"
)

type RateLimitEntry struct {
	count       int
	WindowStart time.Time
}

type RateLimit struct {
	users map[string]RateLimitEntry
	mu    sync.RWMutex
}

func NewRateLimiting() *RateLimit {
	return &RateLimit{
		users: make(map[string]RateLimitEntry),
	}
}

func (rl *RateLimit) IsAllowed(key string, limit int, window time.Duration) bool {
	rl.mu.Lock() // Use Lock (not RLock) because we'll write
	defer rl.mu.Unlock()

	entry, exists := rl.users[key]
	now := time.Now()

	// Case 1 : Frist request OR Window expired -> start fresh
	if !exists || now.After(entry.WindowStart.Add(window)) {
		rl.users[key] = RateLimitEntry{
			count:       1,
			WindowStart: now,
		}
		return true // Allowed
	}

	// Case 2 : Within window, check if over limit
	if entry.count >= limit {
		return false // Rate limited!
	}

	// Case 3 : Within window, under limit -> increment and allow
	entry.count++
	rl.users[key] = entry
	return true
}

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get identifier - use IP address
		key := r.RemoteAddr

		// 2. Check if allowed (100 requests per minute)
		if !rateLimiter.IsAllowed(key, 100, time.Minute) {
			http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
			return // IMPORTANT: Stop here, don't call next handler
		}

		// 3. If allowed, call next handler
		next(w, r)
	}
}

var rateLimiter = NewRateLimiting()
