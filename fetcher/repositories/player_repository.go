package repositories

import (
	"errors"
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"goleague/pkg/database/models"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Public Interface.
type PlayerRepository interface {
	CreatePlayersBatch(players []*models.PlayerInfo) error
	CreatePlayerFromRating(rating league_fetcher.LeagueEntry, region regions.SubRegion) (*models.PlayerInfo, error)
	GetPlayerByPuuid(puuid string) (*models.PlayerInfo, error)
	GetPlayersByPuuids(puuids []string) (map[string]*models.PlayerInfo, error)
	GetUnfetchedBySubRegions(subRegion regions.SubRegion) (*models.PlayerInfo, error)
	SetDelayedLastFetch(playerId uint) error
	SetFetched(playerId uint) error
	UpsertPlayerBatch(players []*models.PlayerInfo) error
}

// Player repository structure.
type playerRepository struct {
	db *gorm.DB
}

// Mutex for handling the player upsert.
var playerUpsertMutex sync.Mutex

// Create a player repository.
func NewPlayerRepository() (PlayerRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &playerRepository{db: db}, nil
}

// Create multiple passed players.
func (ps *playerRepository) CreatePlayersBatch(players []*models.PlayerInfo) error {
	if len(players) == 0 {
		return nil
	}

	// Creates in batches of 1000.
	return ps.db.CreateInBatches(&players, 1000).Error
}

// Create a basic player structure.
func (ps *playerRepository) CreatePlayerFromRating(rating league_fetcher.LeagueEntry, region regions.SubRegion) (*models.PlayerInfo, error) {
	insertEntry := &models.PlayerInfo{
		Puuid:          rating.Puuid,
		SummonerId:     rating.SummonerId,
		Region:         region,
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
func (ps *playerRepository) GetPlayerByPuuid(puuid string) (*models.PlayerInfo, error) {
	// Retrieve player by PUUID
	var player models.PlayerInfo
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
func (ps *playerRepository) GetPlayersByPuuids(puuids []string) (map[string]*models.PlayerInfo, error) {
	// Empty list, just return nil.
	if len(puuids) == 0 {
		return nil, nil
	}

	// Get the players.
	var players []models.PlayerInfo
	result := ps.db.Where("puuid IN (?)", puuids).Find(&players)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to make it faster.
	playersMap := make(map[string]*models.PlayerInfo, len(players))
	for i := range players {
		playersMap[players[i].Puuid] = &players[i]
	}

	return playersMap, nil
}

// Get the players with unfetched matches.
func (ps *playerRepository) GetUnfetchedBySubRegions(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	var unfetchedPlayer models.PlayerInfo
	result := ps.db.Where("unfetched_match = ? AND region = ?", true, subRegion).Order("last_match_fetch ASC").First(&unfetchedPlayer)

	// Verify if there is any error.
	if result.Error != nil {
		return nil, result.Error
	}

	return &unfetchedPlayer, nil
}

// Set the date of the last time fetch to the previous + 1 day.
func (ps *playerRepository) SetDelayedLastFetch(playerId uint) error {
	return ps.db.Model(&models.PlayerInfo{}).
		Where("id = ?", playerId).
		UpdateColumn("last_match_fetch", gorm.Expr("last_match_fetch + interval '1 day'")).Error
}

// Set the date of the last time fetch.
func (ps *playerRepository) SetFetched(playerId uint) error {
	return ps.db.Model(&models.PlayerInfo{}).
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
func (ps *playerRepository) UpsertPlayerBatch(players []*models.PlayerInfo) error {
	// In the case of multiple goroutines fetching data, a mutex is needed.
	playerUpsertMutex.Lock()
	defer playerUpsertMutex.Unlock()
	return ps.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "puuid"}, {Name: "region"}},
		DoUpdates: clause.Assignments(map[string]any{
			"profile_icon":      gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.profile_icon ELSE player_infos.profile_icon END"),
			"riot_id_game_name": gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_game_name ELSE player_infos.riot_id_game_name END"),
			"riot_id_tagline":   gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_tagline ELSE player_infos.riot_id_tagline END"),
			"summoner_level":    gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.summoner_level ELSE player_infos.summoner_level END"),
			"updated_at":        gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.updated_at ELSE player_infos.updated_at END"),
		}),
	}).CreateInBatches(&players, 100).Error
}
