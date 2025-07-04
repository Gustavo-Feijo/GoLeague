package matchfetcher

// MatchTimeline is the default return, with the info key.
type MatchTimeline struct {
	Info MatchTimelineData `json:"info"`
}

// MatchTimelineData is the struct containing the main content of the timeline.
type MatchTimelineData struct {
	EndOfGameResult string                      `json:"endOfGameResult"`
	FrameInterval   int64                       `json:"frameInterval"`
	Frames          []MatchTimelineFrame        `json:"frames"`
	Participants    []MatchTimelineParticipants `json:"participants"`
}

// MatchTimelineFrame contains the participant frames and events for a given frame interval.
type MatchTimelineFrame struct {
	Event             []EventFrame                `json:"events"`
	ParticipantFrames map[string]ParticipantFrame `json:"participantFrames"`
}

// EventFrame is a struct of possible values for a event.
type EventFrame struct {
	AfterId       *int           `json:"afterId,omitempty"`
	BeforeId      *int           `json:"beforeId,omitempty"`
	BuildingType  *string        `json:"buildingType,omitempty"`
	CreatorId     *int           `json:"creatorId,omitempty"`
	FeatType      *int           `json:"featType,omitempty"`
	FeatValue     *int           `json:"featValue,omitempty"`
	ItemId        *int           `json:"itemId,omitempty"`
	KillerId      *int           `json:"killerId,omitempty"`
	KillerTeamId  *int           `json:"killerTeamId,omitempty"`
	LaneType      *string        `json:"laneType,omitempty"`
	Level         *int           `json:"level,omitempty"`
	LevelUpType   *string        `json:"levelUpType,omitempty"`
	MonsterType   *string        `json:"monsterType,omitempty"`
	ParticipantId *int           `json:"participantId,omitempty"`
	Position      map[string]int `json:"position,omitempty"`
	RealTimestamp int64          `json:"realTimestamp"`
	SkillSlot     *int           `json:"skillSlot,omitempty"`
	TeamId        *int           `json:"teamId,omitempty"`
	Timestamp     int64          `json:"timestamp"`
	TowerType     *string        `json:"towerType,omitempty"`
	Type          string         `json:"type"`
	VictimId      *int           `json:"victimId,omitempty"`
	WardType      *string        `json:"wardType,omitempty"`
	WinningTeam   *int           `json:"winningTeam,omitempty"`
}

// ParticipantFrame contains a single player frame.
type ParticipantFrame struct {
	CurrentGold         int         `json:"currentGold"`
	DamageStats         DamageStats `json:"damageStats" gorm:"embedded"`
	JungleMinionsKilled int         `json:"jungleMinionsKilled"`
	Level               int         `json:"level"`
	MinionsKilled       int         `json:"minionsKilled"`
	ParticipantId       int         `json:"participantId"`
	TotalGold           int         `json:"totalGold"`
	XP                  int         `json:"xp"`
}

// DamageStats contain damage related statistics for a given participant.
type DamageStats struct {
	MagicDamageDone               int `json:"magicDamageDone"`
	MagicDamageDoneToChampions    int `json:"magicDamageDoneToChampions"`
	MagicDamageTaken              int `json:"magicDamageTaken"`
	PhysicalDamageDone            int `json:"physicalDamageDone"`
	PhysicalDamageDoneToChampions int `json:"physicalDamageDoneToChampions"`
	PhysicalDamageTaken           int `json:"physicalDamageTaken"`
	TotalDamageDone               int `json:"totalDamageDone"`
	TotalDamageDoneToChampions    int `json:"totalDamageDoneToChampions"`
	TotalDamageTaken              int `json:"totalDamageTaken"`
	TrueDamageDone                int `json:"trueDamageDone"`
	TrueDamageDoneToChampions     int `json:"trueDamageDoneToChampions"`
	TrueDamageTaken               int `json:"trueDamageTaken"`
}

// MatchTimelineParticipants conrains each participant with it's respective ID inside the timeline.
type MatchTimelineParticipants struct {
	ParticipantId int    `json:"participantId"`
	Puuid         string `json:"puuid"`
}
