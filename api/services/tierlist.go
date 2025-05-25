package services

import (
	"context"
	"errors"
	"goleague/api/cache"
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

// The fulltierlist we will return.
type FullTierlist struct {
	Champion     map[string]any // Will be a simplified version of champion.Champion, without spell data.
	BanCount     int
	Banrate      float64
	PickCount    int
	PickRate     float64
	TeamPosition string
	WinRate      float64
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
func (ts *TierlistService) GetTierlist(filters map[string]any) ([]*FullTierlist, error) {
	// Get the data from the repository.
	results, err := ts.TierlistRepository.GetTierlist(filters)
	if err != nil {
		return nil, err
	}

	// Create the array of  results.
	fullResult := make([]*FullTierlist, len(results))
	if len(results) == 0 {
		return fullResult, nil
	}

	// Get the champion cache instance.
	cacheChampion := cache.GetChampionCache()
	repo, _ := repositories.NewCacheRepository()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for index, entry := range results {
		// Get a copy of the champion on the cache.
		championData, err := cacheChampion.GetChampionCopy(ctx, strconv.Itoa(entry.ChampionId), repo)
		if err != nil {
			continue
		}
		// Remove the spells and passive from the copied map.
		delete(championData, "spells")
		delete(championData, "passive")

		fullResult[index] = &FullTierlist{
			Champion:     championData,
			BanCount:     entry.Bancount,
			Banrate:      entry.Banrate,
			PickCount:    entry.Pickcount,
			PickRate:     entry.Pickrate,
			TeamPosition: entry.TeamPosition,
			WinRate:      entry.Winrate,
		}
	}
	return fullResult, nil
}
