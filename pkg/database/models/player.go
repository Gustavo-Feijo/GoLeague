package models

import (
	"goleague/fetcher/regions"
	"time"

	"gorm.io/gorm"
)

// PlayerInfo contains the player information data.
type PlayerInfo struct {
	ID             uint `gorm:"primaryKey"`
	ProfileIcon    int
	Puuid          string `gorm:"index;uniqueIndex:idx_player_region;type:char(78)"` // Unique identifier.
	RiotIdGameName string `gorm:"type:varchar(100);index:idx_name_tag"`              // Shouldn't have more than 16, adding 100 due to some edge cases.
	RiotIdTagline  string `gorm:"type:varchar(5);index:idx_name_tag"`
	SummonerLevel  int
	Region         regions.SubRegion `gorm:"type:varchar(5);uniqueIndex:idx_player_region;index:idx_unfetched_region"` // Sometimes the same player can be found on other leagues.
	UnfetchedMatch bool              `gorm:"default:true;index:idx_unfetched_region"`

	// Last time the user match was fetched.
	LastMatchFetch time.Time `gorm:"default:CURRENT_TIMESTAMP;index"`

	// Last time the player data was changed.
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}

// Set the last time of a metch fetch as 3 months ago.
// That way, will not get too old matches.
func (p *PlayerInfo) BeforeCreate(tx *gorm.DB) (err error) {
	if p.LastMatchFetch.IsZero() { // If LastChecked is not set
		p.LastMatchFetch = time.Now().Add(-3 * 30 * 24 * time.Hour) // 3 months ago
	}
	return nil
}
