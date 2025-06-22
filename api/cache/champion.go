package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/api/repositories"
	"goleague/pkg/redis"
	"time"

	"gorm.io/gorm"
)

// ChampionCache  uses the in-memory cache with small TTL to minimize Redis calls.
// Uses db as a fallback as last resource if Redis isn't available.
type ChampionCache struct {
	db       *gorm.DB
	memCache *MemCache
	redis    *redis.RedisClient
}

// NewChampionCache creates the instance of the champion cache.
func NewChampionCache(db *gorm.DB, redis *redis.RedisClient, memCache *MemCache) *ChampionCache {
	cc := &ChampionCache{
		memCache: memCache,
		db:       db,
		redis:    redis,
	}

	return cc
}

// GetChampionCopy returns a champion from the in memory cache, if not already in there, get from the redis.
// Returns a deep copy, so it's safe to change the returned value directly.
func (c *ChampionCache) GetChampionCopy(ctx context.Context, championId string, repo repositories.CacheRepository) (map[string]any, error) {
	// Cache key for memCache and Redis.
	cacheKey := fmt.Sprintf("ddragon:champion:%s", championId)

	// Try to get directly from memory.
	if champCache := c.memCache.Get(cacheKey); champCache != nil {
		if champEntry, ok := champCache.(map[string]any); ok {
			return deepCopyMap(champEntry), nil
		}
	}

	// Get from the redis if doesn't found.
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

	c.memCache.Set(cacheKey, champJson, time.Hour)
	return deepCopyMap(champJson), nil

}

// Initialize pre-loads the cache into memory for faster access without cold start.
func (c *ChampionCache) Initialize(ctx context.Context) error {
	cachePrefix := "ddragon:champion:"

	// Get all the keys by prefix.
	keys, err := c.redis.GetKeysByPrefix(ctx, cachePrefix)
	if err != nil {
		repo, err := repositories.NewCacheRepository(c.db)
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
			c.memCache.Set(champion.CacheKey, champJson, time.Hour)
		}
		return nil
	}

	// Loop through each redis key and store it on memory.
	for _, key := range keys {
		champRedis, err := c.redis.Get(ctx, key)
		if err != nil {
			continue
		}
		var champJson map[string]any
		err = json.Unmarshal([]byte(champRedis), &champJson)
		if err != nil {
			return fmt.Errorf("failed to unmarshal champion data: %w", err)
		}
		c.memCache.Set(key, champJson, time.Hour)
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
