package cache

import (
	"context"
	"sync"
	"time"
)

type MemCache interface {
	StartCleanupWorker()
	Cleanup()
	Close()
	Get(key string) any
	Set(key string, value any, ttl time.Duration)
}

// Create a in-memory cache with small TTL to minimize Redis calls.
type memCache struct {
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
func NewMemCache() *memCache {
	ctx, cancel := context.WithCancel(context.Background())
	mc := &memCache{
		cancel:        cancel,
		cleanupTicker: time.NewTicker(5 * time.Minute),
		ctx:           ctx,
	}
	mc.StartCleanupWorker()

	return mc
}

// startCleanupWorker starts the background worker for memory cleaning.
func (mc *memCache) StartCleanupWorker() {
	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		for {
			select {
			case <-mc.cleanupTicker.C:
				mc.Cleanup()
			case <-mc.ctx.Done():
				return
			}
		}
	}()
}

// cleanup go through each key and clean any expired key.
func (mc *memCache) Cleanup() {
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
func (mc *memCache) Close() {
	mc.cancel()
	mc.cleanupTicker.Stop()
	mc.wg.Wait()
}

// Get returns a key value of the single cache.
func (mc *memCache) Get(key string) any {
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
func (mc *memCache) Set(key string, value any, ttl time.Duration) {
	mc.memoryCache.Store(key, &MemCacheItem{
		value: value,
		ttl:   time.Now().Add(ttl),
	})
}
