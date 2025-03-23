package subregion_queue

import (
	subregion_processor "goleague/fetcher/data/processors/subregion"
	"goleague/fetcher/regions"
	"log"
	"time"
)

// The configuration for the queues we will be executing.
type SubRegionQueueConfig struct {
	Ranks         []string
	HighElos      []string
	Queues        []string
	SleepDuration time.Duration
	Tiers         []string
}

// Type for the sub region main process.
type SubRegionQueue struct {
	config    SubRegionQueueConfig
	processor subregion_processor.SubRegionProcessor
}

// Return a default configuration for the sub region.
func CreateDefaultQueueConfig() *SubRegionQueueConfig {
	return &SubRegionQueueConfig{
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

// Create the sub region queue.
func CreateSubRegionQueue(region regions.SubRegion, rm *regions.RegionManager) (*SubRegionQueue, error) {
	// Create the fetcher.
	fetcher, err := rm.GetSubFetcher(region)
	if err != nil {
		log.Printf("Failed to get sub region fetcher for %v: %v", region, err)
		return nil, err
	}

	// Create the processor.
	processor, err := subregion_processor.CreateSubRegionProcessor(fetcher, region)
	if err != nil {
		log.Printf("Failed to get sub region fetcher for %v: %v", region, err)
		return nil, err
	}

	// Return the new region processor.
	return &SubRegionQueue{
		config:    *CreateDefaultQueueConfig(),
		processor: *processor,
	}, nil
}

// Run the sub region queue.
// Mainly responsible for getting the ratings for each player on the region.
func (q *SubRegionQueue) Run() {
	q.processQueues()

	// Sleep to wait new matches to happen.
	time.Sleep(q.config.SleepDuration)
}

// Process the high elo and the other leagues for each queue.
func (q *SubRegionQueue) processQueues() {
	for _, queue := range q.config.Queues {
		q.processHighElo(queue)
		q.processLeagues(queue)
	}
}

// Process the high elo league.
func (q *SubRegionQueue) processHighElo(queue string) {
	// Go through each high elo entry.
	for _, highElo := range q.config.HighElos {
		if err := q.processor.ProcessHighElo(highElo, queue); err != nil {
			log.Printf("Couldn't process the high elo %s for the queue %s on region %s: %v", highElo, queue, q.processor.SubRegion, err)
			continue
		}
	}
}

// Process each league and sub rank.
func (q *SubRegionQueue) processLeagues(queue string) {
	// Loop through each available tier.
	for _, tier := range q.config.Tiers {
		// Loop through each available rank.
		for _, rank := range q.config.Ranks {

			if err := q.processor.ProcessLeagueRank(tier, rank, queue); err != nil {
				log.Printf("Couldn't process the league %s - rank %s for the queue %s on region %s: %v", tier, rank, queue, q.processor.SubRegion, err)
				continue
			}
		}
	}
}
