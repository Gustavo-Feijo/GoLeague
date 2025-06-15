package repositories

import (
	"fmt"
	"goleague/api/dto"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"goleague/pkg/database/models"
	"strings"

	"gorm.io/gorm"
)

const searchLimit = 20

// PlayerRepository is the public interface for accessing the player repository.
type PlayerRepository interface {
	SearchPlayer(filters map[string]any) ([]*dto.PlayerSearch, error)
	GetPlayerMatchHistory(filters map[string]any) error
	GetPlayerIdByNameTagRegion(name string, tag string, region string) (*dto.PlayerId, error)
}

// playerRepository repository structure.
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a player repository.
func NewPlayerRepository() (PlayerRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
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

func (ps *playerRepository) GetPlayerMatchHistory(filters map[string]any) error {
	return nil
}

// GetPlayerIdByNameTagRegion retrieves the id of a given player based on the params.
func (ps *playerRepository) GetPlayerIdByNameTagRegion(name string, tag string, region string) (*dto.PlayerId, error) {
	var result *dto.PlayerId

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
		return &dto.PlayerId{ID: 0}, err
	}

	return result, nil
}
