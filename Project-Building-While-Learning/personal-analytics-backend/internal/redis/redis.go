package redis

import (
	"context"
	"log"

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

	log.Println("Redis connected successfully")
	return nil
}

func CloseRedis() {
	if Client != nil {
		Client.Close()
		log.Println("Redis connection closed")
	}
}
