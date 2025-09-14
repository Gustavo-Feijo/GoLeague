package services

import (
	"context"
	"encoding/json"
	"errors"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TierlistRedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

// Tierlist service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type TierlistService struct {
	championCache      cache.ChampionCache
	db                 *gorm.DB
	memCache           cache.MemCache
	redis              TierlistRedisClient
	TierlistRepository repositories.TierlistRepository
}

// TierlistServiceDeps is the dependency list for the tierlist service.
type TierlistServiceDeps struct {
	DB            *gorm.DB
	ChampionCache cache.ChampionCache
	MemCache      cache.MemCache
	Redis         TierlistRedisClient
}

// NewTierlistService creates a tierlist service.
func NewTierlistService(deps *TierlistServiceDeps) *TierlistService {
	return &TierlistService{
		championCache:      deps.ChampionCache,
		db:                 deps.DB,
		memCache:           deps.MemCache,
		redis:              deps.Redis,
		TierlistRepository: repositories.NewTierlistRepository(deps.DB),
	}
}

// GetTierlist get the tierlist based on the filters.
func (ts *TierlistService) GetTierlist(filters *filters.TierlistFilter) ([]*dto.FullTierlist, error) {
	key := ts.getTierlistKey(filters)

	if mem := ts.getFromMemCache(key); mem != nil {
		return mem, nil
	}

	if redisData := ts.getFromRedis(key); redisData != nil {
		ts.memCache.Set(key, redisData, 15*time.Minute)
		return redisData, nil
	}

	// Get the data from the repository.
	results, err := ts.TierlistRepository.GetTierlist(filters)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []*dto.FullTierlist{}, nil
	}

	fullResult, cacheFailed := ts.buildFullTierlist(results)

	if !cacheFailed {
		ts.populateCaches(key, fullResult)
	}

	// Couldn't get all entries from the redis cache.
	// Return the error so it can revalidate the tierlist cache.
	if cacheFailed {
		return fullResult, errors.New("cache failed")
	}

	return fullResult, nil
}

// buildFullTierlist clean unecessary data from the champion data and build the full tierlist result.
func (ts *TierlistService) buildFullTierlist(results []*dto.TierlistResult) ([]*dto.FullTierlist, bool) {
	// Create the array of  results.
	fullResult := make([]*dto.FullTierlist, len(results))
	cacheFailed := false

	// Get the champion cache instance.
	championCtx, championCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer championCancel()

	for index, entry := range results {
		fullResult[index] = &dto.FullTierlist{
			BanCount:     entry.BanCount,
			Banrate:      entry.BanRate,
			PickCount:    entry.PickCount,
			PickRate:     entry.PickRate,
			TeamPosition: entry.TeamPosition,
			WinRate:      entry.WinRate,
		}

		// Get a copy of the champion on the cache.
		championData, err := ts.championCache.GetChampionCopy(championCtx, strconv.Itoa(entry.ChampionId))
		if err != nil {
			fullResult[index].Champion = map[string]any{"id": strconv.Itoa(entry.ChampionId)}
			cacheFailed = true
			continue
		}
		// Remove the spells and passive from the copied map.
		delete(championData, "spells")
		delete(championData, "passive")

		fullResult[index].Champion = championData
	}

	return fullResult, cacheFailed
}

// getFromMemCache retrieves the data from the memory and returns it.
func (ts *TierlistService) getFromMemCache(key string) []*dto.FullTierlist {
	if memCachedData := ts.memCache.Get(key); memCachedData != nil {
		return memCachedData.([]*dto.FullTierlist)
	}
	return nil
}

// getFromRedis retrieves the data from the redis.
func (ts *TierlistService) getFromRedis(key string) []*dto.FullTierlist {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	// Try to get it on redis.
	redisCached, err := ts.redis.Get(ctx, key)
	if err != nil || redisCached == "" {
		return nil
	}
	// Unmarshal the value to save it as binary on the cache.
	var fulltierlist []*dto.FullTierlist
	if err := json.Unmarshal([]byte(redisCached), &fulltierlist); err != nil {
		return nil
	}

	return fulltierlist
}

// getTierListKey generates the cache key.
func (ts *TierlistService) getTierlistKey(filters *filters.TierlistFilter) string {
	var builder strings.Builder
	builder.WriteString("tierlist")

	if filters.Queue != 0 {
		builder.WriteString(":queue_" + strconv.Itoa(filters.Queue))
	}

	if filters.NumericTier != 0 {
		builder.WriteString(":tier_" + strconv.Itoa(filters.NumericTier))
	}

	if filters.GetTiersAbove {
		builder.WriteString(":with_higher_tiers")
	}

	return builder.String()
}

// populateCaches will set the mem cache and redis cache.
func (ts *TierlistService) populateCaches(key string, data []*dto.FullTierlist) {
	ts.memCache.Set(key, data, 15*time.Minute)

	if j, err := json.Marshal(data); err == nil {
		ts.redis.Set(context.Background(), key, string(j), time.Hour)
	}
}
