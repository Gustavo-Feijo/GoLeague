package mainregion_queue

import (
	mainregion_processor "goleague/fetcher/data/processors/mainregion"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"log"
	"time"
)

// The configuration for the queues we will be executing.
type MainRegionQueueConfig struct {
	SleepDuration time.Duration
}

// Type for the sub region main process.
type MainRegionQueue struct {
	config     MainRegionQueueConfig
	processor  mainregion_processor.MainRegionProcessor
	subRegions []regions.SubRegion
}

// Return a default configuration for the sub region.
func CreateDefaultQueueConfig() *MainRegionQueueConfig {
	return &MainRegionQueueConfig{
		SleepDuration: 5 * time.Second,
	}
}

// Create the main region queue.
func CreateMainRegionQueue(region regions.MainRegion, rm *regions.RegionManager) (*MainRegionQueue, error) {
	// Create the fetcher.
	fetcher, err := rm.GetMainFetcher(region)
	if err != nil {
		log.Printf("Failed to get sub region fetcher for %v: %v", region, err)
		return nil, err
	}

	// Create the processor.
	processor, err := mainregion_processor.CreateMainRegionProcessor(fetcher, region)
	if err != nil {
		log.Printf("Failed to get sub region fetcher for %v: %v", region, err)
		return nil, err
	}

	// Get the sub regions to define which region must be fetched.
	subRegions, err := rm.GetSubRegions(region)
	if err != nil {
		log.Printf("Failed to run the main region processor for the region %v: %v", region, err)
		return nil, err
	}

	// Return the new region processor.
	return &MainRegionQueue{
		config:     *CreateDefaultQueueConfig(),
		processor:  *processor,
		subRegions: subRegions,
	}, nil
}

// Run the main region.
func (q *MainRegionQueue) Run() {
	// Infinite loop, must be always getting data.
	for {
		// Loop through each possible subRegion so we can get a evenly distributed amount of matches.
		for _, subRegion := range q.subRegions {
			player, err := q.processQueue(subRegion)
			if err != nil && player != nil {
				// Delay the player next fetch to avoid the queue getting stuck.
				if err := q.processor.PlayerService.SetDelayedLastFetch(player.ID); err != nil {
					log.Printf("Couldn't delay the next fetch for the player.")
				}
			}
		}
	}
}

// Process the queue.
func (q *MainRegionQueue) processQueue(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	player, err := q.processor.PlayerService.GetUnfetchedBySubRegions(subRegion)
	if err != nil {
		log.Printf("Couldn't get any unfetched player: %v", err)

		// Could be the first fetch, wait to the sub regions to start filling the database.
		time.Sleep(5 * time.Second)
		return nil, err
	}

	trueMatchList, err := q.processor.GetTrueMatchList(player)
	if err != nil {
		log.Printf("Couldn't get the true match list: %v", err)
		return player, err
	}

	// Loop through each match.
	for _, matchId := range trueMatchList {
		matchData, err := q.processor.GetMatchData(matchId)
		if err != nil {
			log.Printf("Couldn't get the match data for the match %s: %v", matchId, err)
			// Set the date of the last fetch of the player to 1 day in the future, so the queue doesn't stay stuck.
			return player, err
		}

		// Process the match data.
		if err := q.processor.ProcessMatchData(matchData, matchId, subRegion); err != nil {
			log.Printf("Couldn't process the data for the match %s: %v", matchId, err)
			return player, err
		}
	}

	// Set the last fetch.
	if err := q.processor.PlayerService.SetFetched(player.ID); err != nil {
		log.Printf("Couldn't set the last fetch date for the player with ID %d: %v", player.ID, err)
	}

	return player, nil
}
