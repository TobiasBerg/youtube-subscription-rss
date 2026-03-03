package service

import (
	"sync"
	"time"
)

// FeedCache is a thread-safe in-memory cache for the generated feed XML.
type FeedCache struct {
	mu          sync.RWMutex
	data        []byte
	generatedAt time.Time
	ttl         time.Duration
}

// NewFeedCache creates a new FeedCache with the given TTL.
func NewFeedCache(ttl time.Duration) *FeedCache {
	return &FeedCache{ttl: ttl}
}

// Get returns the cached feed bytes and true if the cache is still valid,
// or nil and false if the cache is empty or has expired.
func (c *FeedCache) Get() ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil || time.Since(c.generatedAt) >= c.ttl {
		return nil, false
	}

	return c.data, true
}

// Set stores new feed bytes in the cache and records the current time.
func (c *FeedCache) Set(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.generatedAt = time.Now()
}
