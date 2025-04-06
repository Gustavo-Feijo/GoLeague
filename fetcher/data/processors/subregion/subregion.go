package subregion_processor

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"
)

// Type for the default configuration.
type subRegionConfig struct {
	maxRetries int
}

// Type for the sub region main process.
type SubRegionProcessor struct {
	config        subRegionConfig
	fetcher       data.SubFetcher
	Logger        *logger.NewLogger
	PlayerService models.PlayerService
	RatingService models.RatingService
	SubRegion     regions.SubRegion
}

// Return a default configuration for the sub region.
func createSubRegionConfig() *subRegionConfig {
	return &subRegionConfig{
		maxRetries: 3,
	}
}

// Create the sub region processor.
func CreateSubRegionProcessor(fetcher *data.SubFetcher, region regions.SubRegion) (*SubRegionProcessor, error) {
	// Create the services.
	ratingService, err := models.CreateRatingService()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	PlayerService, err := models.CreatePlayerService()
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	// Create the logger used to save the logs in a bucket.
	logger, err := logger.CreateLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to start the logger on sub region %s: %v", region, err)
	}

	// Return the new region processor.
	return &SubRegionProcessor{
		config:        *createSubRegionConfig(),
		fetcher:       *fetcher,
		Logger:        logger,
		PlayerService: *PlayerService,
		RatingService: *ratingService,
		SubRegion:     region,
	}, nil
}

// Get a list of the already existing players as well as the entries mapped for the player puuid.
func (p *SubRegionProcessor) getExistingPlayer(
	entries []league_fetcher.LeagueEntry,
) (
	map[string]*models.PlayerInfo,
	map[string]league_fetcher.LeagueEntry,
	error,
) {
	// Get each puuid from the entry.
	puuids := make([]string, len(entries))
	entryByPuuid := make(map[string]league_fetcher.LeagueEntry)

	// Fill the map for faster lookup.
	for i, entry := range entries {
		puuids[i] = entry.Puuid
		entryByPuuid[entry.Puuid] = entry
	}

	// Fetch all the players from those puuids.
	existingPlayers, err := p.PlayerService.GetPlayersByPuuids(puuids)
	if err != nil {
		return nil, nil, err
	}

	return existingPlayers, entryByPuuid, nil
}

// Get the already existing player in the database.
// Process a given high elo league.
func (p *SubRegionProcessor) ProcessHighElo(highElo string, queue string) error {
	// Get the data on the Riot API.
	highRating, err := p.fetcher.League.GetHighEloLeagueEntries(highElo, queue)
	if err != nil {
		return err
	}

	// Array for the batch insert.
	entries := make([]league_fetcher.LeagueEntry, len(highRating.Entries))

	// Process each rating entry.
	// If more resources available, can be converted to use Goroutines.
	for i, entry := range highRating.Entries {
		// For high elo we don't have the tier inside the entries array, so we set manually.
		entry.Tier = &highRating.Tier
		entry.QueueType = &queue
		entries[i] = entry
	}

	return p.processBatchEntry(entries, queue)
}

// Process a given league and sub rank.
func (p *SubRegionProcessor) ProcessLeagueRank(tier string, rank string, queue string) error {
	// Starting page that will be fetched.
	page := 1

	// Infinite loop that will run until the return of rating is a empty array.
	for {
		var ratingEntries []league_fetcher.LeagueEntry
		var err error

		// Try to get the entries with retry.
		for attempt := 1; attempt < p.config.maxRetries; attempt += 1 {
			ratingEntries, err = p.fetcher.League.GetLeagueEntries(tier, rank, queue, page)
			if err == nil {
				break
			}
		}

		// The retries were not effective.
		if err != nil {
			return err
		}

		// If no entry is found, we just break and go to the next rank.
		if len(ratingEntries) == 0 {
			break
		}

		// Array for the batch insert.
		entries := make([]league_fetcher.LeagueEntry, len(ratingEntries))

		// Process each rating entry.
		// If more resources available, can be converted to use Goroutines.
		for i, entry := range ratingEntries {
			entry.QueueType = &queue
			entries[i] = entry
		}

		// If can't insert into the database, simply return the error without retrying.
		if err := p.processBatchEntry(entries, queue); err != nil {
			return fmt.Errorf("error at processing page %d: %v", page, err)
		}
		page += 1
	}
	return nil
}

