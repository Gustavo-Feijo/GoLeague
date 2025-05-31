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
)

// Tierlist service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type TierlistService struct {
	TierlistRepository repositories.TierlistRepository
	grpcClient         *grpc.ClientConn
}

// Create a tierlist service.
func NewTierlistService(grpcClient *grpc.ClientConn) (*TierlistService, error) {
	// Create the repository.
	repo, err := repositories.NewTierlistRepository()
	if err != nil {
		return nil, errors.New("failed to start the tierlist repository")
	}

	return &TierlistService{
		grpcClient:         grpcClient,
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
	cacheChampion := cache.GetChampionCache()
	repo, _ := repositories.NewCacheRepository()
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
		championData, err := cacheChampion.GetChampionCopy(ctx, strconv.Itoa(entry.ChampionId), repo)
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
