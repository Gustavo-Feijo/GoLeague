package subregion_queue

import (
	"errors"
	"goleague/fetcher/data"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"log"
	"time"
)

// The configuration for the queues we will be executing.
type subRegionConfig struct {
	Ranks         []string
	HighElos      []string
	Queues        []string
	SleepDuration time.Duration
	Tiers         []string
}

// Type for the sub region main process.
type subRegionProcessor struct {
	config        subRegionConfig
	fetcher       data.SubFetcher
	playerService models.PlayerService
	ratingService models.RatingService
	subRegion     regions.SubRegion
}

// Return a default configuration for the sub region.
func CreateDefaultQueueConfig() *subRegionConfig {
	return &subRegionConfig{
		Ranks:         []string{"I", "II", "III", "IV"},
		HighElos:      []string{"challenger", "grandmaster", "master"},
		Queues:        []string{"RANKED_SOLO_5x5", "RANKED_FLEX_SR"},
		SleepDuration: 60 * time.Minute,
		Tiers: []string{
			"DIAMOND",
			"EMERALD",
			"PLATINUM",
			"GOLD",
			"SILVER",
			"BRONZE",
			"IRON",
		},
	}
}

// Create the sub region processor.
func CreatesubRegionProcessor(fetcher *data.SubFetcher, region regions.SubRegion) (*subRegionProcessor, error) {
	// Create the services.
	ratingService, err := models.CreateRatingService()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	playerService, err := models.CreatePlayerService()
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	// Return the new region processor.
	return &subRegionProcessor{
		config:        *CreateDefaultQueueConfig(),
		fetcher:       *fetcher,
		playerService: *playerService,
		ratingService: *ratingService,
		subRegion:     region,
	}, nil
}

// Run the sub region queue.
// Mainly responsible for getting the ratings for each player on the region.
func RunSubRegionQueue(region regions.SubRegion, rm *regions.RegionManager) {
	fetcher, err := rm.GetSubFetcher(region)
	if err != nil {
		log.Printf("Failed to get main region fetcher for %v: %v", region, err)
		return
	}

	// Start the processor for the sub region.
	processor, err := CreatesubRegionProcessor(fetcher, region)
	if err != nil {
		log.Printf("Failed to start the sub region processor for the region %v: %v", region, err)
		return
	}

	processor.processQueues()

	// Sleep to wait new matches to happen.
	time.Sleep(processor.config.SleepDuration)
}

// Process the high elo and the other leagues for each queue.
func (p *subRegionProcessor) processQueues() {
	for _, queue := range p.config.Queues {
		p.processHighElo(queue)
		p.processLeagues(queue)
	}
}

// Process the high elo league.
func (p *subRegionProcessor) processHighElo(queue string) {
	// Go through each high elo entry.
	for _, highElo := range p.config.HighElos {
		// Get the data on the Riot API.
		highRating, err := p.fetcher.League.GetHighEloLeagueEntries(highElo, queue)
		if err != nil {
			log.Printf("Couldn't get the high elo league %v: %v", highElo, err)
			continue
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

		p.processBatchEntry(entries, queue)
	}
}

// Process each league and sub rank.
func (p *subRegionProcessor) processLeagues(queue string) {
	// Loop through each available tier.
	for _, tier := range p.config.Tiers {
		// Loop through each available rank.
		for _, rank := range p.config.Ranks {
			// Starting page that will be fetched.
			page := 1

			// Infinite loop that will run until the return of rating is a empty array.
			for {
				rating, err := p.fetcher.League.GetLeagueEntries(tier, rank, queue, page)
				if err != nil {
					log.Printf("Couldn't get the tier %v on the rank %v: %v", tier, rank, err)
					break
				}

				// If no entry is found, we just break and go to the next rank.
				if len(rating) == 0 {
					break
				}

				// Array for the batch insert.
				entries := make([]league_fetcher.LeagueEntry, len(rating))

				// Process each rating entry.
				// If more resources available, can be converted to use Goroutines.
				for i, entry := range rating {
					entry.QueueType = &queue
					entries[i] = entry
				}
				p.processBatchEntry(entries, queue)
				page += 1
			}
		}
	}
}

// Process a single league entry.
func (p *subRegionProcessor) processBatchEntry(entries []league_fetcher.LeagueEntry, queue string) {
	// If empty just return.
	if len(entries) == 0 {
		return
	}

	// Get each puuid from the entry.
	puuids := make([]string, len(entries))
	entryByPuuid := make(map[string]league_fetcher.LeagueEntry)

	// Fill the map for faster lookup.
	for i, entry := range entries {
		puuids[i] = entry.Puuid
		entryByPuuid[entry.Puuid] = entry
	}

	// Fetch all the players from those puuids.
	existingPlayers, err := p.playerService.GetPlayersByPuuids(puuids)
	if err != nil {
		log.Printf("Error fetching the puuids on the entries: %v", err)
		return
	}

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
				Region:     string(p.subRegion),
			})
		}

	}

	// Creates the list of players.
	if len(playersToCreate) > 0 {
		if err := p.playerService.CreatePlayersBatch(playersToCreate); err != nil {
			log.Printf("Error inserting %v new players: %v", len(playersToCreate), err)
			return
		}

		// Add newly created players to the existing players map
		for _, player := range playersToCreate {
			existingPlayers[player.Puuid] = player
		}

		log.Printf("Created: %d Players - Region: %v", len(playersToCreate), p.subRegion)
	}

	// Get the player IDs from the inserted results.
	playerIDs := make([]uint, 0, len(existingPlayers))
	playerByID := make(map[uint]*models.PlayerInfo)

	// Loop through each player and set the value as the database model.
	for _, player := range existingPlayers {
		playerIDs = append(playerIDs, player.ID)
		playerByID[player.ID] = player
	}

	// Get the rating for each player.
	lastRatings, err := p.ratingService.GetLastRatingEntryByPlayerIdsAndQueue(playerIDs, queue)
	if err != nil {
		log.Printf("Error fetching last ratings: %v", err)
		return
	}

	var ratingsToCreate []models.RatingEntry

	for _, player := range existingPlayers {
		// Get the corresponding entry for the player, as well as the last rating.
		entry := entryByPuuid[player.Puuid]
		lastRating, exists := lastRatings[player.ID]

		// If the last rating doesn't exist or it changed, then c reate a new rating.
		if !exists || p.ratingService.RatingNeedsUpdate(lastRating, entry) {
			newRating := models.RatingEntry{
				PlayerId:     player.ID,
				Region:       p.subRegion,
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
		if err := p.ratingService.CreateBatchRating(ratingsToCreate); err != nil {
			log.Printf("Error creating rating entries: %v", err)
			return
		}

		// Extract the tier for printing.
		tier := ratingsToCreate[0].Tier
		rank := ratingsToCreate[0].Rank

		log.Printf("Created: %d - Queue: %s - Region: %v - Tier: %v - Rank: %v",
			len(ratingsToCreate), queue, p.subRegion, tier, rank)
	}
}
