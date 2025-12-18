package mainregionservice

import (
	"context"
	"errors"
	"fmt"
	"goleague/fetcher/data"
	"goleague/fetcher/repositories"
	"goleague/pkg/config"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"
	"goleague/pkg/regions"
	queuevalues "goleague/pkg/riotvalues/queue"
	"slices"
	"sync"
	"time"

	playerfetcher "goleague/fetcher/data/player"

	eventservice "goleague/fetcher/services/mainregion/events"
	matchservice "goleague/fetcher/services/mainregion/match"
	playerservice "goleague/fetcher/services/mainregion/player"

	"gorm.io/gorm"
)

// MainRegionConfig is the configuration of the main region.
type MainRegionConfig struct {
	MaxRetries int
}

// Result of a single match fetch.
type matchResult struct {
	matchId     string
	totalTime   time.Duration
	fetchTime   time.Duration
	processTime time.Duration
	err         error
}

// MainRegionService coordinates data fetching and processing for a specific main region.
type MainRegionService struct {
	config             MainRegionConfig
	fetcher            data.MainFetcher
	eventService       *eventservice.EventService
	matchService       *matchservice.MatchService
	playerService      *playerservice.PlayerService
	timelineService    *matchservice.TimelineService
	MatchRepository    repositories.MatchRepository
	PlayerRepository   repositories.PlayerRepository
	RatingRepository   repositories.RatingRepository
	TimelineRepository repositories.TimelineRepository
	logger             *logger.NewLogger
	MainRegion         regions.MainRegion
}

// newMainRegionConfig creates the main region default config.
func newMainRegionConfig() *MainRegionConfig {
	return &MainRegionConfig{
		MaxRetries: 3,
	}
}

// NewMainRegionService creates the main region service.
func NewMainRegionService(
	config *config.Config,
	db *gorm.DB,
	fetcher *data.MainFetcher,
	region regions.MainRegion,
) (*MainRegionService, error) {
	// Create the repositores.
	ratingRepository, err := repositories.NewRatingRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	playerRepository, err := repositories.NewPlayerRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	matchRepository, err := repositories.NewMatchRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the match service")
	}

	timelineRepository, err := repositories.NewTimelineRepository(db)
	if err != nil {
		return nil, errors.New("failed to start the timeline service")
	}

	mainRegionConfig := *newMainRegionConfig()

	// Create the logger.
	logger, err := logger.CreateLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to start the logger on sub region %s: %v", region, err)
	}

	// Create the services.
	eventservice := eventservice.NewEventService(
		matchRepository,
	)

	playerService := playerservice.NewPlayerService(
		matchRepository,
		playerRepository,
		ratingRepository,
	)

	matchService := matchservice.NewMatchService(
		*fetcher,
		matchRepository,
		playerRepository,
		ratingRepository,
		timelineRepository,
		playerService,
		mainRegionConfig.MaxRetries,
	)

	// Passing the raw db as well to use in the batch collector.
	timelineService := matchservice.NewTimelineService(
		db,
		*fetcher,
		timelineRepository,
		mainRegionConfig.MaxRetries,
	)

	// Return the new region service.
	return &MainRegionService{
		config:             mainRegionConfig,
		fetcher:            *fetcher,
		eventService:       eventservice,
		matchService:       matchService,
		playerService:      playerService,
		timelineService:    timelineService,
		MatchRepository:    matchRepository,
		PlayerRepository:   playerRepository,
		RatingRepository:   ratingRepository,
		TimelineRepository: timelineRepository,
		logger:             logger,
		MainRegion:         region,
	}, nil
}

// GetLogger returns the logger instance.
// Used for manual closing of the logs.
func (p *MainRegionService) GetLogger() *logger.NewLogger {
	return p.logger
}

// getFullMatchList retrieves the full match list of a given player.
func (p *MainRegionService) getFullMatchList(
	player *models.PlayerInfo,
) ([]string, error) {
	var matchList []string

	// Go through each page of the match history.
	// The only condition for the stop is the match history being empty.
	for offset := 0; ; offset += 100 {
		// Holds matches and errors through the attempts.
		var matches []string
		var err error

		for attempt := 1; attempt < int(p.config.MaxRetries); attempt += 1 {
			// Get the player matches.
			matches, err = p.fetcher.Player.GetMatchList(player.Puuid, player.LastMatchFetch, offset, false)

			// Everything went right, just continue normally..
			if err == nil {
				break
			}

			// Wait 5 seconds in case anything is wrong with the Riot API and try again.
			time.Sleep(5 * time.Second)
		}

		// Couldn't get even after multiple attempts.
		if err != nil {
			return nil, fmt.Errorf("couldn't get the players match list: %v", err)
		}

		// No matches found, we got the entire match list.
		if len(matches) == 0 {
			return matchList, nil
		}
		matchList = append(matchList, matches...)
	}
}

// GetTrueMatchList retrieves the matches that need to be fetched for a given player.
// Remove all matches that were already fetched.
func (p *MainRegionService) GetTrueMatchList(
	player *models.PlayerInfo,
) ([]string, error) {
	var trueMatchList []string

	matchList, err := p.getFullMatchList(player)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the full match list even after retrying: %v", err)
	}

	alreadyFetchedList, err := p.MatchRepository.GetAlreadyFetchedMatches(matchList)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the already fetched matches: %v", err)
	}

	matchByMatchId := make(map[string]models.MatchInfo)
	for _, fetched := range alreadyFetchedList {
		matchByMatchId[fetched.MatchId] = fetched
	}

	// Loop through each match.
	for _, matchId := range matchList {
		// If it wasn't fetched already, then fetch it.
		_, exists := matchByMatchId[matchId]
		if !exists {
			trueMatchList = append(trueMatchList, matchId)
		}
	}
	return trueMatchList, nil
}

