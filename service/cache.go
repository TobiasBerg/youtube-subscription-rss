package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
)

// FeedCache is a thread-safe in-memory cache for the generated feed XML.
type FeedCache struct {
	mu          sync.RWMutex
	data        []byte
	generatedAt time.Time
	ttl         time.Duration

	cfg         config.AppConfig
	ctx         context.Context
	cancel      context.CancelFunc
	refreshChan chan struct{}
}

// NewFeedCache creates a new FeedCache with the given TTL.
func NewFeedCache(ttl time.Duration, cfg config.AppConfig) *FeedCache {
	ctx, cancel := context.WithCancel(context.Background())
	return &FeedCache{
		ttl:         ttl,
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancel,
		refreshChan: make(chan struct{}, 1),
	}
}

// Start begins the background refresh loop. It performs an initial fetch
// and then refreshes at the configured interval.
func (c *FeedCache) Start() error {
	log.Println("Cache: performing initial feed fetch...")
	if err := c.refresh(); err != nil {
		return err
	}

	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Cache: scheduled refresh")
			if err := c.refresh(); err != nil {
				log.Printf("Cache: refresh failed: %v\n", err)
			}
		case <-c.refreshChan:
			log.Println("Cache: manual refresh triggered")
			if err := c.refresh(); err != nil {
				log.Printf("Cache: refresh failed: %v\n", err)
			}
		case <-c.ctx.Done():
			log.Println("Cache: stopping background refresh")
			return nil
		}
	}
}

// Stop stops the background refresh loop.
func (c *FeedCache) Stop() {
	c.cancel()
}

// Trigger forces an immediate refresh.
func (c *FeedCache) Trigger() {
	select {
	case c.refreshChan <- struct{}{}:
	default:
	}
}

func (c *FeedCache) refresh() error {
	data, err := GenerateFeed(c.ctx, c.cfg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.data = data
	c.generatedAt = time.Now()
	c.mu.Unlock()

	log.Printf("Cache: refreshed feed (%d bytes)", len(data))
	return nil
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
