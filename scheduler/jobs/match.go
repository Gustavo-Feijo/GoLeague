package jobs

import (
	"fmt"
	"goleague/pkg/database"
	"log"
)

// RecalculateMatchRating calculates the average rating from all matches that happened in the last one day.
func RecalculateMatchRating() error {
	log.Println("Starting recalculate match rating.")

	// Create a new connection pool.
	db, err := database.NewConnection()
	if err != nil {
		return fmt.Errorf("couldn't get database connection: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Run the querty to revalidate the match ratings.
	err = db.Exec(`
	WITH avg_scores AS (
	  SELECT
	    mi.id AS match_id,
	    AVG(sub.numeric_score) AS avg_score
	  FROM match_infos mi
	  JOIN LATERAL (
	    SELECT DISTINCT ON (re.player_id)
	      re.player_id,
	      re.numeric_score
	    FROM rating_entries re
	    WHERE re.queue = 'RANKED_SOLO_5x5'
	      AND re.player_id IN (
	        SELECT ms.player_id
	        FROM match_stats ms
	        WHERE ms.match_id = mi.id
	      )
	    ORDER BY re.player_id, ABS(EXTRACT(EPOCH FROM (re.fetch_time - mi.match_start))) ASC
	  ) sub ON TRUE
	  GROUP BY mi.id
	)
	UPDATE match_infos mi
	SET average_rating = avg_scores.avg_score
	FROM avg_scores
	WHERE mi.id = avg_scores.match_id
	AND mi.created_at >= CURRENT_DATE - INTERVAL '1 day'
  	AND mi.created_at < CURRENT_DATE;
`).Error

	if err != nil {
		return fmt.Errorf("update of match scores failed: %v", err)
	}
	log.Println("Finished job")
	return nil
}
