package models

import (
	"errors"
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Create the player into the database.
type PlayerInfo struct {
	ID             uint `gorm:"primaryKey"`
	ProfileIcon    int
	Puuid          string `gorm:"index;uniqueIndex:idx_player_region;type:char(78)"` // Unique identifier.
	RiotIdGameName string `gorm:"type:varchar(100);index:idx_name_tag"`              // Shouldn't have more than 16, adding 100 due to some edge cases.
	RiotIdTagline  string `gorm:"type:varchar(5);index:idx_name_tag"`
	SummonerId     string `gorm:"type:char(63)"`
	SummonerLevel  int
	Region         string `gorm:"type:varchar(5);uniqueIndex:idx_player_region"` // Sometimes the same player can be found on other leagues.
	UnfetchedMatch bool   `gorm:"default:true"`

	// Last time the user match was fetched.
	LastMatchFetch time.Time `gorm:"default:CURRENT_TIMESTAMP"`

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
func (ps *PlayerService) GetPlayersByPuuids(puuids []string) (map[string]*PlayerInfo, error) {
	// Empty list, just return nil.
	if len(puuids) == 0 {
		return nil, nil
	}

	// Get the players.
	var players []PlayerInfo
	result := ps.db.Where("puuid IN (?)", puuids).Find(&players)
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

// Get the players with unfetched matches.
func (ps *PlayerService) GetUnfetchedBySubRegions(subRegion regions.SubRegion) (*PlayerInfo, error) {
	var unfetchedPlayer PlayerInfo
	result := ps.db.Where("unfetched_match = ? AND region = ?", true, subRegion).Order("last_match_fetch ASC").First(&unfetchedPlayer)

	// Verify if there is any error.
	if result.Error != nil {
		return nil, result.Error
	}

	return &unfetchedPlayer, nil
}

// Set the date of the last time fetch to the previous + 1 day.
func (ps *PlayerService) SetDelayedLastFetch(playerId uint) error {
	return ps.db.Model(&PlayerInfo{}).
		Where("id = ?", playerId).
		UpdateColumn("last_match_fetch", gorm.Expr("last_match_fetch + interval '1 day'")).Error
}

// Set the date of the last time fetch.
func (ps *PlayerService) SetFetched(playerId uint) error {
	return ps.db.Model(&PlayerInfo{}).
		Where("id = ?", playerId).
		Updates(
			map[string]interface{}{
				"last_match_fetch": time.Now().UTC(),
				"unfetched_match":  false,
			},
		).Error
}

// Update or create multiple players.
// Only update if the data is newer.
func (ps *PlayerService) UpsertPlayerBatch(players []*PlayerInfo) error {
	return ps.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "puuid"}, {Name: "region"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"profile_icon":      gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.profile_icon ELSE player_infos.profile_icon END"),
			"riot_id_game_name": gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_game_name ELSE player_infos.riot_id_game_name END"),
			"riot_id_tagline":   gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_tagline ELSE player_infos.riot_id_tagline END"),
			"summoner_level":    gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.summoner_level ELSE player_infos.summoner_level END"),
			"updated_at":        gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.updated_at ELSE player_infos.updated_at END"),
		}),
	}).CreateInBatches(&players, 100).Error
}
