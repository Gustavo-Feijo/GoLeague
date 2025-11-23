package tierlistservice

import (
	"context"
	"encoding/json"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	TierlistMemoryCacheDuration = 15 * time.Minute
	TierlistRedisCacheDuration  = time.Hour
)

type TierlistRedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

// Tierlist service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type TierlistService struct {
	db                 *gorm.DB
	memCache           cache.MemCache
	redis              TierlistRedisClient
	TierlistRepository repositories.TierlistRepository
}

// TierlistServiceDeps is the dependency list for the tierlist service.
type TierlistServiceDeps struct {
	DB       *gorm.DB
	MemCache cache.MemCache
	Redis    TierlistRedisClient
}

// NewTierlistService creates a tierlist service.
func NewTierlistService(deps *TierlistServiceDeps) *TierlistService {
	return &TierlistService{
		db:                 deps.DB,
		memCache:           deps.MemCache,
		redis:              deps.Redis,
		TierlistRepository: repositories.NewTierlistRepository(deps.DB),
	}
}

// GetTierlist get the tierlist based on the filters.
func (ts *TierlistService) GetTierlist(filters *filters.TierlistFilter) ([]*dto.TierlistResult, error) {
	key := ts.getTierlistKey(filters)

	if mem := ts.getFromMemCache(key); mem != nil {
		return mem, nil
	}

	if redisData := ts.getFromRedis(key); redisData != nil {
		ts.memCache.Set(key, redisData, TierlistMemoryCacheDuration)
		return redisData, nil
	}

	// Get the data from the repository.
	results, err := ts.TierlistRepository.GetTierlist(filters)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []*dto.TierlistResult{}, nil
	}

	var dtoHelper dto.TierlistResult
	tierlistResultDTO := dtoHelper.FromRepositorySlice(results)

	ts.populateCaches(key, tierlistResultDTO)

	return tierlistResultDTO, nil
}

// getFromMemCache retrieves the data from the memory and returns it.
func (ts *TierlistService) getFromMemCache(key string) []*dto.TierlistResult {
	if memCachedData := ts.memCache.Get(key); memCachedData != nil {
		return memCachedData.([]*dto.TierlistResult)
	}
	return nil
}

// getFromRedis retrieves the data from the redis.
func (ts *TierlistService) getFromRedis(key string) []*dto.TierlistResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	// Try to get it on redis.
	redisCached, err := ts.redis.Get(ctx, key)
	if err != nil || redisCached == "" {
		return nil
	}
	// Unmarshal the value to save it as binary on the cache.
	var fulltierlist []*dto.TierlistResult
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
func (ts *TierlistService) populateCaches(key string, data []*dto.TierlistResult) {
	ts.memCache.Set(key, data, TierlistMemoryCacheDuration)

	if j, err := json.Marshal(data); err == nil {
		ts.redis.Set(context.Background(), key, string(j), TierlistRedisCacheDuration)
	}
}
