package mainregion_processor

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"log"
	"strconv"
	"time"
)

// Type for the default configuration.
type mainRegionConfig struct {
	maxRetries int
}

// Type for the main region main process.
type MainRegionProcessor struct {
	config          mainRegionConfig
	fetcher         data.MainFetcher
	MatchService    models.MatchService
	PlayerService   models.PlayerService
	RatingService   models.RatingService
	TimelineService models.TimelineService
	SubRegion       regions.MainRegion
}

// Create the main region default config.
func createMainRegionConfig() *mainRegionConfig {
	return &mainRegionConfig{
		maxRetries: 3,
	}
}

// Create the main region processor.
func CreateMainRegionProcessor(fetcher *data.MainFetcher, region regions.MainRegion) (*MainRegionProcessor, error) {
	// Create the services.
	RatingService, err := models.CreateRatingService()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	PlayerService, err := models.CreatePlayerService()
	if err != nil {
		return nil, errors.New("failed to start the player service")
	}

	MatchService, err := models.CreateMatchService()
	if err != nil {
		return nil, errors.New("failed to start the match service")
	}

	TimelineService, err := models.CreateTimelineService()
	if err != nil {
		return nil, errors.New("failed to start the timeline service")
	}

	// Return the new region processor.
	return &MainRegionProcessor{
		config:          *createMainRegionConfig(),
		fetcher:         *fetcher,
		MatchService:    *MatchService,
		PlayerService:   *PlayerService,
		RatingService:   *RatingService,
		TimelineService: *TimelineService,
		SubRegion:       region,
	}, nil
}

