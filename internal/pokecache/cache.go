package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu       sync.Mutex
	store    map[string]cacheEntry
	interval time.Duration
}

// NewCache creates a cache whose entries expire after `interval`.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		store:    make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

// Add stores val under key.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = cacheEntry{createdAt: time.Now(), val: val}
}

// Get returns the cached value (copy) and true if present.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.store[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

// periodically deletes expired entries
func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		for k, v := range c.store {
			if time.Since(v.createdAt) > c.interval {
				delete(c.store, k)
			}
		}
		c.mu.Unlock()
	}
}
