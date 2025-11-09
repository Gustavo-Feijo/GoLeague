package eventservice

import (
	"errors"
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
) (*models.EventPlayerKill, error) {
	// Validate the positions existence for caution.
	x, xExist := event.Position["x"]
	y, yExist := event.Position["y"]

	// Default values if not defined.
	if !xExist || !yExist {
		x = 0
		y = 0
	}

	eventInsert := &models.EventPlayerKill{
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.KillerId,
			Timestamp:     event.Timestamp,
		},
		VictimParticipantId: event.VictimId,
		X:                   x,
		Y:                   y,
	}

	return eventInsert, nil
}

// PrepareEvents prepares each event to be inserted.
// Handle the events as any/interface{}.
// Add each event to the batch collector for further batch insertion.
func (es *EventService) PrepareEvents(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
	batchCollector *batchservice.BatchCollector,
) error {
	// Use a default error variable.
	var err error

	// Interface for handling the batch collection.
	var eventData any

	// Handle multiple event types in different ways.
	switch event.Type {

	case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
		eventData, err = es.prepareStructKillEvent(event, matchInfo)

	case "CHAMPION_KILL":
		eventData, err = es.prepareChampionKill(event, matchInfo)

	case "FEAT_UPDATE":
		eventData, err = es.prepareFeatUpdateEvent(event, matchInfo)

	case "ITEM_DESTROYED", "ITEM_PURCHASED", "ITEM_SOLD", "ITEM_UNDO":
		eventData, err = es.prepareItemEvent(event, matchInfo)

	case "LEVEL_UP":
		eventData, err = es.prepareLevelUpEvent(event, matchInfo)

	case "SKILL_LEVEL_UP":
		eventData, err = es.prepareSkillLevelUpEvent(event, matchInfo)

	case "WARD_KILL", "WARD_PLACED":
		eventData, err = es.prepareWardEvent(event, matchInfo)

	case "ELITE_MONSTER_KILL":
		eventData, err = es.prepareMonsterKill(event, matchInfo)

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
		MatchID:   matchInfo.ID,
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
	matchInfo *models.MatchInfo,
) (*models.EventItem, error) {
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
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.ParticipantId,
			Timestamp:     event.Timestamp,
		},
		ItemId:  itemId,
		AfterId: event.AfterId,
		Action:  event.Type,
	}

	return eventInsert, nil
}

// prepareLevelUpEvent prepares a level up event.
func (es *EventService) prepareLevelUpEvent(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
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

	eventInsert := &models.EventLevelUp{
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.ParticipantId,
			Timestamp:     event.Timestamp,
		},
		Level: *event.Level,
	}

	return eventInsert, nil
}

// prepareMonsterKill prepares a monster kill event.
func (es *EventService) prepareMonsterKill(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
) (*models.EventMonsterKill, error) {

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
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.KillerId,
			Timestamp:     event.Timestamp,
		},
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
	matchInfo *models.MatchInfo,
) (*models.EventSkillLevelUp, error) {
	// Get the skill that was leveled up.
	var skillSlot int
	if event.SkillSlot != nil {
		skillSlot = *event.SkillSlot
	} else {
		return nil, errors.New("the skill slot was not defined for the level up")
	}

	eventInsert := &models.EventSkillLevelUp{
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.ParticipantId,
			Timestamp:     event.Timestamp,
		},
		SkillSlot:   skillSlot,
		LevelUpType: *event.LevelUpType,
	}

	return eventInsert, nil
}

// prepareStructKillEvent prepares a struct kill event.
func (es *EventService) prepareStructKillEvent(
	event matchfetcher.EventFrame,
	matchInfo *models.MatchInfo,
) (*models.EventKillStruct, error) {
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
		EventBase: models.EventBase{
			MatchID:       matchInfo.ID,
			ParticipantID: event.KillerId,
			Timestamp:     event.Timestamp,
		},
		EventType:    event.Type,
		TeamId:       teamId,
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
	matchInfo *models.MatchInfo,
) (*models.EventWard, error) {
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

		eventInsert := &models.EventWard{
			EventBase: models.EventBase{
				MatchID:       matchInfo.ID,
				ParticipantID: actorIdPtr,
				Timestamp:     event.Timestamp,
			},
			EventType: event.Type,
			WardType:  event.WardType,
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
