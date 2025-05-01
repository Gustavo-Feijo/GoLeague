package mainregionqueue

import (
	"fmt"
	regionmanager "goleague/fetcher/region_manager"
	"goleague/fetcher/regions"
	mainregionservice "goleague/fetcher/services/main_region"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"
	"log"
	"slices"
	"sync"
	"time"
)

// The configuration for the queues we will be executing.
type MainRegionQueueConfig struct {
	SleepDuration time.Duration
}

// Type for the sub region main process.
type MainRegionQueue struct {
	config           MainRegionQueueConfig
	fetchedMatches   int
	fetchedMatchesMu sync.Mutex
	logger           *logger.NewLogger
	mainRegion       regions.MainRegion
	service          mainregionservice.MainRegionService
	subRegions       []regions.SubRegion
}

var (
	processError error
	errorOnce    sync.Once
)

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

// Run the main region.
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

		q.fetchedMatchesMu.Lock()
		q.fetchedMatches = 0
		q.fetchedMatchesMu.Unlock()
		startTime = time.Now()
	}
}

func (q *MainRegionQueue) processQueue(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	player, err := q.service.PlayerRepository.GetUnfetchedBySubRegions(subRegion)
	if err != nil {
		q.logger.Errorf("Couldn't get any unfetched player on regions %v: %v", subRegion, err)
		// Could be the first fetch, wait to the sub regions to start filling the database.
		time.Sleep(q.config.SleepDuration)
		return nil, err
	}

	q.logger.EmptyLine()
	q.logger.Infof("Starting fetching for player: %v", player.ID)
	q.logger.EmptyLine()

	trueMatchList, err := q.service.GetTrueMatchList(player)
	if err != nil {
		q.logger.Errorf("Couldn't get the true match list: %v", err)
		return player, err
	}

	// Number of max concurrent workers.
	// For the queue with a low rate limit, 2 workers are used to avoid the total fetch time of a match to be greater than the interval.
	const maxWorkers = 2

	// WaitGroup and mutex.
	var wg sync.WaitGroup
	var mu sync.Mutex

	jobChan := make(chan string, len(trueMatchList))

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
					q.logger.Errorf("Couldn't get the match data for the match %s: %v", matchId, err)
					errorOnce.Do(func() {
						processError = err
					})
					continue
				}

				// Skip modes that are not treated.
				// They can have bots, which mess with the PUUIDs logic.
				if !slices.Contains([]int{400, 420, 430, 440, 450, 490, 900, 1020, 1300, 1400, 1700, 1710, 1900}, matchData.Info.QueueId) {
					q.logger.Errorf("Match %s is of untreated gamemode: %d", matchId, matchData.Info.QueueId)
					continue
				}

				matchParseStart := time.Now()

				mu.Lock()
				matchInfo, _, matchStats, err := q.service.ProcessMatchData(matchData, matchId, subRegion)
				mu.Unlock()

				if err != nil {
					q.logger.Errorf("Couldn't process the data for the match %s: %v", matchId, err)
					errorOnce.Do(func() {
						processError = err
					})
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
					q.logger.Errorf("Couldn't get the match timeline for the match %s: %v", matchId, err)
					errorOnce.Do(func() {
						processError = err
					})
					continue
				}

				timelineParseStart := time.Now()

				// Lock the processing to avoid deadlocks.
				mu.Lock()
				err = q.service.ProcessMatchTimeline(matchTimeline, statByPuuid, matchInfo)
				mu.Unlock()

				if err != nil {
					q.logger.Errorf("Couldn't process the timeline data for the match %s: %v", matchId, err)
					errorOnce.Do(func() {
						processError = err
					})
					continue
				}

				// The lock is not needed.
				q.service.MatchRepository.SetFullyFetched(matchInfo.ID)

				// Log the complete creation of a given match and the elapsed time for verifying performance.
				q.logger.Infof("Created: Match %-15s on %1.2f seconds: FetchTime (%1.2f) - ProcessingTime(%1.2f)",
					matchId,
					time.Since(matchfetchStart).Seconds(),
					matchParseStart.Sub(matchfetchStart).Seconds()+timelineParseStart.Sub(timelineFetchStart).Seconds(),
					timelineFetchStart.Sub(matchParseStart).Seconds()+time.Since(timelineParseStart).Seconds(),
				)

				q.fetchedMatchesMu.Lock()
				q.fetchedMatches++
				q.fetchedMatchesMu.Unlock()
			}
		}()
	}

	// Send the matches to the channel.
	for _, matchId := range trueMatchList {
		jobChan <- matchId
	}

	close(jobChan)

	// Set the last fetch regardless of any match processing errors
	if err := q.service.PlayerRepository.SetFetched(player.ID); err != nil {
		q.logger.Errorf("Couldn't set the last fetch date for the player with ID %d: %v", player.ID, err)
		if processError == nil {
			processError = err
		}
	}

	return player, processError
}
