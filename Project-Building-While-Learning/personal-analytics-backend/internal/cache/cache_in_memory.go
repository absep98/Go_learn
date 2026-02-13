package cache

/*
=== IN-MEMORY CACHE WITH BACKGROUND CLEANUP ===

Problem: When cache entries expire, Get() returns false but the DATA
stays in memory forever. After months: thousands of dead entries = memory leak!

Solution: Background goroutine runs every 30 seconds, scans all entries,
deletes expired ones. Like a janitor who sweeps the building every night.

=== WHY time.NewTicker? ===

time.Sleep(30s) in a loop ALSO works, but Ticker is more idiomatic Go:
- Ticker compensates for processing time (stays on schedule)
- Sleep doesn't: if cleanup takes 5s, next run is 35s later
- Ticker can be stopped cleanly (ticker.Stop())

=== CLEANUP GOROUTINE LIFECYCLE ===

1. NewCache() creates cache + starts cleanup goroutine
2. Goroutine runs forever (stopped via stopCleanup channel)
3. Every 30 seconds: Lock map → scan all entries → delete expired → Unlock
4. On shutdown: close(stopCleanup) tells goroutine to exit

=== WHY WRITE LOCK DURING CLEANUP? ===

cleanup uses c.mu.Lock() (not RLock) because it DELETES entries.
This briefly blocks readers, but:
- Runs only every 30 seconds (not on every request)
- Scan is fast (just comparing timestamps)
- Alternative (no cleanup) = memory leak = much worse
*/

import (
	"log/slog"
	"sync"
	"time"
)

type CacheEntry struct {
	value      interface{}
	Expiration time.Time
}

type Cache struct {
	data        map[string]CacheEntry
	mu          sync.RWMutex
	stopCleanup chan struct{} // Signal channel to stop cleanup goroutine
}

// NewCache creates a new cache and starts the background cleanup goroutine
func NewCache() *Cache {
	c := &Cache{
		data:        make(map[string]CacheEntry),
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go c.cleanupLoop()

	return c
}

// cleanupLoop runs every 30 seconds, deleting expired entries
// This prevents memory leaks from expired-but-never-read entries
func (c *Cache) cleanupLoop() {
	// time.NewTicker sends a value every 30 seconds
	// More accurate than time.Sleep because it compensates for processing time
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop() // Clean up ticker when goroutine exits

	for {
		select {
		case <-ticker.C:
			// Ticker fired — time to clean up
			c.deleteExpired()

		case <-c.stopCleanup:
			// Someone called StopCleanup() — exit the goroutine
			slog.Info("Cache cleanup goroutine stopped")
			return
		}
	}
}

// deleteExpired scans all entries and deletes expired ones
func (c *Cache) deleteExpired() {
	now := time.Now()
	expired := 0

	c.mu.Lock() // Write lock — we're deleting entries
	for key, entry := range c.data {
		if now.After(entry.Expiration) {
			delete(c.data, key)
			expired++
		}
	}
	c.mu.Unlock()

	if expired > 0 {
		slog.Debug("Cache cleanup completed", "expired_entries_removed", expired, "remaining_entries", len(c.data))
	}
}

// StopCleanup signals the cleanup goroutine to stop
// Call this during graceful shutdown
func (c *Cache) StopCleanup() {
	close(c.stopCleanup) // Closing a channel unblocks all receivers
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
