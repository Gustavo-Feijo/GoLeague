package repositories

import (
	"fmt"
	leaguefetcher "goleague/fetcher/data/league"
	"goleague/pkg/database"
	"goleague/pkg/database/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RatingRepository is the public interface for handling rating changes.
type RatingRepository interface {
	CreateBatchRating(entries []models.RatingEntry) error
	GetAverageRatingOnMatchByPlayerId(ids []uint, matchID uint, matchTimestamp time.Time, queue string) float64
	GetLastRatingEntryByPlayerIdsAndQueue(ids []uint, queue string) (map[uint]*models.RatingEntry, error)
	RatingNeedsUpdate(lastRating *models.RatingEntry, entry leaguefetcher.LeagueEntry) bool
}

// ratingRepository is the repository instance.
type ratingRepository struct {
	db *gorm.DB
}

// NewRatingRepository creates a new repository and return it.
func NewRatingRepository() (RatingRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &ratingRepository{db: db}, nil
}

// CreateBatchRating creates multiple rating entries at a time.
func (rs *ratingRepository) CreateBatchRating(
	entries []models.RatingEntry,
) error {
	if len(entries) == 0 {
		return nil
	}

	return rs.db.CreateInBatches(&entries, 1000).Error
}

// GetAverageRatingOnMatchByPlayerId gets the average rating of the match.
func (rs *ratingRepository) GetAverageRatingOnMatchByPlayerId(ids []uint, matchID uint, matchTimestamp time.Time, queue string) float64 {

	// Build placeholders for the clause.
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")

	query := fmt.Sprintf(`
	SELECT AVG(sub.numeric_score) AS avg_score
	FROM (
    SELECT DISTINCT ON (re.player_id)
        re.numeric_score
    FROM rating_entries re
    JOIN match_infos mi ON mi.id = ?
    WHERE re.player_id IN (%s)
    ORDER BY re.player_id, ABS(EXTRACT(EPOCH FROM (re.fetch_time - mi.match_start))) ASC
	) AS sub;
	`, placeholders)

	args := make([]any, 0, len(ids)+1)
	args = append(args, matchID)
	for _, id := range ids {
		args = append(args, id)
	}

	type EntryResult struct {
		AvgScore float64
	}

	var results EntryResult
	rs.db.Raw(query, args...).Scan(&results)

	return results.AvgScore
}

// GetLastRatingEntryByPlayerIdsAndQueue returns a map of ratings by the playerID.
func (rs *ratingRepository) GetLastRatingEntryByPlayerIdsAndQueue(ids []uint, queue string) (map[uint]*models.RatingEntry, error) {

	// Empty list, just return nil.
	if len(ids) == 0 {
		return nil, nil
	}

	// Get the ratings.
	var ratings []models.RatingEntry
	result := rs.db.Raw(`
        SELECT DISTINCT ON (player_id) * 
        FROM rating_entries
        WHERE player_id IN (?)
        AND queue = ?
        ORDER BY player_id, fetch_time DESC
    `, ids, queue).Scan(&ratings)

	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to make it faster.
	ratingMap := make(map[uint]*models.RatingEntry, len(ratings))
	for i := range ratings {
		ratingMap[ratings[i].PlayerId] = &ratings[i]
	}

	return ratingMap, nil
}

// RatingNeedsUpdate is a function to determine if a rating needs to be updated.
func (rs *ratingRepository) RatingNeedsUpdate(lastRating *models.RatingEntry, entry leaguefetcher.LeagueEntry) bool {
	if lastRating == nil {
		return true // No previous rating, needs update
	}

	// Check if any important fields have changed.
	if lastRating.LeaguePoints != entry.LeaguePoints ||
		lastRating.Wins != entry.Wins ||
		lastRating.Losses != entry.Losses ||
		(entry.Tier != nil && lastRating.Tier != *entry.Tier) ||
		(entry.Tier != nil && lastRating.Rank != *entry.Rank) {
		return true
	}

	return false
}
