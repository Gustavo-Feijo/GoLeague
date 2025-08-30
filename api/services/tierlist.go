package services

import (
	"context"
	"encoding/json"
	"errors"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/pkg/redis"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Tierlist service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type TierlistService struct {
	championCache      *cache.ChampionCache
	db                 *gorm.DB
	grpcClient         *grpc.ClientConn
	memCache           *cache.MemCache
	redis              *redis.RedisClient
	TierlistRepository repositories.TierlistRepository
}

// TierlistServiceDeps is the dependency list for the tierlist service.
type TierlistServiceDeps struct {
	GrpcClient    *grpc.ClientConn
	DB            *gorm.DB
	ChampionCache *cache.ChampionCache
	MemCache      *cache.MemCache
	Redis         *redis.RedisClient
}

// NewTierlistService creates a tierlist service.
func NewTierlistService(deps *TierlistServiceDeps) (*TierlistService, error) {
	// Create the repository.
	repo, err := repositories.NewTierlistRepository(deps.DB)
	if err != nil {
		return nil, errors.New("failed to start the tierlist repository")
	}

	return &TierlistService{
		championCache:      deps.ChampionCache,
		db:                 deps.DB,
		grpcClient:         deps.GrpcClient,
		memCache:           deps.MemCache,
		redis:              deps.Redis,
		TierlistRepository: repo,
	}, nil
}

// GetTierlist get the tierlist based on the filters.
func (ts *TierlistService) GetTierlist(filters *filters.TierlistFilter) ([]*dto.FullTierlist, error) {
	key := ts.getTierlistKey(filters)

	// Get a instance of the memory cache and retrieve the key.
	memCachedData := ts.memCache.Get(key)
	if memCachedData != nil {
		memCachedTierlist := memCachedData.([]*dto.FullTierlist)
		return memCachedTierlist, nil
	}

	// Create context for fast  redis lookup.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	// Try to get it on redis.
	redisCached, err := ts.redis.Get(ctx, key)
	if err == nil {
		// Unmarshal the value to save it as binary on the cache.
		var fulltierlist []*dto.FullTierlist
		json.Unmarshal([]byte(redisCached), &fulltierlist)
		ts.memCache.Set(key, fulltierlist, 15*time.Minute)
		return fulltierlist, nil
	}
	// Get the data from the repository.
	results, err := ts.TierlistRepository.GetTierlist(filters)
	if err != nil {
		return nil, err
	}

	// Create the array of  results.
	fullResult := make([]*dto.FullTierlist, len(results))
	if len(results) == 0 {
		return fullResult, nil
	}

	// Get the champion cache instance.
	repo, _ := repositories.NewCacheRepository(ts.db)
	championCtx, championCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer championCancel()

	// We return a error if the cache failed, while still returning the data.
	cacheFailed := false
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
		championData, err := ts.championCache.GetChampionCopy(championCtx, strconv.Itoa(entry.ChampionId), repo)
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

	// Couldn't get all entries from the redis cache.
	// Return the error so it can revalidate the tierlist cache.
	if cacheFailed {
		return fullResult, errors.New("cache failed")
	}

	// Set the value in memory and redis.
	ts.memCache.Set(key, fullResult, 15*time.Minute)

	// Marshal it to set on Redis.
	j, err := json.Marshal(fullResult)
	if err == nil {
		ts.redis.Set(context.Background(), key, string(j), time.Hour)
	}

	return fullResult, nil
}

// getTierList generates the cache key.
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
