package services

import (
	"goleague/api/cache"
	"goleague/api/repositories"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// MatchService with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type MatchService struct {
	championCache   cache.ChampionCache
	db              *gorm.DB
	memCache        cache.MemCache
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
		memCache:        deps.MemCache,
		redis:           deps.Redis,
	}
}

// GetMatchByMatchId is a simple wrapper for getting the match repository data.
func (ms *MatchService) GetMatchByMatchId(matchId string) (*models.MatchInfo, error) {
	return ms.MatchRepository.GetMatchByMatchId(matchId)
}