// Get the full match list of a given player.
func (p *MainRegionProcessor) getFullMatchList(player *models.PlayerInfo) ([]string, error) {
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
func (p *MainRegionProcessor) GetMatchData(matchId string) (*match_fetcher.MatchData, error) {
	var matchData *match_fetcher.MatchData
	var err error

	for attempt := 1; attempt < p.config.maxRetries; attempt += 1 {
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

// Get the data of the match timeline from the Riot API.
func (p *MainRegionProcessor) GetMatchTimeline(matchId string) (*match_fetcher.MatchTimeline, error) {
	var matchData *match_fetcher.MatchTimeline
	var err error

	for attempt := 1; attempt < p.config.maxRetries; attempt += 1 {
		// Get the match timeline.
		matchData, err = p.fetcher.Match.GetMatchTimelineData(matchId, false)

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
func (p *MainRegionProcessor) GetTrueMatchList(player *models.PlayerInfo) ([]string, error) {
	var trueMatchList []string

	matchList, err := p.getFullMatchList(player)
	if err != nil {
		log.Printf("Couldn't get the full match list even after retrying: %v", err)
		return nil, err
	}

	alreadyFetchedList, err := p.MatchService.GetAlreadyFetchedMatches(matchList)
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

// Process to insert the match info.
func (p *MainRegionProcessor) processMatchInfo(match *match_fetcher.MatchData, matchId string) (*models.MatchInfo, error) {
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
	// Return the match that we tried to insert and the error result of the insert (Nil or error)
	return matchInfo, p.MatchService.CreateMatchInfo(matchInfo)
}

// Process to insert the match info.
func (p *MainRegionProcessor) processMatchBans(matchTeams []match_fetcher.TeamInfo, matchInfo *models.MatchInfo) ([]*models.MatchBans, error) {
	var bans []*models.MatchBans

	// Get all the bans.
	for _, team := range matchTeams {
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
		if err := p.MatchService.CreateMatchBans(bans); err != nil {
			return nil, err
		}
	}

	return bans, nil
}

// Process each player.
func (p *MainRegionProcessor) processPlayers(
	participants []match_fetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
	region regions.SubRegion,
) (
	[]*models.PlayerInfo,
	map[string]match_fetcher.MatchPlayer,
	error,
) {
	// Variables for batching or search.
	var playersToUpsert []*models.PlayerInfo
	participantByPuuid := make(map[string]match_fetcher.MatchPlayer)

	for _, participant := range participants {
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
	if err := p.PlayerService.UpsertPlayerBatch(playersToUpsert); err != nil {
		log.Printf("Couldn't create/update the players for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, err
	}

	return playersToUpsert, participantByPuuid, nil
}

// Process to insert the match stats.
func (p *MainRegionProcessor) processMatchStats(
	playersToUpsert []*models.PlayerInfo,
	participants map[string]match_fetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
) (
	[]*models.MatchStats,
	error,
) {
	var statsToUpsert []*models.MatchStats
	for _, player := range playersToUpsert {
		participant, exists := participants[player.Puuid]
		if !exists {
			// Should never occur.
			log.Println("The participant is not present in the map.")
			return nil, errors.New("the participant is not present in the map")
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
	if err := p.MatchService.CreateMatchStats(statsToUpsert); err != nil {
		return nil, err
	}

	return statsToUpsert, nil
}

// Process the match data to insert it into the database.
func (p *MainRegionProcessor) ProcessMatchData(match *match_fetcher.MatchData, matchId string, region regions.SubRegion) (*models.MatchInfo, []*models.MatchBans, []*models.MatchStats, error) {
	matchInfo, err := p.processMatchInfo(match, matchId)
	if err != nil {
		log.Printf("Couldn't create the match info for the match %s: %v", matchId, err)
		return nil, nil, nil, err
	}

	// For now, the returned bans are not needed.
	bans, err := p.processMatchBans(match.Info.Teams, matchInfo)
	if err != nil {
		log.Printf("Couldn't create the bans for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, nil, err
	}

	// Process each player.
	playersToUpsert, participantByPuuid, err := p.processPlayers(match.Info.Participants, matchInfo, region)
	if err != nil {
		log.Printf("Couldn't create the players for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, nil, err
	}

	// Process the match stats.
	stats, err := p.processMatchStats(playersToUpsert, participantByPuuid, matchInfo)
	if err != nil {
		log.Printf("Couldn't create the stats for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, nil, err
	}

	log.Printf("Created: Match: %s - ID: %v", matchId, matchInfo.ID)
	return matchInfo, bans, stats, nil
}

// Process the match timeline data to insert it into the database.
func (p *MainRegionProcessor) ProcessMatchTimeline(
	matchTimeline *match_fetcher.MatchTimeline,
	statIdByPuuid map[string]uint64,
	matchInfo *models.MatchInfo,
	region regions.SubRegion,
) error {
	// Extract the stat ID for each participant entry.
	statIdByParticipantId := make(map[string]uint64)
	for _, participant := range matchTimeline.Info.Participants {
		pId := strconv.Itoa(participant.ParticipantId) // Converting to string to match the frames type.
		statIdByParticipantId[pId] = statIdByPuuid[participant.Puuid]
	}

	// Get the default frame interval.
	frameInterval := matchTimeline.Info.FrameInterval
	if err := p.MatchService.SetFrameInterval(matchInfo.ID, frameInterval); err != nil {
		return fmt.Errorf("couldn't save the frame interval: %v", err)
	}

	// Loop through each available frame.
	for frameIndex, frame := range matchTimeline.Info.Frames {
		// Loop through the map of participants.
		for participantId, frameData := range frame.ParticipantFrames {

			// Get the stat id based on the participant id.
			matchStatId := statIdByParticipantId[participantId]

			// Process the participant frame.
			_, err := p.processParticipantsFrames(frameData, matchStatId, frameIndex)
			if err != nil {
				return fmt.Errorf("couldn't  save the participant frame with index %d for the participant %s: %v", frameIndex, participantId, err)
			}
		}

		// Loop through each event frame available.
		for _, event := range frame.Event {
			p.processEvent(event, matchInfo, statIdByParticipantId)
		}
	}
	return nil
}

// Process each participant frame.
func (p *MainRegionProcessor) processParticipantsFrames(frame match_fetcher.ParticipantFrames, matchStatId uint64, frameId int) (*models.ParticipantFrame, error) {
	// Create the participant to be inserted in the database.
	participant := &models.ParticipantFrame{
		MatchStatId:       matchStatId,
		FrameIndex:        frameId,
		ParticipantFrames: frame,
	}

	return participant, p.TimelineService.CreateParticipantFrame(participant)
}

func (p *MainRegionProcessor) processEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) error {
	// Handle multiple event types in different ways.
	switch event.Type {
	case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
		_, err := p.processStructKillEvent(event, matchInfo, statIdByParticipantId)
		if err != nil {
			return fmt.Errorf("couldn't process a struct kill on match %s: %v", matchInfo.MatchId, err)
		}

	case "CHAMPION_KILL":

	case "FEAT_UPDATE":

	case "ITEM_DESTROYED":

	case "ITEM_PURCHASED":
	case "ITEM_SOLD":

	case "LEVEL_UP":

	case "SKILL_LEVEL_UP":

	case "WARD_KILL", "WARD_PLACED":
		_, err := p.processWardEvent(event, matchInfo, statIdByParticipantId)
		if err != nil {
			return fmt.Errorf("couldn't process a struct kill on match %s: %v", matchInfo.MatchId, err)
		}
	default:
		log.Printf("Missing event type %s:", event.Type)
	}

	return nil
}

// Process a struct kill event.
func (p *MainRegionProcessor) processStructKillEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) (
	*models.EventKillStruct,
	error,
) {
	// Get the killer id as string if setted.
	var killerId string
	if event.KillerId != nil {
		killerId = strconv.Itoa(*event.KillerId)
	}

	// Get a pointer to the match statId.
	// Must be nil if the kill was by a minion.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[killerId]; exists {
		matchStatId = &val
	}

	// Get the team.
	var teamId int
	if event.TeamId != nil {
		teamId = *event.TeamId
	}

	// Validate the positions existence for caution.
	x, xExist := event.Position["x"]
	y, yExist := event.Position["y"]

	if !xExist || !yExist {
		x = 0
		y = 0
	}

	// Generate the insert.
	eventInsert := &models.EventKillStruct{
		MatchId:      matchInfo.ID,
		EventType:    event.Type,
		TeamId:       teamId,
		MatchStatId:  matchStatId,
		Timestamp:    event.Timestamp,
		BuildingType: event.BuildingType,
		LaneType:     event.LaneType,
		TowerType:    event.TowerType,
		X:            x,
		Y:            y,
	}

	return eventInsert, p.TimelineService.CreateStructKill(eventInsert)
}

// Process a struct kill event.
func (p *MainRegionProcessor) processWardEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) (
	*models.EventWard,
	error,
) {
	var actorId string
	var actorIdPtr *int

	// Get a pointer to who started the event.
	if event.Type == "WARD_KILL" {
		actorIdPtr = event.KillerId
	} else { // WARD_PLACED
		actorIdPtr = event.CreatorId
	}

	// Shouldn't be nil.
	if actorIdPtr != nil {
		// Get the match stat by the id.
		actorId = strconv.Itoa(*actorIdPtr)

		// Get a pointer to the match statId.
		// Must be nil if the kill was by a minion.
		var matchStatId *uint64
		if val, exists := statIdByParticipantId[actorId]; exists {
			matchStatId = &val
		}
		eventInsert := &models.EventWard{
			MatchStatId: matchStatId,
			Timestamp:   event.Timestamp,
			EventType:   event.Type,
			WardType:    event.WardType,
		}
		err := p.TimelineService.CreateWardEvent(eventInsert)
		if err != nil {
			log.Printf("Couldn't save ward event for stat id %v on match %s: %v", matchStatId, matchInfo.MatchId, err)
		}
		return eventInsert, err
	}

	return nil, fmt.Errorf("couldn't find the actor of the ward event on match %s", matchInfo.MatchId)
}
