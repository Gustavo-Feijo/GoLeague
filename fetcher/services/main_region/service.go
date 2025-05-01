package mainregionservice

import (
	"errors"
	"fmt"

	"goleague/fetcher/data"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	eventservice "goleague/fetcher/services/events"
	matchservice "goleague/fetcher/services/match"
	playerservice "goleague/fetcher/services/player"
	"goleague/pkg/database/models"
	"log"
	"time"
)

// Type for the default configuration.
type MainRegionConfig struct {
	MaxRetries int
}

// Type for the main region service.
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
	MainRegion         regions.MainRegion
}

// Create the main region default config.
func createMainRegionConfig() *MainRegionConfig {
	return &MainRegionConfig{
		MaxRetries: 3,
	}
}

// Create the main region service.
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

	config := *createMainRegionConfig()

	// Create the services

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
		MainRegion:         region,
	}, nil
}

// Get the full match list of a given player.
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

// Get the matches that need to be fetched for a given player.
// Remove all matches that were already fetched.
func (p *MainRegionService) GetTrueMatchList(
	player *models.PlayerInfo,
) ([]string, error) {
	var trueMatchList []string

	matchList, err := p.getFullMatchList(player)
	if err != nil {
		log.Printf("Couldn't get the full match list even after retrying: %v", err)
		return nil, err
	}

	alreadyFetchedList, err := p.MatchRepository.GetAlreadyFetchedMatches(matchList)
	if err != nil {
		log.Printf("Couldn't get the already fetched matches: %v", err)
		return nil, err
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

// Get the data of the match from the Riot API.
func (p *MainRegionService) GetMatchData(
	matchId string,
) (*match_fetcher.MatchData, error) {
	return p.matchService.GetMatchData(matchId)
}

// Process the match data to insert it into the database.
// Wrapper to call the service.
func (p *MainRegionService) ProcessMatchData(
	match *match_fetcher.MatchData,
	matchId string,
	region regions.SubRegion,
) (*models.MatchInfo, []*models.MatchBans, []*models.MatchStats, error) {
	return p.matchService.ProcessMatchData(match, matchId, region)
}

// Get the match timeline.
func (p *MainRegionService) GetMatchTimeline(
	matchId string,
) (*match_fetcher.MatchTimeline, error) {
	return p.timelineService.GetMatchTimeline(matchId)
}

// Process the match timeline to insert it into the database.
// Wrapper to call the service.
func (p *MainRegionService) ProcessMatchTimeline(
	matchTimeline *match_fetcher.MatchTimeline,
	statIdByPuuid map[string]uint64,
	matchInfo *models.MatchInfo,
) error {
	return p.timelineService.ProcessMatchTimeline(matchTimeline, statIdByPuuid, matchInfo, p.MatchRepository)
}
