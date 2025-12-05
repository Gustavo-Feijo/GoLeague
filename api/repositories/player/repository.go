package repositories

import (
	"context"
	"errors"
	"fmt"
	"goleague/api/filters"
	"goleague/pkg/database/models"
	"goleague/pkg/messages"
	"goleague/pkg/regions"
	"strings"
	"time"

	"gorm.io/gorm"
)

const searchLimit = 20

// PlayerRepository is the public interface for accessing the player repository.
type PlayerRepository interface {
	SearchPlayer(ctx context.Context, filters *filters.PlayerSearchFilter) ([]*models.PlayerInfo, error)
	GetPlayerById(ctx context.Context, playerId uint) (*models.PlayerInfo, error)
	GetPlayerIdByNameTagRegion(ctx context.Context, name string, tag string, region string) (uint, error)
	GetPlayerMatchHistoryIds(ctx context.Context, filters *filters.PlayerMatchHistoryFilter) ([]uint, error)
	GetPlayerRatingsById(ctx context.Context, playerId uint) ([]models.RatingEntry, error)
	GetPlayerStats(ctx context.Context, filters *filters.PlayerStatsFilter) ([]RawPlayerStatsStruct, error)
}

// playerRepository repository structure.
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a player repository.
func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

// RawPlayerStatsStruct is the raw data from the player stats analysis.
type RawPlayerStatsStruct struct {
	Matches          int     `gorm:"column:matches"`
	QueueId          int     `gorm:"column:queue_id"`
	TeamPosition     string  `gorm:"column:team_position"`
	ChampionId       int     `gorm:"column:champion_id"`
	WinRate          float32 `gorm:"column:win_rate"`
	AverageKills     float32 `gorm:"column:avg_kills"`
	AverageDeaths    float32 `gorm:"column:avg_deaths"`
	AverageAssists   float32 `gorm:"column:avg_assists"`
	CsPerMin         float32 `gorm:"column:cs_per_min"`
	KDA              float32 `gorm:"column:kda"`
	AggregationLevel string  `gorm:"column:aggregation_level"`
}

// SearchPlayer searchs a given player by it's name, tag and region.
func (ps *playerRepository) SearchPlayer(ctx context.Context, filters *filters.PlayerSearchFilter) ([]*models.PlayerInfo, error) {
	if filters == nil {
		return nil, fmt.Errorf(messages.FiltersNotNil)
	}
	var players []*models.PlayerInfo
	query := ps.db

	name := strings.TrimSpace(filters.Name)
	tag := strings.TrimSpace(filters.Tag)
	region := strings.TrimSpace(filters.Region)

	// Add the search parameters only if the respective value was passed.
	if name != "" {
		query = query.Where("riot_id_game_name LIKE ?", name+"%")
	}

	if tag != "" {
		query = query.Where("riot_id_tagline LIKE ?", tag+"%")
	}

	if region != "" {
		query = query.Where("region = ?", region)
	}

	// Handle empty names and add limit.
	query = query.Where("riot_id_game_name != ''")
	query = query.Limit(searchLimit)
	query = query.Order("riot_id_game_name asc")

	query = query.Model(&models.PlayerInfo{}).
		Select("id, riot_id_game_name, profile_icon, puuid, region, summoner_level, riot_id_tagline")
	if err := query.Find(&players).Error; err != nil {
		return nil, err
	}

	return players, nil
}

