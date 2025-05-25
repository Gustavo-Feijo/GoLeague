package models

import (
	matchfetcher "goleague/fetcher/data/match"
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
	PlayerData matchfetcher.MatchPlayer `gorm:"embedded"`
}

// MatchBans contains the bans made in a given match.
type MatchBans struct {
	MatchId    uint `gorm:"primaryKey;autoIncrement:false"`
	PickTurn   int  `gorm:"primaryKey;autoIncrement:false"`
	ChampionId int
}
