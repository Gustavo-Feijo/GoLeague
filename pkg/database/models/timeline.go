package models

import (
	"gorm.io/datatypes"
)

// The timeline entries are almost always attached to a given player stat entry.

// EventBase is the base structure for all events.
type EventBase struct {
	MatchID   uint      `gorm:"index:idx_match_participant;not null"`
	MatchInfo MatchInfo `gorm:"foreignKey:MatchID"`

	ParticipantID *int  `gorm:"index:idx_match_participant"` // Nullable, some events can be made by non-participants (like minions).
	Timestamp     int64 `gorm:"index;not null"`
}

// EventFeatUpdate contains data regarding feats of strength.
type EventFeatUpdate struct {
	MatchID   uint      `gorm:"index"`
	MatchInfo MatchInfo `gorm:"foreignKey:MatchID"`

	Timestamp int64

	FeatType  int
	FeatValue int
	TeamId    int
}

// EventItem contains data regarding players and items.
type EventItem struct {
	EventBase `gorm:"embedded"`

	ItemId int

	// If the event is of ITEM_UNDO, there will be a after ID.
	AfterId *int
	Action  string
}

// EventKillStruct contains data regarding killings structs (Towers, Plates, etc).
type EventKillStruct struct {
	EventBase `gorm:"embedded"`
	TeamId    int

	EventType string

	BuildingType *string
	LaneType     *string
	TowerType    *string
	X            int
	Y            int
}

// EventLevelUp contains data regarding a givne player level up.
type EventLevelUp struct {
	// Composite primary key, a given player can have multiple level ups.
	EventBase `gorm:"embedded"`

	Level int
}

// EventMonsterKill contains data regarding the death of a monster.
type EventMonsterKill struct {
	EventBase `gorm:"embedded"`

	KillerTeam  int
	MonsterType string
	X           int
	Y           int
}

// EventPlayerKill contains data regarding a champion being killed.
// Can come from another players or minions and towers.
type EventPlayerKill struct {
	EventBase `gorm:"embedded"`

	VictimParticipantId *int
	X                   int
	Y                   int
}

// EventSkillLevelUp contains data regarding a skill level up.
type EventSkillLevelUp struct {
	EventBase `gorm:"embedded"`

	LevelUpType string `gorm:"type:varchar(30)"`
	SkillSlot   int
}

// EventWard contains data regarding a ward/vision event.
type EventWard struct {
	EventBase `gorm:"embedded"`

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
	MatchStat MatchStats `gorm:"MatchStatId"`

	CurrentGold                   int
	MagicDamageDone               int
	MagicDamageDoneToChampions    int
	MagicDamageTaken              int
	PhysicalDamageDone            int
	PhysicalDamageDoneToChampions int
	PhysicalDamageTaken           int
	TotalDamageDone               int
	TotalDamageDoneToChampions    int
	TotalDamageTaken              int
	TrueDamageDone                int
	TrueDamageDoneToChampions     int
	TrueDamageTaken               int
	JungleMinionsKilled           int
	Level                         int
	MinionsKilled                 int
	ParticipantId                 int
	TotalGold                     int
	XP                            int
}

// View created to getting all events data together.
type AllEvents struct {
	MatchId       uint
	Timestamp     int64
	EventType     string
	ParticipantId *int
	Data          datatypes.JSON `gorm:"type:jsonb"`
}

func (AllEvents) TableName() string {
	return "all_events"
}
