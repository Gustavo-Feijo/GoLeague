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

// Type for the main region processor.
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
			if err := p.processEvent(event, matchInfo, statIdByParticipantId); err != nil {
				log.Printf("Couldn't insert event %s on timestamp %d on match %s: %v", event.Type, event.Timestamp, matchInfo.MatchId, err)
			}
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

// Process each event.
// Use the event type to define which function will be used for processing the event.
func (p *MainRegionProcessor) processEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) error {
	// Use a default error variable.
	var err error
	// Handle multiple event types in different ways.
	switch event.Type {

	case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
		err = p.processStructKillEvent(event, matchInfo, statIdByParticipantId)

	case "CHAMPION_KILL":

	case "FEAT_UPDATE":
		err = p.processFeatUpdateEvent(event, matchInfo)

	case "ITEM_DESTROYED", "ITEM_PURCHASED", "ITEM_SOLD":
		err = p.processItemEvent(event, statIdByParticipantId)

	case "LEVEL_UP":
		err = p.processLevelUpEvent(event, statIdByParticipantId)

	case "SKILL_LEVEL_UP":
		err = p.processSkillLevelUpEvent(event, statIdByParticipantId)

	case "WARD_KILL", "WARD_PLACED":
		err = p.processWardEvent(event, statIdByParticipantId)

	default:
		log.Printf("Missing event type %s:", event.Type)
	}

	return err
}

// Process a struct kill event.
func (p *MainRegionProcessor) processStructKillEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) error {
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
	} else {
		// Shouldn't happen.
		return errors.New("missing team ID on a struct kill")
	}

	// Validate the positions existence for caution.
	x, xExist := event.Position["x"]
	y, yExist := event.Position["y"]

	// Default values if not defined.
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

	// Insert the struct kill.
	if err := p.TimelineService.CreateStructKill(eventInsert); err != nil {
		return fmt.Errorf("error creating a struct kill: %v", err)
	}

	return nil
}

// Process a item event.
func (p *MainRegionProcessor) processItemEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) error {
	// Get the killer id as string if setted.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return errors.New("the participant ID was not defined for a item event")
	}

	// Get a pointer to the match statId.
	// Must be nil if the kill was by a minion.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[participantId]; exists {
		matchStatId = &val
	}

	var itemId int
	if event.ItemId != nil {
		itemId = *event.ItemId
	} else {
		return errors.New("no item ID found for item event")
	}

	eventInsert := &models.EventItem{
		MatchStatId: matchStatId,
		Timestamp:   event.Timestamp,
		ItemId:      itemId,
		Action:      event.Type,
	}

	// Insert the item event.
	if err := p.TimelineService.CreateItemEvent(eventInsert); err != nil {
		return fmt.Errorf("error creating item event for item %d: %v", itemId, err)
	}

	return nil
}

// Process a ward event.
func (p *MainRegionProcessor) processWardEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) error {
	var actorId string
	var actorIdPtr *int

	// Get a pointer to who started the event.
	// Used to handle both ward kill and placement.
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

		// Insert the ward event.
		if err := p.TimelineService.CreateWardEvent(eventInsert); err != nil {
			return fmt.Errorf("couldn't save ward event for stat id %v: %v", matchStatId, err)
		}

		return nil
	}

	// Can't have a ward without creator or killer.
	return errors.New("couldn't find the actor of the ward event")
}

// Process a skill level up.
func (p *MainRegionProcessor) processSkillLevelUpEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) error {
	// Get the participant that leveled up.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return errors.New("the participant ID was not defined for the skill level up event")
	}

	// Get a pointer to the match statId.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[participantId]; exists {
		matchStatId = &val
	} else {
		return fmt.Errorf("coulnd't find the stat entry for the participant %s", participantId)
	}

	// Get the skill that was leveled up.
	var skillSlot int
	if event.SkillSlot != nil {
		skillSlot = *event.SkillSlot
	} else {
		return errors.New("the skill slot was not defined for the level up")
	}

	eventInsert := &models.EventSkillLevelUp{
		MatchStatId: *matchStatId,
		Timestamp:   event.Timestamp,
		SkillSlot:   skillSlot,
		LevelUpType: *event.LevelUpType,
	}

	// Insert the skill level up event.
	if err := p.TimelineService.CreateSkillLevelUpEvent(eventInsert); err != nil {
		return fmt.Errorf("couldn't save skill level up event for stat id %d on skill %d: %v", matchStatId, skillSlot, err)
	}

	return nil
}

// Process a level up event.
func (p *MainRegionProcessor) processLevelUpEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) error {
	// Get the participant that level up.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return errors.New("the participant ID was not defined for the level up event")
	}

	// Arena and Aram start with level 3 and don't return the participant ID in those cases.
	if participantId == "0" {
		return nil
	}

	// Get a pointer to the match statId.
	// Must be nil if the kill was by a minion.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[participantId]; exists {
		matchStatId = &val
	} else {
		return fmt.Errorf("coulnd't find the stat entry for the participant %s", participantId)
	}

	eventInsert := &models.EventLevelUp{
		MatchStatId: *matchStatId,
		Timestamp:   event.Timestamp,
		Level:       *event.Level,
	}

	// Insert the level up event.
	if err := p.TimelineService.CreateLevelUpEvent(eventInsert); err != nil {
		return fmt.Errorf("couldn't save level up event for stat id %v on participant %s: %v", matchStatId, participantId, err)
	}

	return nil
}

// Process a feat update event.
func (p *MainRegionProcessor) processFeatUpdateEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
) error {
	var (
		featType  int
		featValue int
		teamId    int
	)

	// Verify the values.
	if event.FeatType != nil {
		featType = *event.FeatType
	} else {
		return errors.New("missing feat type")
	}

	if event.FeatValue != nil {
		featValue = *event.FeatValue
	} else {
		return errors.New("missing feat value")
	}

	if event.TeamId != nil {
		teamId = *event.TeamId
	} else {
		return errors.New("missing team ID")
	}

	eventInsert := &models.EventFeatUpdate{
		MatchId:   matchInfo.ID,
		Timestamp: event.Timestamp,
		FeatType:  featType,
		FeatValue: featValue,
		TeamId:    teamId,
	}

	if err := p.TimelineService.CreateFeatUpdateEvent(eventInsert); err != nil {
		return fmt.Errorf("couldn't save feat event %d with value %d for team %d: %v", featType, featValue, teamId, err)
	}

	return nil
}
