package match_fetcher

// Default match timeline.
type MatchTimeline struct {
	Info MatchTimelineData `json:"info"`
}

// Date of the timeline.
type MatchTimelineData struct {
	EndOfGameResult string                      `json:"endOfGameResult"`
	FrameInterval   int64                       `json:"frameInterval"`
	Frames          []MatchTimelineFrame        `json:"frames"`
	Participants    []MatchTimelineParticipants `json:"participants"`
}

// Frame generated every FrameInterval interval.
type MatchTimelineFrame struct {
	Event             []EventFrame                 `json:"events"`
	ParticipantFrames map[string]ParticipantFrames `json:"participantFrames"`
}

// Frame with the events.
type EventFrame struct {
	CreatorId     *int           `json:"creatorId,omitempty"`
	FeatType      *int           `json:"featType,omitempty"`
	FeatValue     *int           `json:"featValue,omitempty"`
	ItemId        *int           `json:"itemId,omitempty"`
	KillerId      *int           `json:"killerId,omitempty"`
	Level         *int           `json:"level,omitempty"`
	LevelUpType   *string        `json:"levelUpType,omitempty"`
	RealTimestamp int64          `json:"realTimestamp"`
	Position      map[string]int `json:"position,omitempty"`
	ParticipantId *int           `json:"participantId,omitempty"`
	SkillSlot     *int           `json:"skillSlot,omitempty"`
	TeamId        *int           `json:"teamId,omitempty"`
	Timestamp     int64          `json:"timestamp"`
	Type          string         `json:"type"`
	VictimId      *int           `json:"victimId,omitempty"`
	WardType      *string        `json:"wardType,omitempty"`
	BuildingType  *string        `json:"buildingType,omitempty"`
	LaneType      *string        `json:"laneType,omitempty"`
	TowerType     *string        `json:"towerType,omitempty"`
}

// Frame for each participant.
type ParticipantFrames struct {
	CurrentGold         int         `json:"currentGold"`
	DamageStats         DamageStats `json:"damageStats" gorm:"embedded"`
	JungleMinionsKilled int         `json:"jungleMinionsKilled"`
	Level               int         `json:"level"`
	MinionsKilled       int         `json:"minionsKilled"`
	ParticipantId       int         `json:"participantId"`
	TotalGold           int         `json:"totalGold"`
	XP                  int         `json:"xp"`
}

// Damage stats for a given participant.
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

// Each participant with it's respective ID inside the timeline.
type MatchTimelineParticipants struct {
	ParticipantId int    `json:"participantId"`
	Puuid         string `json:"puuid"`
}
