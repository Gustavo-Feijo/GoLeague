package services

import (
	"goleague/api/cache"
	"goleague/api/converters"
	"goleague/api/dto"
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

// GetFullMatchData retrieves and parses all data for a given match.
func (ms *MatchService) GetFullMatchData(filter *filters.GetFullMatchDataFilter) (*dto.FullMatchData, error) {
	match, err := ms.GetMatchByMatchId(filter.MatchId)
	if err != nil {
		return nil, err
	}

	matchPreviews, err := ms.MatchRepository.GetMatchPreviewsByInternalId(match.ID)
	if err != nil {
		return nil, err
	}

	formattedPreview, err := ms.matchConverter.ConvertSingleMatch(matchPreviews)
	if err != nil {
		return nil, err
	}

	participantFrames, err := ms.MatchRepository.GetParticipantFramesByInternalId(match.ID)
	if err != nil {
		return nil, err
	}

	formattedParticipantFrames := ms.matchConverter.GroupParticipantFramesByParticipantId(participantFrames)

	fullMatch := &dto.FullMatchData{
		Metadata:             formattedPreview.Metadata,
		ParticipantsPreviews: formattedPreview.Data,
		ParticipantFrames:    formattedParticipantFrames,
	}

	return fullMatch, nil
}

// GetMatchByMatchId is a simple wrapper for getting the match repository data.
func (ms *MatchService) GetMatchByMatchId(matchId string) (*models.MatchInfo, error) {
	return ms.MatchRepository.GetMatchByMatchId(matchId)
}
