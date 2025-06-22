package services

import (
	"context"
	"errors"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/repositories"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Tierlist service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type TierlistService struct {
	championCache      *cache.ChampionCache
	db                 *gorm.DB
	grpcClient         *grpc.ClientConn
	TierlistRepository repositories.TierlistRepository
}

// TierlistServiceDeps is the dependency list for the tierlist service.
type TierlistServiceDeps struct {
	GrpcClient    *grpc.ClientConn
	DB            *gorm.DB
	ChampionCache *cache.ChampionCache
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
		TierlistRepository: repo,
	}, nil
}

// GetTierlist get the tierlist based on the filters.
func (ts *TierlistService) GetTierlist(filters map[string]any) ([]*dto.FullTierlist, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// We return a error if the cache failed, while still returning the data.
	cacheFailed := false
	for index, entry := range results {
		fullResult[index] = &dto.FullTierlist{
			BanCount:     entry.Bancount,
			Banrate:      entry.Banrate,
			PickCount:    entry.Pickcount,
			PickRate:     entry.Pickrate,
			TeamPosition: entry.TeamPosition,
			WinRate:      entry.Winrate,
		}

		// Get a copy of the champion on the cache.
		championData, err := ts.championCache.GetChampionCopy(ctx, strconv.Itoa(entry.ChampionId), repo)
		if err != nil {
			fullResult[index].Champion = map[string]any{"ID": strconv.Itoa(entry.ChampionId)}
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

	return fullResult, nil
}
