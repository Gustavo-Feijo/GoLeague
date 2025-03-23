package mainregion_queue

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"log"
	"time"
)

// Type for the default configuration.
type mainRegionConfig struct {
	maxRetries uint
}

// Type for the main region main process.
type mainRegionProcessor struct {
	config        mainRegionConfig
	fetcher       data.MainFetcher
	matchService  models.MatchService
	playerService models.PlayerService
	ratingService models.RatingService
	subRegion     regions.MainRegion
}

// Create the main region default config.
func createMainRegionConfig() *mainRegionConfig {
	return &mainRegionConfig{
		maxRetries: 3,
	}
}

// Create the main region processor.
func createMainRegionProcessor(fetcher *data.MainFetcher, region regions.MainRegion) (*mainRegionProcessor, error) {
	// Create the services.
	ratingService, err := models.CreateRatingService()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	playerService, err := models.CreatePlayerService()
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	matchService, err := models.CreateMatchService()
	if err != nil {
		return nil, errors.New("failed to start the match service")
	}

	// Return the new region processor.
	return &mainRegionProcessor{
		config:        *createMainRegionConfig(),
		fetcher:       *fetcher,
		matchService:  *matchService,
		playerService: *playerService,
		ratingService: *ratingService,
		subRegion:     region,
	}, nil
}

// Run the main region.
func RunMainRegionQueue(region regions.MainRegion, rm *regions.RegionManager) {
	fetcher, err := rm.GetMainFetcher(region)
	if err != nil {
		log.Printf("Failed to get main region fetcher for %v: %v", region, err)
		return
	}

	// Start the processor for the main region.
	processor, err := createMainRegionProcessor(fetcher, region)
	if err != nil {
		log.Printf("Failed to start the main region processor for the region %v: %v", region, err)
		return
	}

	// Get the sub regions to define which region must be fetched.
	subRegions, err := rm.GetSubRegions(region)
	if err != nil {
		log.Printf("Failed to run the main region processor for the region %v: %v", region, err)
	}

	processor.processUnfetched(subRegions)
}

// Process each unfetched player.
func (p *mainRegionProcessor) processUnfetched(subRegions []regions.SubRegion) {
	// Infinite loop, must be always getting data.
	for {
		// Loop through each possible subRegion so we can get a evenly distributed amount of matches.
		for _, subRegion := range subRegions {
			player, err := p.playerService.GetUnfetchedBySubRegions(subRegion)
			if err != nil {
				log.Printf("Couldn't get any unfetched player: %v", err)

				// Could be the first fetch, wait to the sub regions to start filling the database.
				time.Sleep(5 * time.Second)
				continue
			}

			trueMatchList, err := p.getTrueMatchList(player)
			if err != nil {
				log.Printf("Couldn't get the true match list: %v", err)
				continue
			}

			// Loop through each match.
			for _, matchId := range trueMatchList {
				matchData, err := p.getMatchData(matchId)
				if err != nil {
					log.Printf("Couldn't get the match data for the match %v: %v", matchId, err)
				}

				p.processMatchData(matchData, matchId, subRegion)
			}
		}
	}
}