// Process each player that needs to be  created.
func (p *SubRegionProcessor) processPlayers(
	entries []league_fetcher.LeagueEntry,
	existingPlayers map[string]*models.PlayerInfo,
) ([]*models.PlayerInfo, error) {
	var playersToCreate []*models.PlayerInfo

	// Loop through each entry and verify if the player exists.
	// If doesn't, add to the current batch.
	for _, entry := range entries {

		_, exists := existingPlayers[entry.Puuid]
		// The player doesn't exist.
		if !exists {
			playersToCreate = append(playersToCreate, &models.PlayerInfo{
				SummonerId: entry.SummonerId,
				Puuid:      entry.Puuid,
				Region:     string(p.SubRegion),
			})
		}

	}

	// Creates the list of players.
	if len(playersToCreate) > 0 {
		if err := p.PlayerService.CreatePlayersBatch(playersToCreate); err != nil {
			return nil, fmt.Errorf("error inserting %v new players: %v", len(playersToCreate), err)
		}

		// Add newly created players to the existing players map
		for _, player := range playersToCreate {
			existingPlayers[player.Puuid] = player
		}

	}

	return playersToCreate, nil
}

// Process each rating entry that must be created.
func (p *SubRegionProcessor) processRatings(
	existingPlayers map[string]*models.PlayerInfo,
	entryByPuuid map[string]league_fetcher.LeagueEntry,
	lastRatings map[uint]*models.RatingEntry,
	queue string,
) (
	[]models.RatingEntry,
	error,
) {
	var ratingsToCreate []models.RatingEntry

	for _, player := range existingPlayers {
		// Get the corresponding entry for the player, as well as the last rating.
		entry := entryByPuuid[player.Puuid]
		lastRating, exists := lastRatings[player.ID]

		// If the last rating doesn't exist or it changed, then c reate a new rating.
		if !exists || p.RatingService.RatingNeedsUpdate(lastRating, entry) {
			newRating := models.RatingEntry{
				PlayerId:     player.ID,
				Region:       p.SubRegion,
				Queue:        queue,
				LeaguePoints: entry.LeaguePoints,
				Wins:         entry.Wins,
				Losses:       entry.Losses,
			}

			// Handle Tier and Rank if they are not nil.
			if entry.Tier != nil {
				newRating.Tier = *entry.Tier
			}

			if entry.Rank != nil {
				newRating.Rank = *entry.Rank
			} else {
				// If it's high elo, it will be nil, just set the ranking as I.
				newRating.Rank = "I"
			}

			ratingsToCreate = append(ratingsToCreate, newRating)
		}
	}

	// Create the ratings.
	if len(ratingsToCreate) > 0 {
		if err := p.RatingService.CreateBatchRating(ratingsToCreate); err != nil {
			return nil, fmt.Errorf("error creating rating entries: %v", err)
		}
	}

	return ratingsToCreate, nil
}

// Process a batch of league entries.
func (p *SubRegionProcessor) processBatchEntry(entries []league_fetcher.LeagueEntry, queue string) error {
	// If empty just return.
	if len(entries) == 0 {
		return nil
	}

	// Get the existing players on the entries.
	// Also get the map of the entry by the player puuid.
	existingPlayers, entryByPuuid, err := p.getExistingPlayer(entries)
	if err != nil {
		return fmt.Errorf("couldn't get the existing players by puuid: %v", err)
	}
	// Create the necessary players.
	playersToCreate, err := p.processPlayers(entries, existingPlayers)
	if err != nil {
		return fmt.Errorf("couldn't create t he players from the entries: %v", err)
	}

	if len(playersToCreate) > 0 {
		p.Logger.Infof("Created: %d Players - Region: %v", len(playersToCreate), p.SubRegion)
	}

	// Get the player IDs from the inserted results.
	playerIDs := make([]uint, 0, len(existingPlayers))

	// Loop through each player and set the value as the database model.
	for _, player := range existingPlayers {
		playerIDs = append(playerIDs, player.ID)
	}

	// Get the rating for each player.
	lastRatings, err := p.RatingService.GetLastRatingEntryByPlayerIdsAndQueue(playerIDs, queue)
	if err != nil {
		return fmt.Errorf("error fetching last ratings: %v", err)
	}

	createdRatings, err := p.processRatings(existingPlayers, entryByPuuid, lastRatings, queue)
	if err != nil {
		return fmt.Errorf("error processing the ratings: %v", err)
	}

	// Verify if created any rating.
	if len(createdRatings) > 0 {
		// Extract the tier for printing.
		tier := createdRatings[0].Tier
		rank := createdRatings[0].Rank

		p.Logger.Infof("Created: %d - Queue: %s - Region: %v - Tier: %v - Rank: %v",
			len(createdRatings), queue, p.SubRegion, tier, rank)
	}

	return nil
}
