package repositories

import (
	"errors"
	"fmt"
	"goleague/pkg/database/models"
	"goleague/pkg/regions"
	"log"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PlayerRepository defines the public interface for handling player related data.
type PlayerRepository interface {
	CreatePlayersBatch(players []*models.PlayerInfo) error
	GetPlayerByNameTagRegion(gameName string, gameTag string, region string) (*models.PlayerInfo, error)
	GetPlayerByPuuid(puuid string) (*models.PlayerInfo, error)
	GetPlayersByPuuids(puuids []string) (map[string]*models.PlayerInfo, error)
	GetNextFetchPlayerBySubRegion(subRegion regions.SubRegion) (*models.PlayerInfo, error)
	SetDelayedLastFetch(playerId uint) error
	SetFetched(playerId uint) error
	UpsertPlayerBatch(players []*models.PlayerInfo) error
}

// playerRepository is the repository instance.
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates and return the player repository.
func NewPlayerRepository(db *gorm.DB) (PlayerRepository, error) {
	return &playerRepository{db: db}, nil
}

// CreatePlayersBatch creates multiple players in batches of 1000.
func (ps *playerRepository) CreatePlayersBatch(players []*models.PlayerInfo) error {
	if len(players) == 0 {
		return nil
	}

	// Creates in batches of 1000.
	return ps.db.CreateInBatches(&players, 1000).Error
}

// GetPlayerByNameTagRegion returns a given player by his gamename, tag and region.
func (ps *playerRepository) GetPlayerByNameTagRegion(gameName string, gameTag string, region string) (*models.PlayerInfo, error) {
	var player models.PlayerInfo
	if err := ps.db.
		Where("riot_id_game_name = ? AND riot_id_tagline = ? AND region = ?", gameName, gameTag, region).
		First(&player).Error; err != nil {

		// If the record was not found, doesn't need to return a error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Other database error.
		return nil, fmt.Errorf("player not found: %v", err)
	}

	return &player, nil
}

// GetPlayerByPuuid returns a given player by his PUUID.
func (ps *playerRepository) GetPlayerByPuuid(puuid string) (*models.PlayerInfo, error) {
	// Retrieve player by PUUID.
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

// GetPlayersByPuuids returns a list of players by a list of passed PUUIDs.
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

// GetNextFetchPlayerBySubRegion returns a single player from a region, getting the next player with pending matches ordered by fetch priority.
func (ps *playerRepository) GetNextFetchPlayerBySubRegion(subRegion regions.SubRegion) (*models.PlayerInfo, error) {
	var unfetchedPlayer models.PlayerInfo
	result := ps.db.
		Joins("JOIN player_fetch_priorities pfp ON pfp.player_id = player_infos.id").
		Where("player_infos.unfetched_match = ?", true).
		Where("pfp.region = ?", subRegion).
		Order("pfp.fetch_priority DESC").
		Order("player_infos.last_match_fetch ASC").
		First(&unfetchedPlayer)

	if result.Error == nil {
		return &unfetchedPlayer, nil
	}

	// If no rows found (priorities table might be empty), use fallback.
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("No prioritized players found for region %s, using fallback query", subRegion)

		fallbackResult := ps.db.
			Where("player_infos.unfetched_match = ?", true).
			Where("player_infos.region = ?", subRegion).
			Order("player_infos.last_match_fetch ASC").
			First(&unfetchedPlayer)

		if fallbackResult.Error != nil {
			return nil, fallbackResult.Error
		}

		return &unfetchedPlayer, nil
	}

	return nil, result.Error
}

// SetDelayedLastFetch set the date of the last time fetch to the previous + 1 day.
func (ps *playerRepository) SetDelayedLastFetch(playerId uint) error {
	return ps.db.Model(&models.PlayerInfo{}).
		Where("id = ?", playerId).
		UpdateColumn("last_match_fetch", gorm.Expr("last_match_fetch + interval '1 day'")).Error
}

// SetFetched set the player as fetched and store the date where it was fetched.
func (ps *playerRepository) SetFetched(playerId uint) error {
	return ps.db.Model(&models.PlayerInfo{}).
		Where("id = ?", playerId).
		Updates(
			map[string]any{
				"last_match_fetch": time.Now().UTC(),
				"unfetched_match":  false,
			},
		).Error
}

// UpsertPlayerBatch upsert multiple players with retry.
// The retry is due to the possibility of a deadlock.
// The deadlock could be caused by the main region updating a given player or working with Goroutines for fetching matches.
func (ps *playerRepository) UpsertPlayerBatch(players []*models.PlayerInfo) error {
	const maxRetries = 3

	// Sort to improve deadlock treatment.
	sort.Slice(players, func(i, j int) bool {
		if players[i].Puuid != players[j].Puuid {
			return players[i].Puuid < players[j].Puuid
		}
		return players[i].Region < players[j].Region
	})

	for range maxRetries {
		err := ps.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "puuid"}, {Name: "region"}},
			DoUpdates: clause.Assignments(map[string]any{
				"profile_icon":      gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.profile_icon ELSE player_infos.profile_icon END"),
				"riot_id_game_name": gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_game_name ELSE player_infos.riot_id_game_name END"),
				"riot_id_tagline":   gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.riot_id_tagline ELSE player_infos.riot_id_tagline END"),
				"summoner_level":    gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.summoner_level ELSE player_infos.summoner_level END"),
				"updated_at":        gorm.Expr("CASE WHEN player_infos.updated_at < excluded.updated_at THEN excluded.updated_at ELSE player_infos.updated_at END"),
			}),
		}).CreateInBatches(&players, 100).Error

		if err == nil {
			return nil
		}

		if !isDeadlockError(err) {
			return err
		}
	}
	return errors.New("too many retries on deadlock")
}

// isDeadlockError verify if there is a deadlock in the database.
func isDeadlockError(err error) bool {
	return strings.Contains(err.Error(), "deadlock")
}
