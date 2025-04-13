package mainregionservice

import (
	"errors"
	"fmt"
	"goleague/fetcher/data"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	"log"
	"strconv"
	"time"
)

// Type for the default configuration.
type mainRegionConfig struct {
	maxRetries int
}

// Type for the main region service.
type MainRegionService struct {
	config             mainRegionConfig
	fetcher            data.MainFetcher
	MatchRepository    repositories.MatchRepository
	PlayerService      repositories.PlayerRepository
	RatingRepository   repositories.RatingRepository
	TimelineRepository repositories.TimelineRepository
	MainRegion         regions.MainRegion
}

// Create the main region default config.
func createMainRegionConfig() *mainRegionConfig {
	return &mainRegionConfig{
		maxRetries: 3,
	}
}

// Create the main region service.
func CreateMainRegionService(
	fetcher *data.MainFetcher,
	region regions.MainRegion,
) (*MainRegionService, error) {
	// Create the services.
	ratingRepository, err := repositories.NewRatingRepository()
	if err != nil {
		return nil, errors.New("failed to start the rating service")
	}

	playerService, err := repositories.NewPlayerRepository()
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

	// Return the new region service.
	return &MainRegionService{
		config:             *createMainRegionConfig(),
		fetcher:            *fetcher,
		MatchRepository:    matchRepository,
		PlayerService:      playerService,
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
func (p *MainRegionService) GetMatchData(
	matchId string,
) (*match_fetcher.MatchData, error) {
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
func (p *MainRegionService) GetMatchTimeline(
	matchId string,
) (*match_fetcher.MatchTimeline, error) {
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

// Retrieve the bans and create them.
func (p *MainRegionService) processMatchBans(
	matchTeams []match_fetcher.TeamInfo,
	matchInfo *models.MatchInfo,
) ([]*models.MatchBans, error) {
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
		if err := p.MatchRepository.CreateMatchBans(bans); err != nil {
			return nil, err
		}
	}

	return bans, nil
}

// Retrieve the match info from the received payload, parse it and insert into the database.
func (p *MainRegionService) processMatchInfo(
	match *match_fetcher.MatchData,
	matchId string,
) (*models.MatchInfo, error) {
	// Create a match to be inserted.
	matchInfo := &models.MatchInfo{
		GameVersion:    match.Info.GameVersion,
		MatchId:        matchId,
		MatchStart:     match.Info.GameCreation.Time(),
		MatchDuration:  match.Info.GameDuration,
		MatchSurrender: match.Info.Participants[0].GameEndedInSurrender,
		MatchRemake:    match.Info.Participants[0].GameEndedInEarlySurrender,
		QueueId:        match.Info.QueueId,
	}

	// Create the match.
	// Return the match that we tried to insert and the error result of the insert (Nil or error)
	return matchInfo, p.MatchRepository.CreateMatchInfo(matchInfo)
}

// Process to insert the match stats for each player.
func (p *MainRegionService) processMatchStats(
	playersToUpsert []*models.PlayerInfo,
	participants map[string]match_fetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
) ([]*models.MatchStats, error) {
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
	if err := p.MatchRepository.CreateMatchStats(statsToUpsert); err != nil {
		return nil, err
	}

	return statsToUpsert, nil
}

// Process each player from a given match.
// Upsert the players, only updating the data if the match data is newer.
func (p *MainRegionService) processPlayersFromMatch(
	participants []match_fetcher.MatchPlayer,
	matchInfo *models.MatchInfo,
	region regions.SubRegion,
) ([]*models.PlayerInfo, map[string]match_fetcher.MatchPlayer, error) {
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
			Region:         region,
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

// Process the match data to insert it into the database.
// Wrapper to call all other necessary functions.
func (p *MainRegionService) ProcessMatchData(
	match *match_fetcher.MatchData,
	matchId string,
	region regions.SubRegion,
) (*models.MatchInfo, []*models.MatchBans, []*models.MatchStats, error) {
	// Process the match infos.
	matchInfo, err := p.processMatchInfo(match, matchId)
	if err != nil {
		log.Printf("Couldn't create the match info for the match %s: %v", matchId, err)
		return nil, nil, nil, err
	}

	// Process the bans.
	bans, err := p.processMatchBans(match.Info.Teams, matchInfo)
	if err != nil {
		log.Printf("Couldn't create the bans for the match %s: %v", matchInfo.MatchId, err)
		return nil, nil, nil, err
	}

	// Process each player.
	playersToUpsert, participantByPuuid, err := p.processPlayersFromMatch(match.Info.Participants, matchInfo, region)
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
func (p *MainRegionService) ProcessMatchTimeline(
	matchTimeline *match_fetcher.MatchTimeline,
	statIdByPuuid map[string]uint64,
	matchInfo *models.MatchInfo,
) error {
	// Extract the stat ID for each participant entry.
	statIdByParticipantId := make(map[string]uint64)
	for _, participant := range matchTimeline.Info.Participants {
		pId := strconv.Itoa(participant.ParticipantId) // Converting to string to match the frames type.
		statIdByParticipantId[pId] = statIdByPuuid[participant.Puuid]
	}

	// Get the default frame interval.
	frameInterval := matchTimeline.Info.FrameInterval
	if err := p.MatchRepository.SetFrameInterval(matchInfo.ID, frameInterval); err != nil {
		return fmt.Errorf("couldn't save the frame interval: %v", err)
	}

	// Create the frames slice and the event collector for handling batch insert.
	// Currently creating the batches for the entire matches.
	// Can be changed to running one batch for each frame.
	var framesToInsert []*models.ParticipantFrame
	eventCollector := createBatchCollector()

	// Loop through each available frame.
	for frameIndex, frame := range matchTimeline.Info.Frames {
		// Loop through the map of participants.
		for participantId, frameData := range frame.ParticipantFrames {

			// Get the stat id based on the participant id.
			matchStatId := statIdByParticipantId[participantId]

			// Append the participant frame to the list to batch insert.
			framesToInsert = append(framesToInsert, p.prepareParticipantsFrames(frameData, matchStatId, frameIndex))
		}

		// Loop through each event frame available.
		for _, event := range frame.Event {
			if err := p.prepareEvents(event, matchInfo, statIdByParticipantId, eventCollector); err != nil {
				log.Printf("Couldn't insert event %s on timestamp %d on match %s: %v", event.Type, event.Timestamp, matchInfo.MatchId, err)
			}
		}
	}

	// Insert the participant frames in a batch.
	if err := p.TimelineRepository.CreateBatchParticipantFrame(framesToInsert); err != nil {
		log.Printf("Couldn't insert the participant frames on  match %s: %v", matchInfo.MatchId, err)
		return err
	}

	// Process the events.
	err := eventCollector.processBatches()

	return err
}

// Prepare a champion kill event.
func (p *MainRegionService) prepareChampionKill(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) (*models.EventPlayerKill, error) {
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

	var victimId string

	if event.VictimId != nil {
		victimId = strconv.Itoa(*event.VictimId)
	}

	var victimMatchStatId *uint64
	if val, exists := statIdByParticipantId[victimId]; exists {
		victimMatchStatId = &val
	}

	// Validate the positions existence for caution.
	x, xExist := event.Position["x"]
	y, yExist := event.Position["y"]

	// Default values if not defined.
	if !xExist || !yExist {
		x = 0
		y = 0
	}

	eventInsert := &models.EventPlayerKill{
		MatchId:           matchInfo.ID,
		Timestamp:         event.Timestamp,
		MatchStatId:       matchStatId,
		VictimMatchStatId: victimMatchStatId,
		X:                 x,
		Y:                 y,
	}

	return eventInsert, nil
}

// Prepare each event to be inserted.
// Handle the events as any/interface{}
// Add each event to the batch collector for further batch insertion.
func (p *MainRegionService) prepareEvents(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
	batchCollector *batchCollector,
) error {
	// Use a default error variable.
	var err error

	// Interface for handling the batch collection.
	var eventData any

	// Handle multiple event types in different ways.
	switch event.Type {

	case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
		eventData, err = p.prepareStructKillEvent(event, matchInfo, statIdByParticipantId)

	case "CHAMPION_KILL":
		eventData, err = p.prepareChampionKill(event, matchInfo, statIdByParticipantId)

	case "FEAT_UPDATE":
		eventData, err = p.prepareFeatUpdateEvent(event, matchInfo)

	case "ITEM_DESTROYED", "ITEM_PURCHASED", "ITEM_SOLD", "ITEM_UNDO":
		eventData, err = p.prepareItemEvent(event, statIdByParticipantId)

	case "LEVEL_UP":
		eventData, err = p.prepareLevelUpEvent(event, statIdByParticipantId)

	case "SKILL_LEVEL_UP":
		eventData, err = p.prepareSkillLevelUpEvent(event, statIdByParticipantId)

	case "WARD_KILL", "WARD_PLACED":
		eventData, err = p.prepareWardEvent(event, statIdByParticipantId)

	case "ELITE_MONSTER_KILL":
		eventData, err = p.prepareMonsterKill(event, statIdByParticipantId)

	case "GAME_END":
		err = p.setMatchWinner(event, matchInfo)

	}

	if err != nil {
		return err
	}

	// Add the generated event data.
	// Mostly this verification is not relevant, since nil will still be passed and handled on the processing.
	if eventData != nil {
		batchCollector.Add(event.Type, eventData)
	}

	return nil
}

// Prepare a feat update event.
func (p *MainRegionService) prepareFeatUpdateEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
) (*models.EventFeatUpdate, error) {
	var (
		featType  int
		featValue int
		teamId    int
	)

	// Verify the values.
	if event.FeatType != nil {
		featType = *event.FeatType
	} else {
		return nil, errors.New("missing feat type")
	}

	if event.FeatValue != nil {
		featValue = *event.FeatValue
	} else {
		return nil, errors.New("missing feat value")
	}

	if event.TeamId != nil {
		teamId = *event.TeamId
	} else {
		return nil, errors.New("missing team ID")
	}

	eventInsert := &models.EventFeatUpdate{
		MatchId:   matchInfo.ID,
		Timestamp: event.Timestamp,
		FeatType:  featType,
		FeatValue: featValue,
		TeamId:    teamId,
	}

	return eventInsert, nil
}

// Prepare a item event.
func (p *MainRegionService) prepareItemEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) (*models.EventItem, error) {
	// Get the killer id as string if setted.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return nil, errors.New("the participant ID was not defined for a item event")
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
	} else if event.Type == "ITEM_UNDO" && event.BeforeId != nil {
		// Handle the ITEM_UNDO case.
		itemId = *event.BeforeId
	} else {
		return nil, errors.New("no item ID found for item event")
	}

	eventInsert := &models.EventItem{
		MatchStatId: matchStatId,
		Timestamp:   event.Timestamp,
		ItemId:      itemId,
		AfterId:     event.AfterId,
		Action:      event.Type,
	}

	return eventInsert, nil
}

// Process a level up event.
func (p *MainRegionService) prepareLevelUpEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) (*models.EventLevelUp, error) {
	// Get the participant that level up.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return nil, errors.New("the participant ID was not defined for the level up event")
	}

	// Arena and Aram start with level 3 and don't return the participant ID in those cases.
	if participantId == "0" {
		return nil, nil
	}

	// Get a pointer to the match statId.
	// Must be nil if the kill was by a minion.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[participantId]; exists {
		matchStatId = &val
	} else {
		return nil, fmt.Errorf("coulnd't find the stat entry for the participant %s", participantId)
	}

	eventInsert := &models.EventLevelUp{
		MatchStatId: *matchStatId,
		Timestamp:   event.Timestamp,
		Level:       *event.Level,
	}

	return eventInsert, nil
}

