package cache

import (
	"sync"
	"time"
)

type ResourceCache struct {
	items           map[string]cacheItem
	mu              sync.RWMutex
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopChan        chan struct{}
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func NewResourceCache(defaultTTL, cleanupInterval time.Duration) *ResourceCache {
	c := &ResourceCache{
		items:           make(map[string]cacheItem),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}
	go c.startCleanup()
	return c
}

// Set adds an item to the cache with optional TTL
func (c *ResourceCache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves an item from the cache if it exists and isn't expired
func (c *ResourceCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiration) {
		return nil, false
	}
	return item.value, true
}

// startCleanup begins periodic cache cleanup
func (c *ResourceCache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopChan:
			return
		}
	}
}

// cleanupExpired removes expired items
func (c *ResourceCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

// Stop terminates the background cleanup goroutine
func (c *ResourceCache) Stop() {
	close(c.stopChan)
}
