package redis

/*
=== REDIS CACHE INTEGRATION TESTS ===

Run against a real Redis (>= 6.2 for GETDEL). Skipped when REDIS_ADDR is unset:
  REDIS_ADDR=localhost:6379 go test ./internal/redis/

TestRedisCache_GetDel_ConsumesKeyExactlyOnce proves the property refresh-token
rotation depends on: when many clients race to consume the same key, exactly
one of them gets the value. GET followed by DEL does not have this property.
*/

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/redis/go-redis/v9"
)

func newTestCache(t *testing.T) *RedisCache {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("skipping Redis integration test: REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("cannot reach Redis at %s: %v", addr, err)
	}
	t.Cleanup(func() { client.Close() })
	return &RedisCache{client: client}
}

func TestRedisCache_GetDel_ConsumesKeyExactlyOnce(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	const key = "test:getdel:single-use"
	const racers = 50

	for round := 0; round < 20; round++ {
		if err := c.Set(ctx, key, "user-7", 0); err != nil {
			t.Fatalf("set: %v", err)
		}

		var winners atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := 0; i < racers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				v, err := c.GetDel(ctx, key)
				if err == nil && v == "user-7" {
					winners.Add(1)
				} else if err != nil && err != redis.Nil {
					t.Errorf("unexpected error: %v", err)
				}
			}()
		}
		close(start)
		wg.Wait()

		if got := winners.Load(); got != 1 {
			t.Fatalf("round %d: %d of %d concurrent GetDel calls received the value; want exactly 1", round, got, racers)
		}
	}
}

func TestRedisCache_GetDel_MissingKeyReturnsNil(t *testing.T) {
	c := newTestCache(t)
	if _, err := c.GetDel(context.Background(), "test:getdel:never-set"); err != redis.Nil {
		t.Fatalf("expected redis.Nil for a missing key, got %v", err)
	}
}
