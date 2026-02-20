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
