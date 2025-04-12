package models

import (
	"fmt"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/pkg/database"

	"gorm.io/gorm"
)

// The timeline entries are always attached to a given player stat entry.
// Since events can only be attached to players that played the match.
// By making it like that, we can avoid creation of composite keys like the ones in the stats structure.

// The participant frame already come in a ready to insert structure.
type ParticipantFrame struct {
	// Composite primary key, a given player in a given match can have multiple frames.
	MatchStatId uint64 `gorm:"primaryKey"`
	FrameIndex  int    `gorm:"primaryKey"`

	// Foreign Key
	MatchStat                       MatchStats `gorm:"MatchStatId"`
	match_fetcher.ParticipantFrames `gorm:"embedded"`
}

// Event of a player level up.
type EventLevelUp struct {
	// Composite primary key, a given player can have multiple level ups.
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key
	MatchStat MatchStats `gorm:"MatchStatId"`

	Level int
}

// Event of a player leveling up a skill.
type EventSkillLevelUp struct {
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key
	MatchStat MatchStats `gorm:"MatchStatId"`

	LevelUpType string `gorm:"type:varchar(30)"`
	SkillSlot   int
}

// Event of a player doing something with an item.
type EventItem struct {
	MatchStatId *uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key
	MatchStat MatchStats `gorm:"MatchStatId"`

	ItemId int
	Action string
}

// Event of a ward.
type EventWard struct {
	// Sometimes the participant ID will be setted as 0 for some reason.
	// So we cannot find who created/killed the ward.
	MatchStatId *uint64 `gorm:"index"`
	Timestamp   int64

	EventType string `gorm:"not null"`
	WardType  *string
}

// Event of killing a ward/plate/tower.
type EventKillStruct struct {
	// Struct kills not necessary are attached to a player.
	// A minion can kill a struct, so we need to handle differently.
	MatchId   uint `gorm:"index"`
	Timestamp int64
	TeamId    int

	EventType   string
	MatchStatId *uint64

	BuildingType *string
	LaneType     *string
	TowerType    *string
	X            int
	Y            int
}

type EventFeatUpdate struct {
	MatchId uint `gorm:"index"`

	Timestamp int64

	FeatType  int
	FeatValue int
	TeamId    int
}

// Event of a player kill.
type EventPlayerKill struct {
	// Champions kills not necessary are attached to a player.
	// A minion can kill a player, so we need to handle differently.
	MatchId   uint `gorm:"index"`
	Timestamp int64

	MatchStatId *uint64

	VictimMatchStatId *uint64
	X                 int
	Y                 int
}

// Event of destroying a monster.
type EventMonsterKill struct {
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key
	MatchStat MatchStats `gorm:"MatchStatId"`

	KillerTeam  bool
	MonsterType string
	X           int
	Y           int
}

// Timeline service structure.
type TimelineService struct {
	db *gorm.DB
}

// Create a match service.
func CreateTimelineService() (*TimelineService, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &TimelineService{db: db}, nil
}

// Create the participant frames in batches.
func (ts *TimelineService) CreateBatchParticipantFrame(frames []*ParticipantFrame) error {
	if len(frames) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&frames, 1000).Error
}

// Create the struct kills events in batches.
func (ts *TimelineService) CreateBatchStructKill(events []*EventKillStruct) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the item events in batches.
func (ts *TimelineService) CreateBatchItemEvent(events []*EventItem) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the ward events in batches.
func (ts *TimelineService) CreateBatchWardEvent(events []*EventWard) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the skill level up events in batches.
func (ts *TimelineService) CreateBatchSkillLevelUpEvent(events []*EventSkillLevelUp) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the level up events in batches.
func (ts *TimelineService) CreateBatchLevelUpEvent(events []*EventLevelUp) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the feat updates events in batches.
func (ts *TimelineService) CreateBatchFeatUpdateEvent(events []*EventFeatUpdate) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}

// Create the player kill events in batches.
func (ts *TimelineService) CreateBatchPlayerKillEvent(events []*EventPlayerKill) error {
	if len(events) == 0 {
		return nil
	}
	return ts.db.CreateInBatches(&events, 1000).Error
}
