package models

import (
	match_fetcher "goleague/fetcher/data/match"
	"time"
)

// Database model for the match information.
type MatchInfo struct {
	ID             uint   `gorm:"primaryKey"`
	GameVersion    string `gorm:"type:varchar(20)"`
	MatchId        string `gorm:"type:varchar(20);uniqueIndex"`
	MatchStart     time.Time
	MatchDuration  int
	MatchWinner    int
	MatchSurrender bool
	MatchRemake    bool
	AverageRating  float64
	FrameInterval  int64
	FullyFetched   bool
	QueueId        int `gorm:"index"`
}

// Database model for saving a player perfomance in a given match.
type MatchStats struct {
	// Ids and identifiers for the match stats.
	ID       uint64 `gorm:"primaryKey"`
	MatchId  uint   `gorm:"not null;index:idx_match_player,unique"`
	PlayerId uint   `gorm:"not null;index:idx_match_player,unique"`

	// Foreign keys.
	Match  MatchInfo  `gorm:"MatchId"`
	Player PlayerInfo `gorm:"PlayerId"`

	// Embedded match stats.
	PlayerData match_fetcher.MatchPlayer `gorm:"embedded"`
}

// Database model for saving the match bans.
type MatchBans struct {
	MatchId    uint `gorm:"primaryKey;autoIncrement:false"`
	PickTurn   int  `gorm:"primaryKey;autoIncrement:false"`
	ChampionId int
}
