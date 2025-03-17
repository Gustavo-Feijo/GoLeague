package models

import (
	"fmt"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/pkg/database"
	"time"

	"gorm.io/gorm"
)

// Database model for the match information.
type MatchInfo struct {
	ID             uint   `gorm:"primaryKey"`
	GameVersion    string `gorm:"type:varchar(20)"`
	MatchId        string `gorm:"type:varchar(20);uniqueIndex"`
	MatchStart     time.Time
	MatchDuration  int
	MatchWinner    bool
	MatchSurrender bool
	MatchRemake    bool
	QueueId        int `gorm:"index"`
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

// Database model for saving the match bans.
type MatchBans struct {
	MatchId    uint  `gorm:"primaryKey;autoIncrement:false"`
	PickTurn   uint8 `gorm:"primaryKey;autoIncrement:false"`
	ChampionId int16
}

// Match service structure.
type MatchService struct {
	db *gorm.DB
}

// Create a match service.
func CreateMatchService() (*MatchService, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &MatchService{db: db}, nil
}
