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
	Puuid          string `gorm:"index;uniqueIndex:idx_player_region;type:char(78)"` // Unique identifier.
	RiotIdGameName string `gorm:"type:varchar(30)"`                                  // Shouldn't have more than 16.
	RiotIdTagline  string `gorm:"type:varchar(5)"`
	SummonerId     string `gorm:"type:char(63)"`
	SummonerLevel  uint16
	Region         string `gorm:"type:varchar(5);uniqueIndex:idx_player_region"` // Sometimes the same player can be found on other leagues.
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

// Create multiple passed players.
func (ps *PlayerService) CreatePlayersBatch(players []*PlayerInfo) error {
	if len(players) == 0 {
		return nil
	}

	// Creates in batches of 1000.
	return ps.db.CreateInBatches(&players, 1000).Error
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

// Get a list of players by a list of passed PUUIDs.
func (s *PlayerService) GetPlayersByPuuids(puuids []string) (map[string]*PlayerInfo, error) {
	// Empty list, just return nil.
	if len(puuids) == 0 {
		return nil, nil
	}

	// Get the players.
	var players []PlayerInfo
	result := s.db.Where("puuid IN (?)", puuids).Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to make it faster.
	playersMap := make(map[string]*PlayerInfo, len(players))
	for i := range players {
		playersMap[players[i].Puuid] = &players[i]
	}

	return playersMap, nil
}
