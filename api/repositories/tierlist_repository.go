package repositories

import (
	"fmt"
	"goleague/pkg/database"
	queuevalues "goleague/pkg/riotvalues/queue"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// Public Interface.
type TierlistRepository interface {
	GetTierlist(filters map[string]any) ([]*TierlistResult, error)
}

// Tierlist repository structure.
type tierlistRepository struct {
	db *gorm.DB
}

// Result of a tierlist fetch.
type TierlistResult struct {
	BanCount     int
	ChampionId   int
	Pickcount    int
	Pickrate     float64
	TeamPosition string
	Winrate      float64
}

// Create a tierlist repository.
func NewTierlistRepository() (TierlistRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &tierlistRepository{db: db}, nil
}

// GetTierlist is the only necessary function for the tierlist.
// Handle the query building and fetching.
func (ts *tierlistRepository) GetTierlist(filters map[string]any) ([]*TierlistResult, error) {
	var results []*TierlistResult

	// Initialize query parts
	whereConditions := []string{}
	singleQueryArgs := []any{}

	// Filtering by queue is obrigatory.
	whereConditions = append(whereConditions, "mi.queue_id = ?")

	// If on the filters, get it, else, default to 420 (Ranked Solo Duo).
	defaultQueue := 420
	if queueID, ok := filters["queue"].(int); ok {
		defaultQueue = queueID
	}
	singleQueryArgs = append(singleQueryArgs, defaultQueue)

	// Process tier/average_rating filter if provided
	if avgScore, ok := filters["tier"].(int); ok {
		whereConditions = append(whereConditions, "mi.average_rating >= ?")
		singleQueryArgs = append(singleQueryArgs, avgScore)
	}

	// Format the WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Sometimes in queues that must have positions, the RIOT API returns invalid data.
	// The team position of some players can be '', leading to bad data.
	positionFix := whereClause
	if slices.Contains(queuevalues.QueuesWithPositions, defaultQueue) {
		if len(whereClause) > 0 {
			positionFix += " AND ms.team_position != ''"
		} else {
			positionFix = "WHERE ms.team_position != ''"
		}
	}

	args := []any{}
	// Append the single query args 3 times.
	// Three for the CTEs and one for the main query.
	args = append(args, singleQueryArgs...)
	args = append(args, singleQueryArgs...)
	args = append(args, singleQueryArgs...)
	args = append(args, singleQueryArgs...)

	// Construct CTE subqueries with proper WHERE clause placement
	// Should have only 6 possible values, 5 from normal queue and empty for Aram.
	positionCountsCTE := `
	WITH positionCounts AS (
		SELECT
			team_position,
			COUNT(*) AS positionCount
		FROM
			match_stats ms
		JOIN
			match_infos mi ON mi.id = ms.match_id
		` + whereClause + ` GROUP BY
			team_position
	)`

	championBansCTE := `
	, championBans AS (
		SELECT
			mb.champion_id,
			COUNT(*) AS ban_count
		FROM
			match_bans mb
		JOIN
			match_infos mi ON mi.id = mb.match_id
		` + whereClause + ` GROUP BY
			mb.champion_id
	)`

	totalMatchesCTE := `
	, totalMatches AS (
    SELECT COUNT(*) AS total
    FROM match_infos mi ` + whereClause + `)`

	// Construct the main query
	mainQuery := `
	SELECT
		COUNT(*) AS pickCount,
		ms.champion_id,
		ms.team_position,
		AVG(ms.win::int) * 100 AS winRate,
		(COUNT(*) * 100.0) / pc.positionCount AS pickRate,
		COALESCE(cb.ban_count, 0) AS banCount,
		(COALESCE(cb.ban_count, 0) * 100.0) / tm.total AS banRate
	FROM
		match_stats ms
	JOIN
		match_infos mi ON mi.id = ms.match_id
	JOIN
		positionCounts pc ON ms.team_position = pc.team_position
	LEFT JOIN
		championBans cb ON ms.champion_id = cb.champion_id
	JOIN
    	totalMatches tm ON TRUE
	` + positionFix + ` GROUP BY
		ms.champion_id,
		ms.team_position,
		pc.positionCount,
		cb.ban_count,
		tm.total
	HAVING
		(COUNT(*) * 100.0) / pc.positionCount > 0.5
	ORDER BY winRate desc
    `
	// Combine all parts of the query
	query := positionCountsCTE + championBansCTE + totalMatchesCTE + mainQuery

	// Execute the query with arguments
	err := ts.db.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
