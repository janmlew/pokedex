package pokecache

import (
	"sync"
	"time"
)

// cacheEntry is a single cached value together with the time it was stored,
// which the reap loop uses to decide when the entry has gone stale.
type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache is a concurrency-safe, time-expiring key/value store. Entries older
// than the configured interval are periodically removed by a background
// goroutine started in NewCache.
type Cache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
}

// NewCache returns a Cache and starts a background reap loop that evicts
// entries older than interval every interval.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

// Add stores val under key, recording the current time as its creation time.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

// Get returns the cached value for key. The bool is true if the key was
// present and false otherwise.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	return entry.val, ok
}

// reapLoop runs forever, and on every tick of interval removes any entry whose
// createdAt is older than interval.
func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.reap(interval)
	}
}

// reap removes all entries older than interval. It is split out from reapLoop
// so the locking stays scoped to a single tick.
func (c *Cache) reap(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) > interval {
			delete(c.entries, key)
		}
	}
}
