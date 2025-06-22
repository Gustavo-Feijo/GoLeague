package cache

import (
	"context"
	"sync"
	"time"
)

// Create a in-memory cache with small TTL to minimize Redis calls.
type MemCache struct {
	memoryCache   sync.Map
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// Simple cache item.
type MemCacheItem struct {
	value any
	ttl   time.Time
}

// NewMemCache creates a new memory cache.
func NewMemCache() *MemCache {
	ctx, cancel := context.WithCancel(context.Background())
	mc := &MemCache{
		cancel:        cancel,
		cleanupTicker: time.NewTicker(5 * time.Minute),
		ctx:           ctx,
	}
	mc.startCleanupWorker()

	return mc
}

// startCleanupWorker starts the background worker for memory cleaning.
func (mc *MemCache) startCleanupWorker() {
	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		for {
			select {
			case <-mc.cleanupTicker.C:
				mc.cleanup()
			case <-mc.ctx.Done():
				return
			}
		}
	}()
}

// cleanup go through each key and clean any expired key.
func (mc *MemCache) cleanup() {
	now := time.Now()
	mc.memoryCache.Range(func(key, value any) bool {
		item := value.(*MemCacheItem)
		if now.After(item.ttl) {
			mc.memoryCache.Delete(key)
		}
		return true
	})
}

// Close shutdown the memory cache worker.
func (mc *MemCache) Close() {
	mc.cancel()
	mc.cleanupTicker.Stop()
	mc.wg.Wait()
}

// Get returns a key value of the single cache.
func (mc *MemCache) Get(key string) any {
	value, exists := mc.memoryCache.Load(key)
	if !exists {
		return nil
	}

	item := value.(*MemCacheItem)

	// If the reset time was reached, remove the cache.
	if time.Now().After(item.ttl) {
		mc.memoryCache.Delete(key)
		return nil
	}

	return item.value
}

// Set a given key on the cache.
func (mc *MemCache) Set(key string, value any, ttl time.Duration) {
	mc.memoryCache.Store(key, &MemCacheItem{
		value: value,
		ttl:   time.Now().Add(ttl),
	})
}
