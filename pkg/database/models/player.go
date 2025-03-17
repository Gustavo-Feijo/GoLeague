package models

import (
	"errors"
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"time"

	"gorm.io/gorm"
)

// Create the player into the database.
type PlayerInfo struct {
	ID             uint `gorm:"primaryKey"`
	ProfileIcon    uint16
	Puuid          string `gorm:"index;uniqueIndex:idx_player;type:char(78)"` // Unique identifier.
	RiotIdGameName string `gorm:"type:varchar(30)"`                           // Shouldn't have more than 16.
	RiotIdTagline  string `gorm:"type:varchar(5)"`
	SummonerId     string `gorm:"type:char(63)"`
	SummonerLevel  uint16
	Region         string `gorm:"type:varchar(5)"`
	UnfetchedMatch bool

	// Last time the user match was fetched.
	LastMatchFetch time.Time
}

// Player service structure.
type PlayerService struct {
	db *gorm.DB
}

// Create a player service.
func CreatePlayerService() (*PlayerService, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &PlayerService{db: db}, nil
}

// Create a basic player structure.
func (ps *PlayerService) CreatePlayerFromRating(rating league_fetcher.LeagueEntry, region regions.SubRegion) (*PlayerInfo, error) {
	insertEntry := &PlayerInfo{
		Puuid:          rating.Puuid,
		SummonerId:     rating.SummonerId,
		Region:         string(region),
		UnfetchedMatch: true,
	}

	// Create a basic player structure.
	err := ps.db.Create(insertEntry).Error
	if err != nil {
		return nil, fmt.Errorf("couldn't create the player basic entry for the player with puuid %v: %v", rating.Puuid, err)
	}

	return insertEntry, nil
}

// Get a given player by his PUUID.
func (ps *PlayerService) GetPlayerByPuuid(puuid string) (*PlayerInfo, error) {
	// Retrieve player by PUUID
	var player PlayerInfo
	if err := ps.db.Where("puuid = ?", puuid).First(&player).Error; err != nil {
		// If the record was not found, doesn't need to return a error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Other database error.
		return nil, fmt.Errorf("player not found: %v", err)
	}

	return &player, nil
}
