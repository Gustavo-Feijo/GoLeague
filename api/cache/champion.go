package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/api/repositories"
	"goleague/pkg/redis"
	"sync"
	"time"
)

// Create a in-memory cache with small TTL to minimize Redis calls.
type ChampionCache struct {
	redis       *redis.RedisClient
	memoryCache sync.Map
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
			redis:     redis.GetClient(),
			TTL:       30 * time.Minute,
			lastReset: time.Now(),
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
		c.memoryCache = sync.Map{}
		c.mu.Lock()
		c.lastReset = time.Now()
		c.mu.Unlock()
	}
}

// Get a champion from the in memory cache, if not already in there, get from the redis.
// Returns a deep copy, so it's safe to change the returned value directly.
func (c *ChampionCache) GetChampionCopy(ctx context.Context, championId string, repo repositories.CacheRepository) (map[string]any, error) {
	// Create a copy from the map.

	// Try to get directly from memory.
	if champCache, exists := c.memoryCache.Load(championId); exists {
		if champEntry, ok := champCache.(map[string]any); ok {
			return deepCopyMap(champEntry), nil
		}
	}

	// Get from the redis if doesn't found.
	cacheKey := fmt.Sprintf("ddragon:champion:%s", championId)
	champRedis, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		if repo == nil {
			// It should exist on redis, unless it's out.
			return nil, fmt.Errorf("error getting from redis: %w", err)
		}

		// Get from the database fallback in that case.
		// It will be way slower, but will save in memory for the next requests.
		champRedis, err = repo.GetKey(cacheKey)
		if err != nil {
			// Everything went wrong.
			return nil, fmt.Errorf("error getting from the database fallback: %w", err)
		}
	}

	// Unmarshal as a generic map.
	// It will be of type champion.Champion.
	var champJson map[string]any
	err = json.Unmarshal([]byte(champRedis), &champJson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal champion data: %w", err)
	}

	c.memoryCache.Store(championId, champJson)
	return deepCopyMap(champJson), nil

}

func (c *ChampionCache) Initialize(ctx context.Context) error {
	cachePrefix := "ddragon:champion:"

	// Get all the keys by prefix.
	keys, err := redis.GetClient().GetKeysByPrefix(ctx, cachePrefix)
	if err != nil {
		repo, err := repositories.NewCacheRepository()
		if err != nil {
			return err
		}

		// Get all champions by the prefix and save in memory.
		champions, _ := repo.GetByPrefix(cachePrefix)
		for _, champion := range champions {
			var champJson map[string]any
			err := json.Unmarshal([]byte(champion.CacheValue), &champJson)
			if err != nil {
				return err
			}
			championId := champJson["id"].(string)
			c.memoryCache.Store(championId, champJson)
		}
		return nil
	}

	// Loop through each redis key and store it on memory.
	for _, key := range keys {
		champRedis, err := redis.GetClient().Get(ctx, key)
		if err != nil {
			continue
		}
		var champJson map[string]any
		err = json.Unmarshal([]byte(champRedis), &champJson)
		if err != nil {
			return fmt.Errorf("failed to unmarshal champion data: %w", err)
		}
		championId := champJson["id"].(string)
		c.memoryCache.Store(championId, champJson)
	}
	return nil
}

// deepCopyMap creates a deep copy of a map[string]any.
func deepCopyMap(original map[string]any) map[string]any {
	copy := make(map[string]any, len(original))
	for k, v := range original {
		switch val := v.(type) {
		case map[string]any:
			copy[k] = deepCopyMap(val)
		case []any:
			copy[k] = deepCopySlice(val)
		default:
			copy[k] = val
		}
	}
	return copy
}

// deepCopySlice creates a deep copy of a []any.
func deepCopySlice(original []any) []any {
	copy := make([]any, len(original))
	for i, v := range original {
		switch val := v.(type) {
		case map[string]any:
			copy[i] = deepCopyMap(val)
		case []any:
			copy[i] = deepCopySlice(val)
		default:
			copy[i] = val
		}
	}
	return copy
}
