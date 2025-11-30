package repositories

import (
	"goleague/api/filters"
	queuevalues "goleague/pkg/riotvalues/queue"
	tiervalues "goleague/pkg/riotvalues/tier"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// Public Interface.
type TierlistRepository interface {
	GetTierlist(filters *filters.TierlistFilter) ([]*TierlistResult, error)
}

// Tierlist repository structure.
type tierlistRepository struct {
	db *gorm.DB
}

// Create a tierlist repository.
func NewTierlistRepository(db *gorm.DB) TierlistRepository {
	return &tierlistRepository{db: db}
}

type TierlistResult struct {
	BanCount     int
	BanRate      float64
	ChampionId   int
	PickCount    int
	PickRate     float64
	TeamPosition string
	WinRate      float64
}

// GetTierlist is the only necessary function for the tierlist.
// Handle the query building and fetching.
func (ts *tierlistRepository) GetTierlist(filters *filters.TierlistFilter) ([]*TierlistResult, error) {
	var results []*TierlistResult

	// Initialize query parts.
	whereConditions := []string{}
	singleQueryArgs := []any{}

	// Filtering by queue is obrigatory.
	whereConditions = append(whereConditions, "mi.queue_id = ?")

	// If on the filters, get it, else, default to 420 (Ranked Solo Duo).
	defaultQueue := 420
	if queueID := filters.Queue; queueID != 0 {
		defaultQueue = queueID
	}
	singleQueryArgs = append(singleQueryArgs, defaultQueue)

	// Process tier/average_rating filter if provided.
	if avgScore := filters.NumericTier; avgScore != 0 && filters.GetTiersAbove {
		whereConditions = append(whereConditions, "mi.average_rating >= ?")
		singleQueryArgs = append(singleQueryArgs, avgScore)
	}

	if !filters.GetTiersAbove && filters.Tier != "" {
		lower, higher := tiervalues.GetTierLimits(filters.Tier)
		whereConditions = append(whereConditions, "mi.average_rating BETWEEN ? AND ?")
		singleQueryArgs = append(singleQueryArgs, lower, higher)
	}

	// Format the WHERE clause.
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

	// Construct CTE subqueries with proper WHERE clause placement.
	// Should have only 6 possible values, 5 from normal queue and empty for Aram.
	championStatsCTE := `
	WITH champion_stats AS (
		SELECT
			ms.champion_id,
			ms.team_position,
			COUNT(*) as pick_count,
			SUM(ms.win::int) as wins,
			SUM(COUNT(*)) OVER (PARTITION BY ms.team_position) as position_total
		FROM
			match_stats ms
		JOIN 
			match_infos mi ON mi.id = ms.match_id
		` + positionFix + `
		GROUP BY ms.champion_id, ms.team_position
	)
	`

	championBansCTE := `
	, champion_bans AS (
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
	, total_matches AS (
    	SELECT COUNT(*) AS total_match_count
    	FROM match_infos mi 
	` + whereClause +
		`)`

	// Construct the main query.
	mainQuery := `
	SELECT
		cs.pick_count,
    	cs.champion_id,
    	cs.team_position,
    	ROUND((cs.wins * 100.0) / cs.pick_count, 2) AS win_rate,
    	ROUND((cs.pick_count * 100.0) / cs.position_total, 2) AS pick_rate,
    	COALESCE(cb.ban_count, 0) AS ban_count,
    	ROUND((COALESCE(cb.ban_count, 0) * 100.0) / tm.total_match_count, 2) AS ban_rate
	FROM
		champion_stats cs
	LEFT JOIN
		champion_bans cb ON cs.champion_id = cb.champion_id
	CROSS JOIN total_matches tm
	WHERE (cs.pick_count * 100) / cs.position_total > 0.5
	ORDER BY win_rate DESC
    `
	// Combine all parts of the query.
	query := championStatsCTE + championBansCTE + totalMatchesCTE + mainQuery

	// Execute the query with arguments.
	err := ts.db.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