// Get the full match list of a given player.
func (p *mainRegionProcessor) getFullMatchList(player *models.PlayerInfo) ([]string, error) {
	var matchList []string

	// Go through each page of the match history.
	// The only condition for the stop is the match history being empty.
	for offset := 0; ; offset += 100 {
		// Holds matches and errors through the attempts.
		var matches []string
		var err error

		for attempt := 1; attempt < int(p.config.maxRetries); attempt += 1 {
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

// Get the data of the match from the Riot API.
func (p *mainRegionProcessor) getMatchData(matchId string) (*match_fetcher.MatchData, error) {
	var matchData *match_fetcher.MatchData
	var err error

	for attempt := 1; attempt < int(p.config.maxRetries); attempt += 1 {
		// Get the match data.
		matchData, err = p.fetcher.Match.GetMatchData(matchId, false)

		// Everything went right, just continue normally..
		if err == nil {
			break
		}

		// Wait 5 seconds in case anything is wrong with the Riot API and try again.
		time.Sleep(5 * time.Second)
	}

	// Couldn't get even after multiple attempts.
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match data: %v", err)
	}

	return matchData, nil
}

// Get the matches that need to be fetched for a given player.
// Remove all matches that were already fetched.
func (p *mainRegionProcessor) getTrueMatchList(player *models.PlayerInfo) ([]string, error) {
	var trueMatchList []string

	matchList, err := p.getFullMatchList(player)
	if err != nil {
		log.Printf("Couldn't get the full match list even after retrying: %v", err)
		return nil, err
	}

	alreadyFetchedList, err := p.matchService.GetAlreadyFetchedMatches(matchList)
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

// Process the match data to insert it into the database.
func (p *mainRegionProcessor) processMatchData(match *match_fetcher.MatchData, matchId string, region regions.SubRegion) error {
	// Create a match to be inserted.
	matchInfo := &models.MatchInfo{
		GameVersion:    match.Info.GameVersion,
		MatchId:        matchId,
		MatchStart:     match.Info.GameCreation.Time(),
		MatchDuration:  match.Info.GameDuration,
		MatchWinner:    match.Info.Teams[0].Win,
		MatchSurrender: match.Info.Participants[0].GameEndedInSurrender,
		MatchRemake:    match.Info.Participants[0].GameEndedInEarlySurrender,
		QueueId:        match.Info.QueueId,
	}

	// Create the match.
	p.matchService.CreateMatchInfo(matchInfo)

	var bans []*models.MatchBans

	// Get all the bans.
	for _, team := range match.Info.Teams {
		for _, ban := range team.Bans {
			bans = append(bans, &models.MatchBans{
				MatchId:    matchInfo.ID,
				PickTurn:   ban.PickTurn,
				ChampionId: ban.ChampionId,
			})
		}
	}

	// Only need to insert if there is any bans.
	// Some modes don't have bans.
	if len(bans) != 0 {
		// Create the bans.
		if err := p.matchService.CreateMatchBans(bans); err != nil {
			log.Printf("Couldn't create the bans for the match %s: %v", matchId, err)
			return err
		}
	}

	// Variables for batching or search.
	var playersToUpsert []*models.PlayerInfo
	participantByPuuid := make(map[string]match_fetcher.MatchPlayer)

	for _, participant := range match.Info.Participants {
		// Create a player to be inserted.
		player := &models.PlayerInfo{
			ProfileIcon:    participant.ProfileIcon,
			Puuid:          participant.Puuid,
			RiotIdGameName: participant.RiotIdGameName,
			RiotIdTagline:  participant.RiotIdTagline,
			SummonerId:     participant.SummonerId,
			SummonerLevel:  participant.SummonerLevel,
			Region:         string(region),
			UpdatedAt:      matchInfo.MatchStart,
		}

		participantByPuuid[player.Puuid] = participant
		playersToUpsert = append(playersToUpsert, player)
	}

	// Create/update the players.
	if err := p.playerService.UpsertPlayerBatch(playersToUpsert); err != nil {
		log.Printf("Couldn't create/update the players for the match %s: %v", matchId, err)
		return err
	}

	var statsToUpsert []*models.MatchStats
	for _, player := range playersToUpsert {
		participant, exists := participantByPuuid[player.Puuid]
		if !exists {
			// Should never occur.
			log.Println("The participant is not present in the map.")
			return errors.New("the participant is not present in the map")
		}

		// Create the match stats.
		newStat := &models.MatchStats{
			MatchId:    matchInfo.ID,
			PlayerId:   player.ID,
			PlayerData: participant,
		}

		statsToUpsert = append(statsToUpsert, newStat)
	}

	// Create/update the players.
	if err := p.matchService.CreateMatchStats(statsToUpsert); err != nil {
		log.Printf("Couldn't create the stats for the match %s: %v", matchId, err)
		return err
	}

	log.Printf("Created: Match: %s - ID: %v", matchId, matchInfo.ID)
	return p.matchService.SetFullyFetched(matchInfo.ID)
}