// Prepare a monster kill event.
func (p *MainRegionService) prepareMonsterKill(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) (*models.EventMonsterKill, error) {
	// Get the killer id as string if setted.
	var killerId string
	if event.KillerId != nil {
		killerId = strconv.Itoa(*event.KillerId)
	}

	// Get the match statId.
	var matchStatId uint64
	if val, exists := statIdByParticipantId[killerId]; exists {
		matchStatId = val
	} else {
		return nil, errors.New("missing match stat id for monster kill")
	}

	var team int
	if event.KillerTeamId != nil {
		team = *event.KillerTeamId
	} else {
		return nil, errors.New("missing team that killed the monster")
	}

	var monsterType string
	if event.MonsterType != nil {
		monsterType = *event.MonsterType
	} else {
		return nil, errors.New("monster type not defined")
	}

	// Validate the positions existence for caution.
	x, xExist := event.Position["x"]
	y, yExist := event.Position["y"]

	// Default values if not defined.
	if !xExist || !yExist {
		x = 0
		y = 0
	}

	eventInsert := &models.EventMonsterKill{
		MatchStatId: matchStatId,
		Timestamp:   event.Timestamp,
		KillerTeam:  team,
		X:           x,
		Y:           y,
		MonsterType: monsterType,
	}

	return eventInsert, nil
}