// GetTrueMatchList retrieves the matches that need to be fetched for a given player.
// Remove all matches that were already fetched.
func (p *MainRegionService) GetAccount(
	gameName string,
	tagLine string,
) (*playerfetcher.Account, error) {
	account, err := p.fetcher.Player.GetPlayerAccount(gameName, tagLine, true)
	if err != nil {
		return nil, fmt.Errorf("player not found: %v", err)
	}

	return account, nil
}

// GetPlayerByNameTagRegion is a wripper to the player service call..
func (p *MainRegionService) GetPlayerByNameTagRegion(
	gameName string,
	gameTag string,
	region string,
) (*models.PlayerInfo, error) {
	return p.playerService.GetPlayerByNameTagRegion(gameName, gameTag, region)
}

// ProcessPlayerHistory process the player match history with Goroutines.
func (p *MainRegionService) ProcessPlayerHistory(
	ctx context.Context,
	player *models.PlayerInfo,
	subRegion regions.SubRegion,
	logger *logger.NewLogger,
	maxConcurrency int,
	onDemand bool,
) (
	*models.PlayerInfo,
	int,
	error,
) {
	fetchedMatches := 0
	select {
	default:
		trueMatchList, err := p.GetTrueMatchList(player)
		if err != nil {
			logger.Errorf("Couldn't get the true match list: %v", err)
			return player, 0, err
		}

		matchChan := make(chan string, len(trueMatchList))
		resultChan := make(chan matchResult, len(trueMatchList))

		// Fill the match channel
		for _, matchId := range trueMatchList {
			matchChan <- matchId
		}
		close(matchChan)

		// Start worker goroutines
		var wg sync.WaitGroup

		for range maxConcurrency {
			wg.Add(1)
			go p.matchWorker(matchChan, resultChan, subRegion, &wg, onDemand)
		}

		// Close result channel when all workers are done
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		var firstError error

		for result := range resultChan {
			if result.err != nil {
				logger.Errorf("Error processing match %s: %v", result.matchId, result.err)
				if firstError == nil {
					firstError = result.err
				}
				continue
			}
			fetchedMatches++
			logger.Infof("Created: Match %-15s on %1.2f seconds: FetchTime (%1.2f) - ProcessingTime(%1.2f)",
				result.matchId,
				result.totalTime.Seconds(),
				result.fetchTime.Seconds(),
				result.processTime.Seconds(),
			)
		}

		// Set the last fetch regardless of any match processing errors
		if err := p.PlayerRepository.SetFetched(player.ID); err != nil {
			logger.Errorf("Couldn't set the last fetch date for the player with ID %d: %v", player.ID, err)
		}

		return player, fetchedMatches, firstError
	case <-ctx.Done():
		return nil, fetchedMatches, ctx.Err()
	}
}

// matchWorker processes matches from the channel.
func (p *MainRegionService) matchWorker(
	matchChan <-chan string,
	resultChan chan<- matchResult,
	subRegion regions.SubRegion,
	wg *sync.WaitGroup,
	onDemand bool,
) {
	defer wg.Done()

	for matchId := range matchChan {
		result := p.processMatch(matchId, subRegion, onDemand)
		resultChan <- result
	}
}

func (p *MainRegionService) processMatch(
	matchId string,
	subRegion regions.SubRegion,
	onDemand bool,
) matchResult {
	matchfetchStart := time.Now()
	matchData, err := p.matchService.GetMatchData(matchId, onDemand)
	if err != nil {
		return matchResult{
			matchId: matchId,
			err:     fmt.Errorf("couldn't get the match data for the match %s: %v", matchId, err),
		}
	}

	// Skip modes that are not treated.
	// They can have bots, which mess with the PUUIDs logic.
	if !slices.Contains(queuevalues.TreatedQueues, matchData.Info.QueueId) {
		return matchResult{
			matchId: matchId,
			err:     fmt.Errorf("match %s is of untreated gamemode: %d", matchId, matchData.Info.QueueId),
		}
	}

	matchParseStart := time.Now()

	matchInfo, _, matchStats, err := p.matchService.ProcessMatchData(matchData, matchId, subRegion)
	if err != nil {
		return matchResult{
			matchId: matchId,
			err:     fmt.Errorf("couldn't process the data for the match %s: %v", matchId, err),
		}
	}

	// Create a map of each inserted stat id by the puuid.
	statByPuuid := make(map[string]uint64)
	for _, stat := range matchStats {
		statByPuuid[stat.PlayerData.Puuid] = stat.ID
	}

	timelineFetchStart := time.Now()
	matchTimeline, err := p.timelineService.GetMatchTimeline(matchId, onDemand)
	if err != nil {
		return matchResult{
			matchId: matchId,
			err:     fmt.Errorf("couldn't get the match timeline for the match %s: %v", matchId, err),
		}

	}

	timelineParseStart := time.Now()
	err = p.timelineService.ProcessMatchTimeline(matchTimeline, statByPuuid, matchInfo, p.MatchRepository)
	if err != nil {
		return matchResult{
			matchId: matchId,
			err:     fmt.Errorf("couldn't process the timeline data for the match %s: %v", matchId, err),
		}
	}

	p.MatchRepository.SetFullyFetched(matchInfo.ID)

	return matchResult{
		matchId:     matchId,
		err:         nil,
		totalTime:   time.Since(matchfetchStart),
		fetchTime:   matchParseStart.Sub(matchfetchStart) + timelineParseStart.Sub(timelineFetchStart),
		processTime: timelineFetchStart.Sub(matchParseStart) + time.Since(timelineParseStart),
	}
}
