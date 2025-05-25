package models

import (
	"goleague/fetcher/regions"
	"time"
)

// RatingEntry contains a given rating entry for a player at a given moment.
type RatingEntry struct {
	ID uint `gorm:"primaryKey"`

	// Reference to the player that has the rating.
	PlayerId uint       `gorm:"index:index_player_queue_time,priority:1"`
	Player   PlayerInfo `gorm:"PlayerId"`

	Queue        string `gorm:"type:queue_type;index:idx_player_queue_time,priority:2"`
	Tier         string `gorm:"type:tier_type"`
	Rank         string `gorm:"type:rank_type"`
	NumericScore int
	LeaguePoints int
	Wins         int
	Losses       int
	Region       regions.SubRegion `gorm:"type:varchar(5)"`
	FetchTime    time.Time         `gorm:"autoCreateTime;index:index_player_queue_time,priority:3"`
}
