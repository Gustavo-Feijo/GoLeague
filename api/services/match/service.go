package matchservice

import (
	"goleague/api/cache"
	"goleague/api/converters"
	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// MatchService with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type MatchService struct {
	db              *gorm.DB
	memCache        cache.MemCache
	matchConverter  converters.MatchConverter
	MatchRepository matchrepo.MatchRepository
}

// MatchServiceDeps is the dependency list for the tierlist service.
type MatchServiceDeps struct {
	DB       *gorm.DB
	MemCache cache.MemCache
}

// NewTierlistService creates a tierlist service.
func NewMatchService(deps *MatchServiceDeps) *MatchService {
	return &MatchService{
		db:              deps.DB,
		MatchRepository: matchrepo.NewMatchRepository(deps.DB),
		memCache:        deps.MemCache,
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

	rawEvents, err := ms.MatchRepository.GetAllEvents(match.ID)
	if err != nil {
		return nil, err
	}

	events := ms.matchConverter.ConvertEvents(rawEvents)

	fullMatch := &dto.FullMatchData{
		Metadata:             formattedPreview.Metadata,
		ParticipantsPreviews: formattedPreview.Data,
		ParticipantFrames:    formattedParticipantFrames,
		Events:               events,
	}

	return fullMatch, nil
}

// GetMatchByMatchId is a simple wrapper for getting the match repository data.
func (ms *MatchService) GetMatchByMatchId(matchId string) (*models.MatchInfo, error) {
	return ms.MatchRepository.GetMatchByMatchId(matchId)
}
