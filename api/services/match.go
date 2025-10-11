package services

import (
	"fmt"
	"goleague/api/cache"
	"goleague/api/converters"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// MatchService with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type MatchService struct {
	championCache   cache.ChampionCache
	db              *gorm.DB
	memCache        cache.MemCache
	matchConverter  *converters.MatchConverter
	redis           TierlistRedisClient
	MatchRepository repositories.MatchRepository
}

// MatchServiceDeps is the dependency list for the tierlist service.
type MatchServiceDeps struct {
	DB            *gorm.DB
	ChampionCache cache.ChampionCache
	MemCache      cache.MemCache
	Redis         TierlistRedisClient
}

// NewTierlistService creates a tierlist service.
func NewMatchService(deps *MatchServiceDeps) *MatchService {
	return &MatchService{
		championCache:   deps.ChampionCache,
		db:              deps.DB,
		MatchRepository: repositories.NewMatchRepository(deps.DB),
		matchConverter:  &converters.MatchConverter{},
		memCache:        deps.MemCache,
		redis:           deps.Redis,
	}
}

func (ms *MatchService) GetFullMatchData(filter *filters.GetFullMatchDataFilter) error {
	match, err := ms.GetMatchByMatchId(filter.MatchId)
	if err != nil {
		return err
	}

	matchPreviews, err := ms.MatchRepository.GetMatchPreviewsByInternalId(match.ID)
	if err != nil {
		return err
	}
	formatedPreviews, err := ms.matchConverter.ConvertMultipleMatches(matchPreviews)
	if err != nil {
		return nil
	}
	fmt.Print(formatedPreviews)
	return nil
}

// GetMatchByMatchId is a simple wrapper for getting the match repository data.
func (ms *MatchService) GetMatchByMatchId(matchId string) (*models.MatchInfo, error) {
	return ms.MatchRepository.GetMatchByMatchId(matchId)
}
