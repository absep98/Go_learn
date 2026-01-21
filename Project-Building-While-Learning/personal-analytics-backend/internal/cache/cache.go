package cache

import (
	"context"
	"time"

	"personal-analytics-backend/internal/redis" // Your redis package with Client

	goredis "github.com/redis/go-redis/v9" // For goredis.Nil error check
)

// Get retrieves a value from Redis
// Returns (value, true) if found, ("", false) if not found or error
func Get(key string) (string, bool) {
	// redis.Client = your connection to Redis (from internal/redis/redis.go)
	// context.Background() = empty context (no timeout/cancellation)
	// .Result() = execute command and get (value, error)
	result, err := redis.Client.Get(context.Background(), key).Result()

	// goredis.Nil = special error meaning "key doesn't exist"
	if err == goredis.Nil {
		return "", false // Not found
	}
	if err != nil {
		return "", false // Some other error
	}
	return result, true // Found!
}

// Set stores a value in Redis with TTL (auto-expires)
func Set(key string, value interface{}, ttl time.Duration) {
	// Redis Set automatically handles expiration via ttl parameter
	redis.Client.Set(context.Background(), key, value, ttl)
}

// Delete removes a key from Redis
func Delete(key string) {
	// Del = delete command in Redis
	redis.Client.Del(context.Background(), key)
}
