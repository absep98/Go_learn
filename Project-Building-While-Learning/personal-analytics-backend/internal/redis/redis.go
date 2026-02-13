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
