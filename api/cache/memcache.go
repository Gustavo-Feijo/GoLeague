package cache

import (
	"context"
	"sync"
	"time"
)

const (
	cleanupTickerDuration = 5 * time.Minute
)

type MemCache[T any] interface {
	StartCleanupWorker()
	Cleanup()
	Close()
	Get(key string) T
	Set(key string, value T, ttl time.Duration)
}

// Create a in-memory cache with small TTL to minimize Redis calls.
type memCache[T any] struct {
	memoryCache   map[string]*MemCacheItem[T]
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// Simple cache item.
type MemCacheItem[T any] struct {
	value T
	ttl   time.Time
}

// NewMemCache creates a new memory cache.
func NewMemCache[T any]() *memCache[T] {
	ctx, cancel := context.WithCancel(context.Background())
	mc := &memCache[T]{
		memoryCache:   make(map[string]*MemCacheItem[T]),
		cancel:        cancel,
		cleanupTicker: time.NewTicker(cleanupTickerDuration),
		ctx:           ctx,
	}
	mc.StartCleanupWorker()

	return mc
}

// startCleanupWorker starts the background worker for memory cleaning.
func (mc *memCache[T]) StartCleanupWorker() {
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
func (mc *memCache[T]) Cleanup() {
	now := time.Now()
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for key, item := range mc.memoryCache {
		if now.After(item.ttl) {
			delete(mc.memoryCache, key)
		}
	}
}

// Close shutdown the memory cache worker.
func (mc *memCache[T]) Close() {
	mc.cancel()
	mc.cleanupTicker.Stop()
	mc.wg.Wait()
}

// Get returns a key value of the single cache.
func (mc *memCache[T]) Get(key string) T {
	mc.mu.RLock()
	item, exists := mc.memoryCache[key]
	mc.mu.RUnlock()
	if !exists {
		var zero T
		return zero
	}

	// If the reset time was reached, remove the cache.
	if time.Now().After(item.ttl) {
		mc.mu.Lock()
		delete(mc.memoryCache, key)
		mc.mu.Unlock()
		var zero T
		return zero
	}

	return item.value
}

// Set a given key on the cache.
func (mc *memCache[T]) Set(key string, value T, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.memoryCache[key] = &MemCacheItem[T]{
		value: value,
		ttl:   time.Now().Add(ttl),
	}

}
