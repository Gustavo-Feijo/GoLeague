package jobs

import (
	"goleague/pkg/database"
	"log"
)

func RecalculateMatchRating() {
	log.Println("Starting recalculate match rating.")
	db, err := database.GetConnection()
	if err != nil {
		log.Printf("Couldn't get the database connection: %v", err)
	}

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
	WHERE mi.id = avg_scores.match_id;
`).Error

	if err != nil {
		log.Printf("Update of match scores failed: %v", err)
	}
	log.Println("Finished job")
}