// Prepare the participant frame and return it to be later inserted.
func (p *MainRegionService) prepareParticipantsFrames(
	frame match_fetcher.ParticipantFrames,
	matchStatId uint64,
	frameId int,
) *models.ParticipantFrame {
	// Create the participant to be inserted in the database.
	participant := &models.ParticipantFrame{
		MatchStatId:       matchStatId,
		FrameIndex:        frameId,
		ParticipantFrames: frame,
	}

	return participant
}

// Prepare a skill level up.
func (p *MainRegionService) prepareSkillLevelUpEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) (*models.EventSkillLevelUp, error) {
	// Get the participant that leveled up.
	var participantId string
	if event.ParticipantId != nil {
		participantId = strconv.Itoa(*event.ParticipantId)
	} else {
		return nil, errors.New("the participant ID was not defined for the skill level up event")
	}

	// Get a pointer to the match statId.
	var matchStatId *uint64
	if val, exists := statIdByParticipantId[participantId]; exists {
		matchStatId = &val
	} else {
		return nil, fmt.Errorf("coulnd't find the stat entry for the participant %s", participantId)
	}

	// Get the skill that was leveled up.
	var skillSlot int
	if event.SkillSlot != nil {
		skillSlot = *event.SkillSlot
	} else {
		return nil, errors.New("the skill slot was not defined for the level up")
	}

	eventInsert := &models.EventSkillLevelUp{
		MatchStatId: *matchStatId,
		Timestamp:   event.Timestamp,
		SkillSlot:   skillSlot,
		LevelUpType: *event.LevelUpType,
	}

	return eventInsert, nil
}

// Prepare a struct kill event.
func (p *MainRegionService) prepareStructKillEvent(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
) (*models.EventKillStruct, error) {
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
		return nil, errors.New("missing team ID on a struct kill")
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

	return eventInsert, nil
}

// Prepare a ward event.
func (p *MainRegionService) prepareWardEvent(
	event match_fetcher.EventFrame,
	statIdByParticipantId map[string]uint64,
) (*models.EventWard, error) {
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

		return eventInsert, nil
	}

	// Can't have a ward without creator or killer.
	return nil, errors.New("couldn't find the actor of the ward event")
}

// Set the match winner.
func (p *MainRegionService) setMatchWinner(
	event match_fetcher.EventFrame,
	matchInfo *models.MatchInfo,
) error {
	var teamId int
	if event.WinningTeam != nil {
		teamId = *event.WinningTeam
	}

	return p.MatchRepository.SetMatchWinner(matchInfo.ID, teamId)
}
