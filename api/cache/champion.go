package cache

import (
	"context"
	"encoding/json"
	"fmt"
	cacherepo "goleague/api/repositories/cache"
	champmodel "goleague/pkg/models/champion"
	"goleague/pkg/redis"
	"log"
	"time"

	"gorm.io/gorm"
)

// Default cache duration for the champion keys.
const (
	cacheDuration             = time.Hour
	failedParsingChampionData = "failed to unmarshal champion data: %v"
)

type ChampionCache interface {
	GetAllChampions(ctx context.Context) ([]*champmodel.Champion, error)
	GetChampionCopy(ctx context.Context, championId string) (*champmodel.Champion, error)
	Initialize(ctx context.Context) error
}

type ChampionCacheRedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	GetKeysByPrefix(ctx context.Context, prefix string) ([]string, error)
}

// championCache  uses the in-memory cache with small TTL to minimize Redis calls.
// Uses db as a fallback as last resource if Redis isn't available.
type championCache struct {
	memCache        MemCache[*champmodel.Champion]
	redis           ChampionCacheRedisClient
	cacheRepository cacherepo.CacheRepository
}

// NewchampionCache creates the instance of the champion cache.
func NewChampionCache(db *gorm.DB, redis *redis.RedisClient, memCache MemCache[*champmodel.Champion]) ChampionCache {
	cc := &championCache{
		memCache:        memCache,
		redis:           redis,
		cacheRepository: cacherepo.NewCacheRepository(db),
	}

	return cc
}

// GetChampionCopy returns a champion from the in memory cache, if not already in there, get from the redis.
// Returns a deep copy, so it's safe to change the returned value directly.
func (c *championCache) GetChampionCopy(ctx context.Context, championId string) (*champmodel.Champion, error) {
	// Cache key for memCache and Redis.
	cacheKey := fmt.Sprintf("ddragon:champion:%s", championId)

	// Try to get directly from memory.
	if champCache := c.memCache.Get(cacheKey); champCache != nil {
		return champCache, nil
	}

	// Get from the redis if doesn't found.
	champRedis, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		// Get from the database fallback in that case.
		// It will be way slower, but will save in memory for the next requests.
		champRedis, err = c.cacheRepository.GetKey(cacheKey)
		if err != nil {
			// Everything went wrong.
			return nil, fmt.Errorf("error getting from the database fallback: %w", err)
		}
	}

	// Unmarshal as a generic map.
	// It will be of type champmodel.Champion.
	var champJson *champmodel.Champion
	err = json.Unmarshal([]byte(champRedis), &champJson)
	if err != nil {
		return nil, fmt.Errorf(failedParsingChampionData, err)
	}

	c.memCache.Set(cacheKey, champJson, cacheDuration)
	return champJson, nil

}

// Initialize pre-loads the cache into memory for faster access without cold start.
func (c *championCache) Initialize(ctx context.Context) error {
	cachePrefix := "ddragon:champion:"

	// Get all the keys by prefix.
	keys, err := c.redis.GetKeysByPrefix(ctx, cachePrefix)
	if err != nil {
		log.Printf("Failed pre-loading Redis champion keys: %v", err)

		// Get all champions by the prefix and save in memory.
		champions, _ := c.cacheRepository.GetByPrefix(cachePrefix)
		for _, champion := range champions {
			var champJson *champmodel.Champion
			err := json.Unmarshal([]byte(champion.CacheValue), &champJson)
			if err != nil {
				return err
			}
			c.memCache.Set(champion.CacheKey, champJson, cacheDuration)
		}
		return nil
	}

	// Loop through each redis key and store it on memory.
	for _, key := range keys {
		champRedis, err := c.redis.Get(ctx, key)
		if err != nil {
			continue
		}
		var champJson *champmodel.Champion
		err = json.Unmarshal([]byte(champRedis), &champJson)
		if err != nil {
			return fmt.Errorf(failedParsingChampionData, err)
		}
		c.memCache.Set(key, champJson, cacheDuration)
	}
	return nil
}

// GetAllChampions returns all champions that are cached.
// Don't get from in-memory cache due to TTL, only redis or repo.
func (c *championCache) GetAllChampions(ctx context.Context) ([]*champmodel.Champion, error) {
	cachePrefix := "ddragon:champion:"
	championsList := make([]*champmodel.Champion, 0)

	// Try to get all keys from Redis.
	keys, err := c.redis.GetKeysByPrefix(ctx, cachePrefix)
	if err != nil || len(keys) == 0 {
		log.Printf("Failed getting all champions from Redis: %v", err)
		// Fallback: load champions from persistent cache.
		champions, dbErr := c.cacheRepository.GetByPrefix(cachePrefix)
		if dbErr != nil {
			return nil, fmt.Errorf("failed to load champions from fallback cache: %w", dbErr)
		}

		for _, champion := range champions {
			var champJson *champmodel.Champion
			if err := json.Unmarshal([]byte(champion.CacheValue), &champJson); err != nil {
				log.Printf(failedParsingChampionData, err)
				continue
			}

			championsList = append(championsList, champJson)
		}

		return championsList, nil
	}

	// Redis has the keys
	for _, key := range keys {
		champRedis, err := c.redis.Get(ctx, key)
		if err != nil {
			log.Printf("Failed getting Redis key %s: %v", key, err)
			continue
		}

		var champJson *champmodel.Champion
		if err := json.Unmarshal([]byte(champRedis), &champJson); err != nil {
			return nil, fmt.Errorf(failedParsingChampionData, err)
		}

		championsList = append(championsList, champJson)
	}

	return championsList, nil
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
