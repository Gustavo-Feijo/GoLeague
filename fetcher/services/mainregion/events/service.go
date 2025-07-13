package eventservice

import (
	"errors"
	"fmt"
	matchfetcher "goleague/fetcher/data/match"
	"goleague/fetcher/repositories"
	batchservice "goleague/fetcher/services/mainregion/batch"
	"goleague/pkg/database/models"
	"reflect"

	"strconv"
)

// EventService is a separated service for event operations.
type EventService struct {
	MatchRepository repositories.MatchRepository
}

// NewEventService creates a new event service.
func NewEventService(
	matchRepo repositories.MatchRepository,
) *EventService {
	return &EventService{
		MatchRepository: matchRepo,
	}
}

// prepareChampionKill prepares a champion kill event.
func (es *EventService) prepareChampionKill(
	event matchfetcher.EventFrame,
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

// PrepareEvents prepares each event to be inserted.
// Handle the events as any/interface{}.
// Add each event to the batch collector for further batch insertion.
func (es *EventService) PrepareEvents(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
	statIdByParticipantId map[string]uint64,
	batchCollector *batchservice.BatchCollector,
) error {
	// Use a default error variable.
	var err error

	// Interface for handling the batch collection.
	var eventData any

	// Handle multiple event types in different ways.
	switch event.Type {

	case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
		eventData, err = es.prepareStructKillEvent(event, matchInfo, statIdByParticipantId)

	case "CHAMPION_KILL":
		eventData, err = es.prepareChampionKill(event, matchInfo, statIdByParticipantId)

	case "FEAT_UPDATE":
		eventData, err = es.prepareFeatUpdateEvent(event, matchInfo)

	case "ITEM_DESTROYED", "ITEM_PURCHASED", "ITEM_SOLD", "ITEM_UNDO":
		eventData, err = es.prepareItemEvent(event, statIdByParticipantId)

	case "LEVEL_UP":
		eventData, err = es.prepareLevelUpEvent(event, statIdByParticipantId)

	case "SKILL_LEVEL_UP":
		eventData, err = es.prepareSkillLevelUpEvent(event, statIdByParticipantId)

	case "WARD_KILL", "WARD_PLACED":
		eventData, err = es.prepareWardEvent(event, statIdByParticipantId)

	case "ELITE_MONSTER_KILL":
		eventData, err = es.prepareMonsterKill(event, statIdByParticipantId)

	case "GAME_END":
		err = es.setMatchWinner(event, matchInfo)

	}

	if err != nil {
		return err
	}

	// Add the generated event data.
	// Mostly this verification is not relevant, since nil will still be passed and handled on the processing.
	if eventData != nil && !reflect.ValueOf(eventData).IsNil() {
		batchCollector.Add(event.Type, eventData)
	}

	return nil
}

// prepareFeatUpdateEvent prepares a feat update event.
func (es *EventService) prepareFeatUpdateEvent(
	event matchfetcher.EventFrame,
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

// prepareItemEvent prepares a item event.
func (es *EventService) prepareItemEvent(
	event matchfetcher.EventFrame,
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

// prepareLevelUpEvent prepares a level up event.
func (es *EventService) prepareLevelUpEvent(
	event matchfetcher.EventFrame,
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

// prepareMonsterKill prepares a monster kill event.
func (es *EventService) prepareMonsterKill(
	event matchfetcher.EventFrame,
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

// prepareSkillLevelUpEvent prepares a skill level up.
func (es *EventService) prepareSkillLevelUpEvent(
	event matchfetcher.EventFrame,
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

// prepareStructKillEvent prepares a struct kill event.
func (es *EventService) prepareStructKillEvent(
	event matchfetcher.EventFrame,
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

// prepareWardEvent prepares a ward event.
func (es *EventService) prepareWardEvent(
	event matchfetcher.EventFrame,
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

// setMatchWinner set the match winner.
func (es *EventService) setMatchWinner(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
) error {
	var teamId int
	if event.WinningTeam != nil {
		teamId = *event.WinningTeam
	}

	return es.MatchRepository.SetMatchWinner(matchInfo.ID, teamId)
}
