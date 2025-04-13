package models

import (
	"goleague/fetcher/regions"
	"time"
)

// Create a rating entry for a given player.
type RatingEntry struct {
	ID uint `gorm:"primaryKey"`

	// Reference to the player that has the rating.
	PlayerId uint       `gorm:"primaryKey;autoIncrement:false"`
	Player   PlayerInfo `gorm:"PlayerId"`

	Queue        string `gorm:"type:queue_type"`
	Tier         string `gorm:"type:tier_type"`
	Rank         string `gorm:"type:rank_type"`
	LeaguePoints int
	Wins         int
	Losses       int
	Region       regions.SubRegion `gorm:"type:varchar(5)"`
	FetchTime    time.Time         `gorm:"autoCreateTime"`
}
