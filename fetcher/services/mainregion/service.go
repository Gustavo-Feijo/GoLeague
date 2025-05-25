package mainregionservice

import (
	"errors"
	"fmt"

	"goleague/fetcher/data"
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	eventservice "goleague/fetcher/services/mainregion/events"
	matchservice "goleague/fetcher/services/mainregion/match"
	playerservice "goleague/fetcher/services/mainregion/player"
	"goleague/pkg/database/models"
	"goleague/pkg/logger"
	"time"
)

// MainRegionConfig is the configuration of the main region.
type MainRegionConfig struct {
	MaxRetries int
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
	fetcher *data.MainFetcher,
	region regions.MainRegion,
) (*MainRegionService, error) {
	// Create the repositores.
	ratingRepository, err := repositories.NewRatingRepository()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	playerRepository, err := repositories.NewPlayerRepository()
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	matchRepository, err := repositories.NewMatchRepository()
	if err != nil {
		return nil, errors.New("failed to start the match service")
	}

	timelineRepository, err := repositories.NewTimelineRepository()
	if err != nil {
		return nil, errors.New("failed to start the timeline service")
	}

	config := *newMainRegionConfig()

	// Create the logger.
	logger, err := logger.CreateLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to start the logger on sub region %s: %v", region, err)
	}

	// Create the services.
	eventservice := eventservice.NewEventService(
		matchRepository,
	)

	matchService := matchservice.NewMatchService(
		*fetcher,
		matchRepository,
		playerRepository,
		ratingRepository,
		timelineRepository,
		config.MaxRetries,
	)

	playerService := playerservice.NewPlayerService(
		matchRepository,
		playerRepository,
		ratingRepository,
	)

	timelineService := matchservice.NewTimelineService(
		*fetcher,
		timelineRepository,
		config.MaxRetries,
	)

	// Return the new region service.
	return &MainRegionService{
		config:             config,
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

// GetMatchData retrives the data of the match from the Riot API.
func (p *MainRegionService) GetMatchData(
	matchId string,
) (*matchfetcher.MatchData, error) {
	return p.matchService.GetMatchData(matchId)
}

// ProcessMatchData process the match data to insert it into the database.
// Wrapper to call the service.
func (p *MainRegionService) ProcessMatchData(
	match *matchfetcher.MatchData,
	matchId string,
	region regions.SubRegion,
) (*models.MatchInfo, []*models.MatchBans, []*models.MatchStats, error) {
	return p.matchService.ProcessMatchData(match, matchId, region)
}

// GetMatchTimeline retrives the match timeline from the Riot API.
func (p *MainRegionService) GetMatchTimeline(
	matchId string,
) (*matchfetcher.MatchTimeline, error) {
	return p.timelineService.GetMatchTimeline(matchId)
}

// ProcessMatchTimeline process the match timeline to insert it into the database.
// Wrapper to call the service.
func (p *MainRegionService) ProcessMatchTimeline(
	matchTimeline *matchfetcher.MatchTimeline,
	statIdByPuuid map[string]uint64,
	matchInfo *models.MatchInfo,
) error {
	return p.timelineService.ProcessMatchTimeline(matchTimeline, statIdByPuuid, matchInfo, p.MatchRepository)
}
