package champion

import (
	"context"
	"goleague/api/cache"
	"goleague/api/filters"

	"gorm.io/gorm"
)

// ChampionService with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type ChampionService struct {
	db            *gorm.DB
	championCache cache.ChampionCache
	memCache      cache.MemCache
}

// ChampionServiceDeps is the dependency list for the champion service.
type ChampionServiceDeps struct {
	DB            *gorm.DB
	ChampionCache cache.ChampionCache
	MemCache      cache.MemCache
}

// NewChampionService creates a champion service.
func NewChampionService(deps *ChampionServiceDeps) *ChampionService {
	return &ChampionService{
		db:            deps.DB,
		championCache: deps.ChampionCache,
		memCache:      deps.MemCache,
	}
}

// Wrapper that just returns the cached champion.
func (cs *ChampionService) GetChampionData(ctx context.Context, filters *filters.GetChampionDataFilter) (map[string]any, error) {
	return cs.championCache.GetChampionCopy(ctx, filters.ChampionId)
}

// Wrapper that just returns all cached champions.
func (cs *ChampionService) GetAllChampions(ctx context.Context) ([]map[string]any, error) {
	return cs.championCache.GetAllChampions(ctx)
}
