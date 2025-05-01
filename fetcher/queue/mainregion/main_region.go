package mainregionqueue

import (
	regionmanager "goleague/fetcher/region_manager"
	"goleague/fetcher/regions"
	mainregionservice "goleague/fetcher/services/main_region"
	"goleague/pkg/database/models"
	"log"
	"sync"
	"time"
)

// The configuration for the queues we will be executing.
type MainRegionQueueConfig struct {
	SleepDuration time.Duration
}

// Type for the sub region main process.
type MainRegionQueue struct {
	config     MainRegionQueueConfig
	service    mainregionservice.MainRegionService
	subRegions []regions.SubRegion
}

// Return a default configuration for the sub region.
func NewDefaultQueueConfig() *MainRegionQueueConfig {
	return &MainRegionQueueConfig{
		SleepDuration: 5 * time.Second,
	}
}

// Create the main region queue.
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
		log.Printf("Failed to run the main region service for the region %v: %v", region, err)
		return nil, err
	}

	// Return the new region service.
	return &MainRegionQueue{
		config:     *NewDefaultQueueConfig(),
		service:    *service,
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
				if err := q.service.PlayerRepository.SetDelayedLastFetch(player.ID); err != nil {
					log.Printf("Couldn't delay the next fetch for the player.")
				}
			}
		}
	}
}

func (q *MainRegionQueue) processQueue(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	player, err := q.service.PlayerRepository.GetUnfetchedBySubRegions(subRegion)
	if err != nil {
		log.Printf("Couldn't get any unfetched player on regions %v: %v", subRegion, err)
		// Could be the first fetch, wait to the sub regions to start filling the database.
		time.Sleep(5 * time.Second)
		return nil, err
	}

	trueMatchList, err := q.service.GetTrueMatchList(player)
	if err != nil {
		log.Printf("Couldn't get the true match list: %v", err)
		return player, err
	}

	// Number of max concurrent workers.
	// For the queue with a low rate limit, 2 workers are used to avoid the total fetch time of a match to be greater than the interval.
	const maxWorkers = 2

	// WaitGroup and mutex.
	var wg sync.WaitGroup
	var mu sync.Mutex

	jobChan := make(chan string, len(trueMatchList))
	errChan := make(chan error, len(trueMatchList))

	// Start worker goroutines
	for range maxWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Process matches from the matches channel until closed.
			for matchId := range jobChan {
				matchfetchStart := time.Now()
				matchData, err := q.service.GetMatchData(matchId)
				if err != nil {
					log.Printf("Couldn't get the match data for the match %s: %v", matchId, err)
					errChan <- err
					continue
				}

				// Skip swiftplay matches.
				// They can have bots, which mess with the PUUIDs logic.
				if matchData.Info.GameMode == "SWIFTPLAY" {
					log.Printf("Match %s is Swiftplay", matchId)
					continue
				}

				matchParseStart := time.Now()

				matchInfo, _, matchStats, err := q.service.ProcessMatchData(matchData, matchId, subRegion)
				if err != nil {
					log.Printf("Couldn't process the data for the match %s: %v", matchId, err)
					errChan <- err
					continue
				}

				// Create a map of each inserted stat id by the puuid.
				statByPuuid := make(map[string]uint64)
				for _, stat := range matchStats {
					statByPuuid[stat.PlayerData.Puuid] = stat.ID
				}

				timelineFetchStart := time.Now()
				matchTimeline, err := q.service.GetMatchTimeline(matchId)
				if err != nil {
					log.Printf("Couldn't get the match timeline for the match %s: %v", matchId, err)
					errChan <- err
					continue
				}

				timelineParseStart := time.Now()

				// Lock the processing to avoid deadlocks.
				mu.Lock()
				err = q.service.ProcessMatchTimeline(matchTimeline, statByPuuid, matchInfo)
				mu.Unlock()

				if err != nil {
					log.Printf("Couldn't process the timeline data for the match %s: %v", matchId, err)
					errChan <- err
					continue
				}

				// The lock is not needed.
				q.service.MatchRepository.SetFullyFetched(matchInfo.ID)

				// Log the complete creation of a given match and the elapsed time for verifying performance.
				log.Printf("Created: Match %-15s on %1.2f seconds: FetchTime (%1.2f) - ProcessingTime(%1.2f)",
					matchId,
					time.Since(matchfetchStart).Seconds(),
					matchParseStart.Sub(matchfetchStart).Seconds()+timelineParseStart.Sub(timelineFetchStart).Seconds(),
					timelineFetchStart.Sub(matchParseStart).Seconds()+time.Since(timelineParseStart).Seconds(),
				)
			}
		}()
	}

	// Send the matches to the channel.
	for _, matchId := range trueMatchList {
		jobChan <- matchId
	}

	close(jobChan)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect the first error.
	var processError error
	for err := range errChan {
		if processError == nil {
			processError = err
		}
	}

	// Set the last fetch regardless of any match processing errors
	if err := q.service.PlayerRepository.SetFetched(player.ID); err != nil {
		log.Printf("Couldn't set the last fetch date for the player with ID %d: %v", player.ID, err)
		if processError == nil {
			processError = err
		}
	}

	return player, processError
}
