package redis

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(addr string) error {
	// Create client
	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection
	_, err := Client.Ping(context.Background()).Result()
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
