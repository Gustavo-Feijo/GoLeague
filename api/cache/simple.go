package cache

import (
	"sync"
	"time"
)

// Create a in-memory cache with small TTL to minimize Redis calls.
type SimpleCache struct {
	memoryCache map[string]SimpleCacheItem
	mu          sync.RWMutex
}

// Simple cache item.
type SimpleCacheItem struct {
	value any
	ttl   time.Time
}

// Singleton.
var (
	simpleInstance *SimpleCache
	simpleOnce     sync.Once
)

// Get the instance of the simple cache.
func GetSimpleCache() *SimpleCache {
	simpleOnce.Do(func() {
		simpleInstance = &SimpleCache{
			memoryCache: make(map[string]SimpleCacheItem),
		}
	})

	return simpleInstance
}

// Get returns a key value of the single cache.
func (sc *SimpleCache) Get(key string) any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	item, exists := sc.memoryCache[key]
	if !exists {
		return nil
	}

	// If the reset time was reached, remove the cache.
	if time.Now().After(item.ttl) {
		delete(sc.memoryCache, key)
		return nil
	}

	return item.value
}

// Set a given key on the cache.
func (sc *SimpleCache) Set(key string, value any, ttl time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.memoryCache[key] = SimpleCacheItem{
		value: value,
		ttl:   time.Now().Add(ttl),
	}
}
