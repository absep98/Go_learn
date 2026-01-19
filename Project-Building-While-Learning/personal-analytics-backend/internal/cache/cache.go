package cache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	value      interface{}
	Expiration time.Time
}

type Cache struct {
	data map[string]CacheEntry
	mu   sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]CacheEntry),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {

	c.mu.RLock()         // Lock for reading
	defer c.mu.RUnlock() // Unlock when function ends

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// check if expired
	if time.Now().After(entry.Expiration) {
		return nil, false // Expiration = not found
	}

	return entry.value, true

}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {

	c.mu.Lock()
	defer c.mu.Unlock()

	expiration_time := time.Now().Add(ttl)

	c.data[key] = CacheEntry{
		value:      value,
		Expiration: expiration_time,
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

var AppCache = NewCache()
