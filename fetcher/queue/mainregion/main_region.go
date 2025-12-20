package mainregionqueue

import (
	"context"
	"fmt"
	regionmanager "goleague/fetcher/regionmanager"
	mainregionservice "goleague/fetcher/services/mainregion"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"
	"goleague/pkg/regions"
	"log"
	"time"
)

// MainRegionQueueConfig is the configuration for the queues that will be executed.
type MainRegionQueueConfig struct {
	SleepDuration time.Duration
}

// MainRegionQueue is the type for the main region main process.
type MainRegionQueue struct {
	config         MainRegionQueueConfig
	fetchedMatches int
	logger         *logger.NewLogger
	mainRegion     regions.MainRegion
	service        mainregionservice.MainRegionService
	subRegions     []regions.SubRegion
}

// NewDefaultQueueConfig returns a default configuration for the main region.
func NewDefaultQueueConfig() *MainRegionQueueConfig {
	return &MainRegionQueueConfig{
		SleepDuration: 5 * time.Second,
	}
}

// NewMainRegionQueue creates the main region queue.
func NewMainRegionQueue(region regions.MainRegion, rm *regionmanager.RegionManager) (*MainRegionQueue, error) {
	// Create the service.
	service, err := rm.GetMainService(region)
	if err != nil {
		log.Printf("Failed to get main service for %v: %v", region, err)
		return nil, err
	}

	// Get the sub regions to define which region must be fetched.
	subRegions, err := rm.GetSubRegions(region)
	if err != nil {
		log.Printf("Failed to get the sub regions for: %v: %v", region, err)
		return nil, err
	}

	logger := service.GetLogger()

	// Return the new region service.
	return &MainRegionQueue{
		config:     *NewDefaultQueueConfig(),
		logger:     logger,
		mainRegion: region,
		service:    *service,
		subRegions: subRegions,
	}, nil
}

// Run starts the main region.
func (q *MainRegionQueue) Run() {
	startTime := time.Now()

	// Infinite loop, must be always getting data.
	for {
		// Loop through each possible subRegion so we can get a evenly distributed amount of matches.
		for _, subRegion := range q.subRegions {
			player, err := q.processQueue(subRegion)
			if err == nil || player == nil {
				continue
			}
			// Delay the player next fetch to avoid the queue getting stuck.
			if err := q.service.PlayerRepository.SetDelayedLastFetch(player.ID); err != nil {
				q.logger.Errorf("Couldn't delay the next fetch for the player.")
			}

		}

		if q.fetchedMatches < 100 {
			continue
		}

		// if we processed 100 matches, upload the log and continue fetching.
		q.logger.EmptyLine()
		q.logger.Infof("Finished executing after %v minutes.", time.Since(startTime).Minutes())

		objectKey := fmt.Sprintf("mainregions/%s/%s.log", q.mainRegion, time.Now().Format("2006-01-02-15-04-05"))
		if err := q.logger.UploadToS3Bucket(objectKey); err != nil {
			log.Printf("Couldn't send the log to s3: %v", err)
			q.logger.CleanFile()
		} else {
			log.Printf("Successfully sent log to s3 with key: %s", objectKey)
		}

		q.fetchedMatches = 0
		startTime = time.Now()
	}
}

// processQueue gets a unfetched player and starts processing it's matches.
func (q *MainRegionQueue) processQueue(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	player, err := q.service.PlayerRepository.GetNextFetchPlayerBySubRegion(subRegion)
	if err != nil {
		q.logger.Errorf("Couldn't get any unfetched player on regions %v: %v", subRegion, err)
		// Could be the first fetch, wait to the sub regions to start filling the database.
		time.Sleep(q.config.SleepDuration)
		return nil, err
	}

	q.logger.EmptyLine()
	q.logger.Infof("Starting fetching for player: %v", player.ID)
	q.logger.EmptyLine()

	// Background fetching needs only 1 worker at a time.
	jobWorkers := 1
	ctx := context.Background()
	player, fetched, err := q.service.ProcessPlayerHistory(ctx, player, subRegion, q.logger, jobWorkers, false)
	q.fetchedMatches += fetched

	return player, err
}
