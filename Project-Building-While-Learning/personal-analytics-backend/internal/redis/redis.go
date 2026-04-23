package redis

import (
	"context"
	"log/slog"
	"personal-analytics-backend/internal/retry"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(addr string) error {
	// Create client
	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection with retry (Redis might still be starting up)
	// 3 attempts, starting with 500ms delay (500ms → 1s → 2s)
	err := retry.Do(3, 500*time.Millisecond, func() error {
		_, err := Client.Ping(context.Background()).Result()
		return err
	})
	if err != nil {
		return err
	}

	slog.Info("Redis connected", "addr", addr)
	return nil
}

func CloseRedis() {
	if Client != nil {
		Client.Close()
		slog.Info("Redis connection closed")
	}
}

// RedisCache wraps the redis client to satisfy the handlers.Cache interface
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache() *RedisCache {
	return &RedisCache{client: Client}
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
