package playerservice

import (
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	queuevalues "goleague/pkg/riotvalues/queue"
	"log"
)

// PlayerService is a separated service for player operations.
type PlayerService struct {
	MatchRepository  repositories.MatchRepository
	PlayerRepository repositories.PlayerRepository
	RatingRepository repositories.RatingRepository
}

// NewPlayerService creates a new player service.
func NewPlayerService(
	matchRepo repositories.MatchRepository,
	playerRepo repositories.PlayerRepository,
	ratingRepo repositories.RatingRepository,
) *PlayerService {
	return &PlayerService{
		MatchRepository:  matchRepo,
		PlayerRepository: playerRepo,
		RatingRepository: ratingRepo,
	}
}

// ProcessPlayersFromMatch process each player from a given match.
// Upserts the players, only updating the data if the match data is newer.
func (p *PlayerService) ProcessPlayersFromMatch(
	participants []matchfetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
	region regions.SubRegion,
) ([]*models.PlayerInfo, map[string]matchfetcher.MatchPlayer, error) {
	// Variables for batching or search.
	var playersToUpsert []*models.PlayerInfo
	participantByPuuid := make(map[string]matchfetcher.MatchPlayer)

	for _, participant := range participants {
		// Create a player to be inserted.
		player := &models.PlayerInfo{
			ProfileIcon:    participant.ProfileIcon,
			Puuid:          participant.Puuid,
			RiotIdGameName: participant.RiotIdGameName,
			RiotIdTagline:  participant.RiotIdTagline,
			SummonerId:     participant.SummonerId,
			SummonerLevel:  participant.SummonerLevel,
			Region:         region,
			UpdatedAt:      matchInfo.MatchStart,
		}

		participantByPuuid[player.Puuid] = participant
		playersToUpsert = append(playersToUpsert, player)
	}

	// Create/update the players.
	if err := p.PlayerRepository.UpsertPlayerBatch(playersToUpsert); err != nil {
		log.Printf("Couldn't create/update the players for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, err
	}

	var playerIds []uint

	for _, player := range playersToUpsert {
		playerIds = append(playerIds, player.ID)
	}

	// Extract the queue value to get from the rating entries.
	// Must be ranked solo/duo or flex.
	queue, exists := queuevalues.RankedQueueValue[matchInfo.QueueId]
	if exists {
		avgRating := p.RatingRepository.GetAverageRatingOnMatchByPlayerId(playerIds, matchInfo.ID, matchInfo.MatchStart, queue)
		if err := p.MatchRepository.SetAverageRating(matchInfo.ID, avgRating); err != nil {
			log.Printf("Couldn't set average rating for the match %s: %v", matchInfo.MatchId, err)
		}
	}

	return playersToUpsert, participantByPuuid, nil
}
