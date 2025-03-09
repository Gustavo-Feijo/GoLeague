package models

import "time"

// Create the player into the database.
type PlayerInfo struct {
	ID             uint `gorm:"primaryKey"`
	ProfileIcon    uint16
	Puuid          string `gorm:"index;uniqueIndex:idx_player"` // Unique identifier.
	RiotIdGameName string
	RiotIdTagline  string
	SummonerId     string
	SummonerLevel  uint16
	Region         string

	// Last time the user match was fetched.
	LastMatchFetch time.Time
}
