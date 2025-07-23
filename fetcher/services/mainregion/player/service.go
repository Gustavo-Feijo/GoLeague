package playerservice

import (
	"errors"
	"fmt"
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	queuevalues "goleague/pkg/riotvalues/queue"
	"log"
	"sync"
)

// PlayerService is a separated service for player operations.
type PlayerService struct {
	MatchRepository  repositories.MatchRepository
	PlayerRepository repositories.PlayerRepository
	RatingRepository repositories.RatingRepository
	upsertMu         sync.Mutex
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

// GetPlayerByNameTagRegion get the player data from the database based on the provided conditions.
func (p *PlayerService) GetPlayerByNameTagRegion(
	gameName string,
	gameTag string,
	region string,
) (*models.PlayerInfo, error) {
	player, err := p.PlayerRepository.GetPlayerByNameTagRegion(gameName, gameTag, region)
	if err != nil {
		return nil, fmt.Errorf("player not found: %v", err)
	}

	if player == nil {
		return nil, errors.New("player not found")
	}

	return player, nil
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
			SummonerLevel:  participant.SummonerLevel,
			Region:         region,
			UpdatedAt:      matchInfo.MatchStart,
		}

		participantByPuuid[player.Puuid] = participant
		playersToUpsert = append(playersToUpsert, player)
	}

	// Create/update the players.
	// Get the mutex for the player service to avoid deadlocks when using goroutines.
	// Need to be in the service to not slow other regions.
	p.upsertMu.Lock()
	if err := p.PlayerRepository.UpsertPlayerBatch(playersToUpsert); err != nil {
		log.Printf("Couldn't create/update the players for the match %s: %v", matchInfo.MatchId, err)
		p.upsertMu.Unlock()
		return nil, nil, err
	}
	p.upsertMu.Unlock()

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
