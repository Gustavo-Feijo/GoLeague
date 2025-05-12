package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/pkg/redis"
	"maps"
	"sync"
	"time"
)

// Create a in-memory cache with small TTL to minimize Redis calls.
type ChampionCache struct {
	redis       *redis.RedisClient
	memoryCache map[string]map[string]any
	TTL         time.Duration
	lastReset   time.Time
	mu          sync.RWMutex
}

// Singleton.
var (
	instance *ChampionCache
	once     sync.Once
)

// Get the instance of the champion cache.
func GetChampionCache() *ChampionCache {
	once.Do(func() {
		instance = &ChampionCache{
			redis:       redis.GetClient(),
			memoryCache: make(map[string]map[string]any),
			TTL:         30 * time.Minute,
			lastReset:   time.Now(),
		}

		// Start the worker that will reset the cache.
		go instance.cacheExpirationWorker()
	})

	return instance
}

// Invalidate the current cache.
func (c *ChampionCache) cacheExpirationWorker() {
	// Create the ticker.
	ticker := time.NewTicker(c.TTL)
	defer ticker.Stop()

	// For each tick, reset the last reset and empty the cache.
	for range ticker.C {
		c.mu.Lock()
		c.memoryCache = make(map[string]map[string]any)
		c.lastReset = time.Now()
		c.mu.Unlock()
	}
}

// Get a champion from the in memory cache, if not already in there, get from the redis.
// Returns a deep copy, so it's safe to change the returned value directly.
func (c *ChampionCache) GetChampionCopy(ctx context.Context, championId string) (map[string]any, error) {
	// Create a copy from the map.
	newMap := make(map[string]any)

	// Try to get directly from memory.
	c.mu.RLock()
	if champCache, exists := c.memoryCache[championId]; exists {
		c.mu.RUnlock()
		maps.Copy(newMap, champCache)
		return newMap, nil
	}
	c.mu.RUnlock()

	// Get from the redis if doesn't found.
	cacheKey := fmt.Sprintf("ddragon:champion:%s", championId)
	champRedis, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		// It should exist on redis, unless it's out.
		return nil, fmt.Errorf("error getting from redis: %w", err)
	}

	// Unmarshal as a generic map.
	// It will be of type champion.Champion.
	var champJson map[string]any
	err = json.Unmarshal([]byte(champRedis), &champJson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal champion data: %w", err)
	}

	// Set the result on the cache.
	c.mu.Lock()
	c.memoryCache[championId] = champJson
	c.mu.Unlock()

	// Create a new map that will be
	maps.Copy(newMap, champJson)

	return newMap, nil
}
