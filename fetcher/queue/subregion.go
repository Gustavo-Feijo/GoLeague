package queue

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
type SubRegionConfig struct {
	Divisions     []string
	HighElos      []string
	Queues        []string
	SleepDuration time.Duration
	Tiers         []string
}

// Type for the sub region main process.
type SubRegionProcessor struct {
	config        SubRegionConfig
	fetcher       data.SubFetcher
	playerService models.PlayerService
	ratingService models.RatingService
	subRegion     regions.SubRegion
}

// Return a default configuration for the sub region.
func CreateDefaultQueueConfig() *SubRegionConfig {
	return &SubRegionConfig{
		Divisions:     []string{"I", "II", "III", "IV"},
		HighElos:      []string{"challenger", "grandmaster", "master"},
		Queues:        []string{"RANKED_SOLO_5x5", "RANKED_FLEX_SR"},
		SleepDuration: 10 * time.Minute,
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
func CreateSubRegionProcessor(fetcher *data.SubFetcher, region regions.SubRegion) (*SubRegionProcessor, error) {
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
	return &SubRegionProcessor{
		config:        *CreateDefaultQueueConfig(),
		fetcher:       *fetcher,
		playerService: *playerService,
		ratingService: *ratingService,
		subRegion:     region,
	}, nil
}

// Run the sub region queue.
// Mainly responsible for getting the ratings for each player on the region.
func runSubRegionQueue(region regions.SubRegion, rm *regions.RegionManager) {
	fetcher, err := rm.GetSubFetcher(region)
	if err != nil {
		log.Printf("Failed to get main region fetcher for %v: %v", region, err)
		return
	}

	// Start the processor for the sub region.
	processor, err := CreateSubRegionProcessor(fetcher, region)
	if err != nil {
		log.Printf("Failed to start the sub region processor for the region %v: %v", region, err)
		return
	}

	processor.processQueues()

	// Sleep to wait new matches to happen.
	time.Sleep(processor.config.SleepDuration)
}

// Process the high elo and the other leagues for each queue.
func (p *SubRegionProcessor) processQueues() {
	for _, queue := range p.config.Queues {
		p.processHighElo(queue)
		p.processLeagues(queue)
	}
}

// Process the high elo league.
func (p *SubRegionProcessor) processHighElo(queue string) {
	// Go through each high elo entry.
	for _, highElo := range p.config.HighElos {
		// Get the data on the Riot API.
		highRating, err := p.fetcher.League.GetHighEloLeagueEntries(highElo, queue)
		if err != nil {
			log.Printf("Couldn't get the high elo league %v: %v", highElo, err)
			continue
		}

		// Process each rating entry.
		// If more resources available, can be converted to use Goroutines.
		for _, entry := range highRating.Entries {
			// For high elo we don't have the tier inside the entries array, so we set manually.
			entry.Tier = &highRating.Tier

			p.processEntry(entry, queue)
		}
	}
}

// Process each league and sub division.
func (p *SubRegionProcessor) processLeagues(queue string) {
	// Loop through each available tier.
	for _, tier := range p.config.Tiers {
		// Loop through each available division.
		for _, division := range p.config.Divisions {
			// Starting page that will be fetched.
			page := 1

			// Infinite loop that will run until the return of rating is a empty array.
			for {
				rating, err := p.fetcher.League.GetLeagueEntries(tier, division, queue, page)
				if err != nil {
					log.Printf("Couldn't get the tier %v on the division %v: %v", tier, division, err)
					break
				}

				// If no entry is found, we just break and go to the next division.
				if len(rating) == 0 {
					break
				}

				// Process each rating entry.
				// If more resources available, can be converted to use Goroutines.
				for _, entry := range rating {
					// Just process the entry.
					p.processEntry(entry, queue)
				}

				page += 1
			}
		}
	}
}

// Process a single league entry.
func (p *SubRegionProcessor) processEntry(entry league_fetcher.LeagueEntry, queue string) {
	player, err := p.playerService.GetPlayerByPuuid(entry.Puuid)
	if err != nil {
		log.Printf("Couldn't get the player with PUUID: %v", entry.Puuid)
		return
	}

	// The player doesn't exist.
	if player == nil {
		// Reassign the player to the newly created one.
		player, err = p.playerService.CreatePlayerFromRating(entry, p.subRegion)
		if err != nil {
			log.Printf("Couldn't create the player with PUUID %v on the region %v: %v", entry.Puuid, p.subRegion, err)
			return
		}

		log.Printf("Created player with id %v and PUUID %v in region %v", player.ID, player.Puuid, p.subRegion)
	}

	p.updatePlayerRating(player, entry, queue)
}

// Update the player rating if needed.
func (p *SubRegionProcessor) updatePlayerRating(player *models.PlayerInfo, entry league_fetcher.LeagueEntry, queue string) {
	// Get the last rating of this player.
	// Used to verify if the player has played a match recently.
	lastRating, err := p.ratingService.GetLastRatingEntryByPlayerIdAndQueue(player.ID, queue)
	if err != nil {
		log.Printf("Couldn't get the last rating for the player %v: %v", player.ID, err)
		return
	}

	// Finally create the rating entry.
	newRating, err := p.ratingService.CreateRatingEntry(entry, player.ID, p.subRegion, queue, lastRating)
	if err != nil {
		log.Printf("Couldn't insert the new rating entry for the player %v: %v", player.ID, err)
		return
	}

	// Verify if something changed.
	if newRating != nil {
		log.Printf("Created rating entry for the player %v on the region %v", player.ID, p.subRegion)
	}
}
