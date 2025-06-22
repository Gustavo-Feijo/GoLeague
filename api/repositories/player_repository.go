package repositories

import (
	"goleague/api/dto"
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"strings"

	"gorm.io/gorm"
)

const searchLimit = 20

// PlayerRepository is the public interface for accessing the player repository.
type PlayerRepository interface {
	SearchPlayer(filters map[string]any) ([]*dto.PlayerSearch, error)
	GetPlayerMatchHistoryIds(filters map[string]any) ([]uint, error)
	GetPlayerIdByNameTagRegion(name string, tag string, region string) (uint, error)
}

// playerRepository repository structure.
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a player repository.
func NewPlayerRepository(db *gorm.DB) (PlayerRepository, error) {
	return &playerRepository{db: db}, nil
}

// SearchPlayer searchs a given player by it's name, tag and region.
func (ps *playerRepository) SearchPlayer(filters map[string]any) ([]*dto.PlayerSearch, error) {
	var players []*dto.PlayerSearch
	query := ps.db

	name := strings.TrimSpace(filters["name"].(string))
	tag := strings.TrimSpace(filters["tag"].(string))
	region := strings.TrimSpace(filters["region"].(string))

	// Add the search parameters only if the respective value was passed.
	if name != "" {
		query = query.Where("LOWER(riot_id_game_name) LIKE LOWER(?)", name+"%")
	}

	if tag != "" {
		query = query.Where("LOWER(riot_id_tagline) LIKE LOWER(?)", tag+"%")
	}

	if region != "" {
		query = query.Where("LOWER(region) = LOWER(?)", region)
	}

	// Handle empty names and add limit.
	query = query.Where("riot_id_game_name != ''")
	query = query.Limit(searchLimit)
	query = query.Order("riot_id_game_name asc")

	query = query.Model(&models.PlayerInfo{}).
		Select("id, riot_id_game_name as name, profile_icon, puuid, region, summoner_level, riot_id_tagline as tag")
	if err := query.Find(&players).Error; err != nil {
		return nil, err
	}

	return players, nil
}

// GetPlayerMatchHistoryIds returns the internal ids of the matches that a given player played.
func (ps *playerRepository) GetPlayerMatchHistoryIds(filters map[string]any) ([]uint, error) {
	var ids []uint

	defaultLimit := 10
	playerId := filters["playerId"]
	queueId, queueSetted := filters["queue"]

	query := ps.db.Model(&models.MatchInfo{}).
		Select("match_infos.id").
		Joins("JOIN match_stats ms on match_infos.id=ms.match_id").
		Where("ms.player_id = ?", playerId)

	if queueSetted && queueId != 0 {
		query = query.Where("match_infos.queue_id = ?", queueId)
	}

	query = query.Limit(defaultLimit)

	offset, hasOffset := filters["page"]
	if hasOffset {
		if page, ok := offset.(int); ok {
			query = query.Offset(page * defaultLimit)
		}
	}

	err := query.Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// GetPlayerIdByNameTagRegion retrieves the id of a given player based on the params.
func (ps *playerRepository) GetPlayerIdByNameTagRegion(name string, tag string, region string) (uint, error) {
	var result uint

	formattedRegion := regions.SubRegion(strings.ToUpper(region))
	err := ps.db.
		Model(&models.PlayerInfo{}).
		Select("id").
		Where(&models.PlayerInfo{
			RiotIdGameName: name,
			RiotIdTagline:  tag,
			Region:         formattedRegion,
		}).First(&result).Error
	if err != nil {
		return 0, err
	}

	return result, nil
}
