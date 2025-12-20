package jobs

import (
	"fmt"
	"goleague/pkg/config"
	"goleague/pkg/database"
	tiervalues "goleague/pkg/riotvalues/tier"
	"log"
	"time"
)

type FetchPriority int

const (
	PriorityLow FetchPriority = iota
	PriorityMediumLow
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

func RecalculateFetchPriority(config *config.Config) error {
	log.Println("Starting fetch priority recalculation")
	startTime := time.Now()

	db, err := database.NewConnection(config.Database.DSN)
	if err != nil {
		return fmt.Errorf("couldn't get database connection: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Truncate the priority table
	log.Println("Truncating priority table...")
	if err := tx.Exec("TRUNCATE TABLE player_fetch_priorities").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to truncate: %w", err)
	}

	log.Println("Dropping index for insert...")
	// Drop the index in order to speed up inserts.
	tx.Exec("DROP INDEX IF EXISTS idx_fetch_priorities_query")

	// Rebuild with single INSERT SELECT
	log.Println("Rebuilding priority table...")
	result := tx.Exec(`
        INSERT INTO player_fetch_priorities (
            player_id, 
            fetch_priority, 
            last_calculated,
            region
        )
        SELECT 
            p.id,
            CASE
                -- Critical: Grandmaster+ with matches in last 3 days
                WHEN r.numeric_score >= ? AND m.last_match > NOW() - INTERVAL '3 days'
                THEN ?::INTEGER
                
                -- High: Master+ OR matches in last 7 days
                WHEN r.numeric_score >= ? AND m.last_match > NOW() - INTERVAL '7 days'
                THEN ?::INTEGER
                
                -- Medium: Diamond+ OR matches in last 14 days
                WHEN r.numeric_score >= ? AND m.last_match > NOW() - INTERVAL '14 days'
                THEN ?::INTEGER
                
				WHEN m.last_match > NOW() - INTERVAL '3 days' OR r.numeric_score >= ?
    			THEN ?::INTEGER

                -- Low: whoever remains
                ELSE ?::INTEGER
            END,
            NOW(),
            p.region
        FROM player_infos p
        LEFT JOIN (
            SELECT DISTINCT ON (player_id) player_id,numeric_score
            FROM rating_entries re
            ORDER BY player_id,fetch_time DESC
        ) r ON r.player_id = p.id
        LEFT JOIN (
    		SELECT ms.player_id,MAX(mi.match_start) as last_match
    		FROM match_stats ms
    		JOIN match_infos mi ON ms.match_id = mi.id
    		GROUP BY ms.player_id
		) m ON m.player_id = p.id;
    `,
		tiervalues.CalculateRank("GRANDMASTER", "I", 0), int(PriorityCritical),
		tiervalues.CalculateRank("MASTER", "I", 0), int(PriorityHigh),
		tiervalues.CalculateRank("DIAMOND", "IV", 0), int(PriorityMedium),
		tiervalues.CalculateRank("MASTER", "I", 0), int(PriorityMediumLow),
		int(PriorityLow),
	)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rebuild priorities: %w", result.Error)
	}

	log.Println("Recreating table index...")
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_fetch_priorities_query ON player_fetch_priorities (region, fetch_priority DESC, player_id)").Error; err != nil {
		log.Println("failed to create index: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Rebuilt %d player priorities in %v", result.RowsAffected, duration)

	var counts struct {
		Low       int64
		MediumLow int64
		Medium    int64
		High      int64
		Critical  int64
	}

	db.Raw("SELECT COUNT(*) FROM player_fetch_priorities WHERE fetch_priority = ?", PriorityLow).Scan(&counts.Low)
	db.Raw("SELECT COUNT(*) FROM player_fetch_priorities WHERE fetch_priority = ?", PriorityMediumLow).Scan(&counts.MediumLow)
	db.Raw("SELECT COUNT(*) FROM player_fetch_priorities WHERE fetch_priority = ?", PriorityMedium).Scan(&counts.Medium)
	db.Raw("SELECT COUNT(*) FROM player_fetch_priorities WHERE fetch_priority = ?", PriorityHigh).Scan(&counts.High)
	db.Raw("SELECT COUNT(*) FROM player_fetch_priorities WHERE fetch_priority = ?", PriorityCritical).Scan(&counts.Critical)

	log.Printf("Distribution - Low: %d, Medium-Low: %d, Medium: %d, High: %d, Critical: %d",
		counts.Low, counts.MediumLow, counts.Medium, counts.High, counts.Critical)

	return nil
}
