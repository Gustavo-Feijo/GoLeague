package subregion

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	playerfetcher "goleague/fetcher/data/player"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	batchservice "goleague/fetcher/services/subregion/batch"
	leagueservice "goleague/fetcher/services/subregion/league"
	playerservice "goleague/fetcher/services/subregion/player"
	ratingservice "goleague/fetcher/services/subregion/rating"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"

	"gorm.io/gorm"
)

// SubRegionService coordinates data fetching and processing for a specific sub-region.
type SubRegionService struct {
	leagueService *leagueservice.LeagueService
	playerService *playerservice.PlayerService
	ratingService *ratingservice.RatingService
	batchService  *batchservice.BatchService
	logger        *logger.NewLogger
	subRegion     regions.SubRegion
}

// NewSubRegionService creates a new sub-region service.
func NewSubRegionService(db *gorm.DB, fetcher *data.SubFetcher, region regions.SubRegion) (*SubRegionService, error) {
	// Create the repositories.
	ratingRepository, err := repositories.NewRatingRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the rating repository")
	}

	playerRepository, err := repositories.NewPlayerRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the player repository")
	}

	// Create the logger.
	logger, err := logger.CreateLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to start the logger on sub region %s: %v", region, err)
	}

	// Create the services.
	leagueService := leagueservice.NewLeagueService(*fetcher, leagueservice.DefaultConfig())
	playerService := playerservice.NewPlayerService(*fetcher, playerRepository, region)
	ratingService := ratingservice.NewRatingService(ratingRepository, region)
	batchService := batchservice.NewBatchService(leagueService, playerService, ratingService, logger, region)

	// Return the new region service.
	return &SubRegionService{
		leagueService: leagueService,
		playerService: playerService,
		ratingService: ratingService,
		batchService:  batchService,
		logger:        logger,
		subRegion:     region,
	}, nil
}

// GetLogger returns the logger instance.
// Used for manual closing of the logs.
func (s *SubRegionService) GetLogger() *logger.NewLogger {
	return s.logger
}

// ProcessHighElo processes high elo leagues.
func (s *SubRegionService) ProcessHighElo(highElo string, queue string) error {
	// Get the data from the Riot API.
	entries, err := s.leagueService.GetHighEloLeagueEntries(highElo, queue)
	if err != nil {
		return err
	}

	// Process the batch.
	return s.batchService.ProcessBatchEntry(entries, queue)
}

// ProcessLeagueRank processes a specific tier and rank.
func (s *SubRegionService) ProcessLeagueRank(tier string, rank string, queue string) error {
	// Starting page.
	page := 1

	// Fetch pages until we get an empty result.
	for {
		// Get entries for the current page.
		entries, err := s.leagueService.GetLeagueEntries(tier, rank, queue, page)
		if err != nil {
			return err
		}

		// If no entry is found, break and go to the next rank.
		if len(entries) == 0 {
			break
		}

		// Process the batch.
		if err := s.batchService.ProcessBatchEntry(entries, queue); err != nil {
			return fmt.Errorf("error at processing page %d: %v", page, err)
		}

		// Increment page for next iteration.
		page++
	}

	return nil
}

// ProcessSummonerData is a wrapper for the player service call.
func (s *SubRegionService) ProcessSummonerData(playerAccount *playerfetcher.Account, onDemand bool) (*models.PlayerInfo, error) {
	return s.playerService.ProcessSummonerData(playerAccount, onDemand)
}
