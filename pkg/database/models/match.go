package models

import (
	"time"
)

// MatchInfo contains the data regarding the match information.
type MatchInfo struct {
	ID             uint   `gorm:"primaryKey"`
	GameVersion    string `gorm:"type:varchar(20)"`
	MatchId        string `gorm:"type:varchar(20);uniqueIndex"`
	MatchStart     time.Time
	MatchDuration  int
	MatchWinner    int
	MatchSurrender bool
	MatchRemake    bool
	AverageRating  float64 `gorm:"index"`
	FrameInterval  int64
	FullyFetched   bool
	QueueId        int       `gorm:"index"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

// MatchStats contains the data regarding a given player perfomance on a given match.
type MatchStats struct {
	// Ids and identifiers for the match stats.
	ID       uint64 `gorm:"primaryKey"`
	MatchId  uint   `gorm:"not null;index:idx_match_player,unique"`
	PlayerId uint   `gorm:"not null;index:idx_match_player,unique"`

	// Foreign keys.
	Match  MatchInfo  `gorm:"MatchId"`
	Player PlayerInfo `gorm:"PlayerId"`

	// Embedded match stats.
	PlayerData MatchPlayer `gorm:"embedded"`
}

// MatchBans contains the bans made in a given match.
type MatchBans struct {
	MatchId    uint `gorm:"primaryKey;autoIncrement:false"`
	PickTurn   int  `gorm:"primaryKey;autoIncrement:false"`
	ChampionId int
}

// MatchPlayer contains the stats and information about a given player in a Match.
// Same as fetchers API return, intermediate struct so it can be used by API without importing the fetcher package.
type MatchPlayer struct {
	Assists                        int
	AllInPings                     int
	AssistMePing                   int
	BaronKills                     int
	BasicPings                     int
	ChampionLevel                  int
	ChampionId                     int        `gorm:"index_champion_position"`
	Challenges                     Challenges `gorm:"embedded"`
	CommandPings                   int
	DangerPings                    int
	Deaths                         int
	EnemyMissingPings              int
	EnemyVisionPings               int
	GameEndedInEarlySurrender      bool `gorm:"-"`
	GameEndedInSurrender           bool `gorm:"-"`
	GetBackPings                   int
	GoldEarned                     int
	GoldSpent                      int
	HoldPings                      int
	Item0                          int
	Item1                          int
	Item2                          int
	Item3                          int
	Item4                          int
	Item5                          int
	Kills                          int
	MagicDamageDealtToChampions    int
	MagicDamageTaken               int
	NeedVisionPings                int
	NeutralMinionsKilled           int
	OnMyWayPings                   int
	ParticipantId                  int
	PhysicalDamageDealtToChampions int
	PhysicalDamageTaken            int
	ProfileIcon                    int `gorm:"-"`
	PushPings                      int
	Puuid                          string `gorm:"-"`
	RetreatPings                   int
	RiotIdGameName                 string `gorm:"-"`
	RiotIdTagline                  string `gorm:"-"`
	SummonerLevel                  int    `gorm:"-"`
	LongestTimeSpentLiving         int
	MagicDamageDealt               int
	TeamId                         int
	TeamPosition                   string `gorm:"index_champion_position"`
	TimeCCingOthers                int
	TotalDamageDealtToChampions    int
	TotalMinionsKilled             int
	TotalTimeSpentDead             int
	TrueDamageDealtToChampions     int
	VisionClearedPings             int
	VisionScore                    int
	WardsKilled                    int
	WardsPlaced                    int
	Win                            bool
}

type Challenges struct {
	AbilityUses        int
	ControlWardsPlaced int
	SkillshotsDodged   int
	SkillshotsHit      int
}