// GetPlayerMatchHistoryIds returns the internal ids of the matches that a given player played.
func (ps *playerRepository) GetPlayerMatchHistoryIds(ctx context.Context, filters *filters.PlayerMatchHistoryFilter) ([]uint, error) {
	if filters == nil {
		return nil, fmt.Errorf(messages.FiltersNotNil)
	}
	var ids []uint

	defaultLimit := 10
	playerId := filters.PlayerId
	queueId := filters.Queue

	query := ps.db.Model(&models.MatchInfo{}).
		Select("match_infos.id").
		Joins("JOIN match_stats ms on match_infos.id=ms.match_id").
		Where("ms.player_id = ?", playerId)

	if queueId != 0 {
		query = query.Where("match_infos.queue_id = ?", queueId)
	}

	query = query.Limit(defaultLimit)

	offset := filters.Page
	if offset != 0 {
		query = query.Offset(filters.Page * defaultLimit)
	}

	err := query.Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// GetPlayerStats returns the raw player stats.
func (ps *playerRepository) GetPlayerStats(ctx context.Context, filters *filters.PlayerStatsFilter) ([]RawPlayerStatsStruct, error) {
	if filters == nil {
		return nil, fmt.Errorf(messages.FiltersNotNil)
	}
	var playerStats []RawPlayerStatsStruct
	playerId := filters.PlayerId
	interval := filters.Interval

	// Set a default interval if not provided.
	if interval == 0 {
		interval = 30
	}

	timeThreshold := time.Now().AddDate(0, 0, -interval)

	query := `
		WITH top_champions AS (
		    SELECT champion_id
		    FROM match_stats ms
		    JOIN match_infos mi ON ms.match_id = mi.id
		    WHERE ms.player_id = ? 
		      AND mi.match_start >= ?
		      AND ms.champion_id IS NOT NULL
		    GROUP BY champion_id
		    ORDER BY COUNT(*) DESC
		    LIMIT 10
		),
		base_stats AS (
		    SELECT 
		        COUNT(*) AS matches,
		        mi.queue_id,
		        ms.team_position,
		        ms.champion_id,
		        AVG(ms.win::int) * 100 AS win_rate,
		        AVG(ms.kills) AS avg_kills,
		        AVG(ms.deaths) AS avg_deaths,
		        AVG(ms.assists) AS avg_assists,
		        AVG(ms.total_minions_killed + ms.neutral_minions_killed) / (AVG(mi.match_duration) / 60) AS cs_per_min,
		        CASE
		            WHEN AVG(ms.deaths) = 0 THEN AVG(ms.kills) + AVG(ms.assists)
		            ELSE (AVG(ms.kills) + AVG(ms.assists)) / AVG(ms.deaths)
		        END AS kda,
		        MIN(mi.match_start) AS first_game,
		        MAX(mi.match_start) AS last_game,
		        GROUPING(mi.queue_id) AS is_queue_total,
		        GROUPING(ms.team_position) AS is_position_total,
		        GROUPING(ms.champion_id) AS is_champion_total
		    FROM match_stats ms
		    JOIN match_infos mi ON ms.match_id = mi.id
		    WHERE ms.player_id = ?
		      AND mi.match_start >= ?
			  AND mi.queue_id != 1700
		    GROUP BY GROUPING SETS (
		        (),
		        (mi.queue_id),
		        (ms.team_position),
		        (ms.champion_id),
		        (mi.queue_id, ms.team_position),
		        (mi.queue_id, ms.champion_id)
		    )
		    HAVING (GROUPING(ms.champion_id) = 1 OR ms.champion_id IN (SELECT champion_id FROM top_champions))
		)
		SELECT 
		    matches,
		    COALESCE(queue_id, -1) AS queue_id,
		    COALESCE(team_position, 'ALL') AS team_position,
		    COALESCE(champion_id, -1) AS champion_id,
		    ROUND(win_rate, 2) AS win_rate,
		    ROUND(avg_kills, 2) AS avg_kills,
		    ROUND(avg_deaths, 2) AS avg_deaths,
		    ROUND(avg_assists, 2) AS avg_assists,
		    ROUND(cs_per_min, 2) AS cs_per_min,
		    ROUND(kda, 2) AS kda,
		    CASE 
		        WHEN is_queue_total = 1 AND is_position_total = 1 AND is_champion_total = 1 THEN 'overall'
		        WHEN is_queue_total = 0 AND is_position_total = 1 AND is_champion_total = 1 THEN 'by_queue'
		        WHEN is_queue_total = 1 AND is_position_total = 0 AND is_champion_total = 1 THEN 'by_position'
		        WHEN is_queue_total = 1 AND is_position_total = 1 AND is_champion_total = 0 THEN 'by_champion'
		        WHEN is_queue_total = 0 AND is_position_total = 0 AND is_champion_total = 1 THEN 'by_queue_position'
		        WHEN is_queue_total = 0 AND is_position_total = 1 AND is_champion_total = 0 THEN 'by_queue_champion'
		        ELSE 'detailed'
		    END AS aggregation_level
		FROM base_stats
	`

	if err := ps.db.Raw(query, playerId, timeThreshold, playerId, timeThreshold).Scan(&playerStats).Error; err != nil {
		return nil, err
	}

	return playerStats, nil
}

// GetPlayerIdByNameTagRegion retrieves the id of a given player based on the params.
func (ps *playerRepository) GetPlayerIdByNameTagRegion(ctx context.Context, name string, tag string, region string) (uint, error) {
	var id uint

	formattedRegion := regions.SubRegion(strings.ToUpper(region))

	if err := ps.db.
		Model(&models.PlayerInfo{}).
		Select("id").
		Where("riot_id_game_name = ? AND riot_id_tagline = ? AND region = ?", name, tag, formattedRegion).
		First(&id).Error; err != nil {

		// If the record was not found, doesn't need to return an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("player not found: %v", err)
		}

		// Other database error.
		return 0, fmt.Errorf("could not fetch player id: %v", err)
	}

	return id, nil
}

// GetPlayerInfo returns all the player information.
func (ps *playerRepository) GetPlayerById(ctx context.Context, playerId uint) (*models.PlayerInfo, error) {
	var player models.PlayerInfo
	if err := ps.db.Where(&models.PlayerInfo{ID: playerId}).First(&player).Error; err != nil {
		return nil, fmt.Errorf("couldn't get the player by the ID: %v", err)
	}

	return &player, nil
}

// GetPlayerRatingById returns all the rating information regarding a player.
func (ps *playerRepository) GetPlayerRatingsById(ctx context.Context, playerId uint) ([]models.RatingEntry, error) {
	var ratings []models.RatingEntry
	err := ps.db.Raw(`
    SELECT DISTINCT ON (queue, region) *
    	FROM rating_entries
    	WHERE player_id = ?
    	ORDER BY queue, region, id DESC
	`, playerId).Scan(&ratings).Error

	if err != nil {
		return nil, fmt.Errorf("couldn't get latest ratings: %v", err)
	}

	return ratings, nil
}
