package models

import (
	matchfetcher "goleague/fetcher/data/match"
)

// The timeline entries are almost always attached to a given player stat entry.
// Since events can only be attached to players that played the match.
// By making it like that, we can avoid creation of composite keys like the ones in the stats structure.
// The few exceptions are events that can be made by minions.

// EventFeatUpdate contains data regarding feats of strenght.
type EventFeatUpdate struct {
	MatchId uint `gorm:"index"`

	Timestamp int64

	FeatType  int
	FeatValue int
	TeamId    int
}

// EventItem contains data regarding players and items.
type EventItem struct {
	MatchStatId *uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key.
	MatchStat MatchStats `gorm:"MatchStatId"`

	ItemId int

	// If the event is of ITEM_UNDO, there will be a after ID.
	AfterId *int
	Action  string
}

// EventKillStruct contains data regarding killings structs (Towers, Plates, etc).
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

// EventLevelUp contains data regarding a givne player level up.
type EventLevelUp struct {
	// Composite primary key, a given player can have multiple level ups.
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key.
	MatchStat MatchStats `gorm:"MatchStatId"`

	Level int
}

// EventMonsterKill contains data regarding the death of a monster.
type EventMonsterKill struct {
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key.
	MatchStat MatchStats `gorm:"MatchStatId"`

	KillerTeam  int
	MonsterType string
	X           int
	Y           int
}

// EventPlayerKill contains data regarding a champion being killed.
// Can come from another players or minions and towers.
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

// EventSkillLevelUp contains data regarding a skill level up.
type EventSkillLevelUp struct {
	MatchStatId uint64 `gorm:"index"`
	Timestamp   int64

	// Foreign Key.
	MatchStat MatchStats `gorm:"MatchStatId"`

	LevelUpType string `gorm:"type:varchar(30)"`
	SkillSlot   int
}

// EventWard contains data regarding a ward/vision event.
type EventWard struct {
	// Sometimes the participant ID will be setted as 0 for some reason.
	// So we cannot find who created/killed the ward.
	MatchStatId *uint64 `gorm:"index"`
	Timestamp   int64

	EventType string `gorm:"not null"`
	WardType  *string
}

// ParticipantFrame contains data regarding a given player at a given time.
// Can be used for generating graphs.
type ParticipantFrame struct {
	// Composite primary key, a given player in a given match can have multiple frames.
	MatchStatId uint64 `gorm:"primaryKey"`
	FrameIndex  int    `gorm:"primaryKey"`

	// Foreign Key.
	MatchStat                     MatchStats `gorm:"MatchStatId"`
	matchfetcher.ParticipantFrame `gorm:"embedded"`
}
