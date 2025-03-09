package models

import (
	match_fetcher "goleague/fetcher/data/match"
	"time"
)

// Database model for the match information.
type MatchInfo struct {
	ID             uint `gorm:"primaryKey"`
	GameVersion    string
	MatchId        string `gorm:"uniqueIndex"`
	MatchStart     time.Time
	MatchDuration  int
	MatchWinner    bool
	MatchSurrender bool
	MatchRemake    bool
	QueueId        int
}

// Database model for saving a player perfomance in a given match.
type MatchStats struct {
	// Composite primary key, since the same player can't be twice on the same match.
	MatchId  uint `gorm:"primaryKey;autoIncrement:false"`
	PlayerId uint `gorm:"primaryKey;autoIncrement:false"`

	// Foreign keys.
	Match  MatchInfo  `gorm:"MatchId"`
	Player PlayerInfo `gorm:"PlayerId"`

	// Embedded match stats.
	PlayerData match_fetcher.MatchPlayer `gorm:"embedded"`
}
