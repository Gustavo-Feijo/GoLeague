package subregionqueue

import (
	"fmt"
	regionmanager "goleague/fetcher/regionmanager"
	"goleague/fetcher/regions"
	subregionservice "goleague/fetcher/services/subregion"
	"goleague/pkg/logger"
	"log"
	"time"
)

// SubRegionQueueConfig is the configuration for the queues that will be executed.
type SubRegionQueueConfig struct {
	Ranks         []string
	HighElos      []string
	Queues        []string
	SleepDuration time.Duration
	Tiers         []string
}

// SubRegionQueue is the type for the sub region main process.
type SubRegionQueue struct {
	config    SubRegionQueueConfig
	logger    *logger.NewLogger
	service   subregionservice.SubRegionService
	subRegion regions.SubRegion
}

// NewDefaultQueueConfig returns a default configuration for the sub region.
func NewDefaultQueueConfig() *SubRegionQueueConfig {
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

// NewSubRegionQueue creates the sub region queue.
func NewSubRegionQueue(region regions.SubRegion, rm *regionmanager.RegionManager) (*SubRegionQueue, error) {
	// Create the service.
	service, err := rm.GetSubService(region)
	if err != nil {
		log.Printf("Failed to get sub region fetcher for %v: %v", region, err)
		return nil, err
	}

	logger := service.GetLogger()

	// Return the new region service.
	return &SubRegionQueue{
		config:    *NewDefaultQueueConfig(),
		logger:    logger,
		service:   *service,
		subRegion: region,
	}, nil
}

// Run starts the sub region queue.
// Mainly responsible for getting the ratings for each player on the region.
func (q *SubRegionQueue) Run() {
	for {
		startTime := time.Now()
		q.processQueues()

		q.logger.Infof("Finished executing after %v minutes.", time.Since(startTime).Minutes())

		objectKey := fmt.Sprintf("subregions/%s/%s.log", q.subRegion, time.Now().Format("2006-01-02-15-04"))
		if err := q.logger.UploadToS3Bucket(objectKey); err != nil {
			log.Printf("Couldn't send the log to s3: %v", err)

			// Clean the file in the case it was a S3 error and not a file error.
			q.logger.CleanFile()
		} else {
			log.Printf("Successfully sent log to s3 with key: %s", objectKey)
		}

		// Sleep to wait new matches to happen.
		time.Sleep(q.config.SleepDuration)
	}
}

// processQueues process the leagues for the SoloDuo and Flex queue.
func (q *SubRegionQueue) processQueues() {
	for _, queue := range q.config.Queues {
		q.processHighElo(queue)
		q.processLeagues(queue)
	}
}

// processHighElo process the high elo leagues.
func (q *SubRegionQueue) processHighElo(queue string) {
	// Go through each high elo entry.
	for _, highElo := range q.config.HighElos {

		// Add the start for the logger.
		q.logger.EmptyLine()
		q.logger.Infof("Starting fetching on %s: Queue(%s)", highElo, queue)
		q.logger.EmptyLine()
		if err := q.service.ProcessHighElo(highElo, queue); err != nil {
			q.logger.Errorf("Couldn't process the high elo %s for the queue %s on region %s: %v", highElo, queue, q.subRegion, err)
			continue
		}
	}
}

// processLeagues process each league and sub rank.
func (q *SubRegionQueue) processLeagues(queue string) {
	// Loop through each available tier.
	for _, tier := range q.config.Tiers {
		// Loop through each available rank.
		for _, rank := range q.config.Ranks {
			q.logger.EmptyLine()
			q.logger.Infof("Starting fetching on %s-%s: Queue(%s)", tier, rank, queue)
			q.logger.EmptyLine()

			if err := q.service.ProcessLeagueRank(tier, rank, queue); err != nil {
				q.logger.Errorf("Couldn't process the league %s - rank %s for the queue %s on region %s: %v", tier, rank, queue, q.subRegion, err)
				continue
			}
		}
	}
}
