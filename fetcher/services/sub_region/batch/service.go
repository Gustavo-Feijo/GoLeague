package batchservice

import (
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	leagueservice "goleague/fetcher/services/sub_region/league"
	playerservice "goleague/fetcher/services/sub_region/player"
	ratingservice "goleague/fetcher/services/sub_region/rating"
	"goleague/pkg/logger"
)

// Handles batch processing of league entries.
type BatchService struct {
	leagueService *leagueservice.LeagueService
	playerService *playerservice.PlayerService
	ratingService *ratingservice.RatingService
	logger        *logger.NewLogger
	subRegion     regions.SubRegion
}

// Creates a new batch service
func NewBatchService(
	leagueService *leagueservice.LeagueService,
	playerService *playerservice.PlayerService,
	ratingService *ratingservice.RatingService,
	logger *logger.NewLogger,
	subRegion regions.SubRegion,
) *BatchService {
	return &BatchService{
		leagueService: leagueService,
		playerService: playerService,
		ratingService: ratingService,
		logger:        logger,
		subRegion:     subRegion,
	}
}

// Processes a batch of league entries
func (s *BatchService) ProcessBatchEntry(entries []league_fetcher.LeagueEntry, queue string) error {
	// If empty just return
	if len(entries) == 0 {
		return nil
	}

	// Extract PUUIDs from entries.
	puuids, entryByPuuid := s.leagueService.ExtractPuuidsFromEntries(entries)

	// Get existing players.
	existingPlayers, err := s.playerService.GetPlayersByPuuids(puuids)
	if err != nil {
		return fmt.Errorf("couldn't get the existing players by puuid: %v", err)
	}

	// Process players (create missing ones)
	playersToCreate, err := s.playerService.ProcessPlayersFromEntries(entries, existingPlayers)
	if err != nil {
		return fmt.Errorf("couldn't create the players from the entries: %v", err)
	}

	// Log player creation if any.
	if len(playersToCreate) > 0 {
		s.logger.Infof("Created: %d Players - Region: %v", len(playersToCreate), s.subRegion)
	}

	playerIDs := s.playerService.GetPlayerIDsFromMap(existingPlayers)

	// Get last ratings for these players.
	lastRatings, err := s.ratingService.GetLastRatingsByPlayerIdsAndQueue(playerIDs, queue)
	if err != nil {
		return fmt.Errorf("error fetching last ratings: %v", err)
	}

	createdRatings, err := s.ratingService.ProcessRatings(existingPlayers, entryByPuuid, lastRatings, queue)
	if err != nil {
		return fmt.Errorf("error processing the ratings: %v", err)
	}

	// Log rating creation if any.
	if len(createdRatings) > 0 {
		// Extract the tier for printing
		tier := createdRatings[0].Tier
		rank := createdRatings[0].Rank

		s.logger.Infof("Created: %d - Queue: %s - Region: %v - Tier: %v - Rank: %v",
			len(createdRatings), queue, s.subRegion, tier, rank)
	}

	return nil
}
