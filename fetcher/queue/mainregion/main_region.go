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
	"time"
)

// The configuration for the queues we will be executing.
type MainRegionQueueConfig struct {
	SleepDuration time.Duration
}

// Type for the sub region main process.
type MainRegionQueue struct {
	config         MainRegionQueueConfig
	fetchedMatches int
	logger         *logger.NewLogger
	mainRegion     regions.MainRegion
	service        mainregionservice.MainRegionService
	subRegions     []regions.SubRegion
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

		q.fetchedMatches = 0
		startTime = time.Now()
	}
}

// Process each match in order.
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

	// Process matches from the matches channel until closed.
	for _, matchId := range trueMatchList {
		matchfetchStart := time.Now()
		matchData, err := q.service.GetMatchData(matchId)
		if err != nil {
			q.logger.Errorf("Couldn't get the match data for the match %s: %v", matchId, err)
			return player, err
		}

		// Skip modes that are not treated.
		// They can have bots, which mess with the PUUIDs logic.
		if !slices.Contains([]int{400, 420, 430, 440, 450, 490, 900, 1020, 1300, 1400, 1700, 1710, 1900}, matchData.Info.QueueId) {
			q.logger.Errorf("Match %s is of untreated gamemode: %d", matchId, matchData.Info.QueueId)
			continue
		}

		matchParseStart := time.Now()

		matchInfo, _, matchStats, err := q.service.ProcessMatchData(matchData, matchId, subRegion)
		if err != nil {
			q.logger.Errorf("Couldn't process the data for the match %s: %v", matchId, err)
			return player, err
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
			return player, err
		}

		timelineParseStart := time.Now()
		err = q.service.ProcessMatchTimeline(matchTimeline, statByPuuid, matchInfo)

		if err != nil {
			q.logger.Errorf("Couldn't process the timeline data for the match %s: %v", matchId, err)
			return player, err
		}

		q.service.MatchRepository.SetFullyFetched(matchInfo.ID)

		// Log the complete creation of a given match and the elapsed time for verifying performance.
		q.logger.Infof("Created: Match %-15s on %1.2f seconds: FetchTime (%1.2f) - ProcessingTime(%1.2f)",
			matchId,
			time.Since(matchfetchStart).Seconds(),
			matchParseStart.Sub(matchfetchStart).Seconds()+timelineParseStart.Sub(timelineFetchStart).Seconds(),
			timelineFetchStart.Sub(matchParseStart).Seconds()+time.Since(timelineParseStart).Seconds(),
		)

		q.fetchedMatches++
	}

	// Set the last fetch regardless of any match processing errors
	if err := q.service.PlayerRepository.SetFetched(player.ID); err != nil {
		q.logger.Errorf("Couldn't set the last fetch date for the player with ID %d: %v", player.ID, err)
	}

	return player, nil
}
