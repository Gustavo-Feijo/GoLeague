package repositories

import (
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"goleague/pkg/database/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Public Interface.
type RatingRepository interface {
	CreateBatchRating(entries []models.RatingEntry) error
	CreateRatingEntry(
		entry league_fetcher.LeagueEntry,
		playerId uint,
		region regions.SubRegion,
		queue string,
		lastRating *models.RatingEntry,
	) (*models.RatingEntry, error)
	GetAverageRatingOnMatchByPlayerId(ids []uint, matchID uint, matchTimestamp time.Time, queue string) float64
	GetLastRatingEntryByPlayerIdsAndQueue(ids []uint, queue string) (map[uint]*models.RatingEntry, error)
	RatingNeedsUpdate(lastRating *models.RatingEntry, entry league_fetcher.LeagueEntry) bool
}

// Rating repository.
type ratingRepository struct {
	db *gorm.DB
}

// Create a rating repository.
func NewRatingRepository() (RatingRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &ratingRepository{db: db}, nil
}

// Create a list of entries.
func (rs *ratingRepository) CreateBatchRating(
	entries []models.RatingEntry,
) error {
	if len(entries) == 0 {
		return nil
	}

	return rs.db.CreateInBatches(&entries, 1000).Error
}

// Create a rating entry to be saved.
func (rs *ratingRepository) CreateRatingEntry(
	entry league_fetcher.LeagueEntry,
	playerId uint,
	region regions.SubRegion,
	queue string,
	lastRating *models.RatingEntry,
) (*models.RatingEntry, error) {
	insertEntry := &models.RatingEntry{
		PlayerId:     playerId,
		Region:       region,
		Queue:        queue,
		LeaguePoints: entry.LeaguePoints,
		Losses:       entry.Losses,
		Wins:         entry.Wins,
	}

	// Handle Tier and Rank if they are not nil.
	if entry.Tier != nil {
		insertEntry.Tier = *entry.Tier
	}

	if entry.Rank != nil {
		insertEntry.Rank = *entry.Rank
	} else {
		// If it's high elo, it will be nil, just set the ranking as I.
		insertEntry.Rank = "I"
	}

	// If nothing changed, just return nil at both.
	if lastRating != nil &&
		lastRating.Tier == insertEntry.Tier &&
		lastRating.Rank == insertEntry.Rank &&
		lastRating.LeaguePoints == insertEntry.LeaguePoints &&
		lastRating.Losses == insertEntry.Losses &&
		lastRating.Wins == insertEntry.Wins {
		return nil, nil
	}

	// Set the player to be fetched by the matches queue, since something changed.
	rs.db.Model(&models.PlayerInfo{}).Where("id = ?", playerId).Update("unfetched_match", true)

	// Create the entry.
	err := rs.db.Create(insertEntry).Error
	if err != nil {
		return nil, fmt.Errorf("couldn't create the rating entry for the player %d: %v", playerId, err)
	}

	return insertEntry, nil
}

// Get the closest rating entry to a given match timestamp.
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

	// Coreate the args.
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, matchID)
	for _, id := range ids {
		args = append(args, id)
	}

	// Run query
	type EntryResult struct {
		AvgScore float64
	}

	var results EntryResult
	rs.db.Raw(query, args...).Scan(&results)

	return results.AvgScore
}

// Return a map of ratings by the playerID.
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

// Helper function to determine if a rating needs to be updated
func (rs *ratingRepository) RatingNeedsUpdate(lastRating *models.RatingEntry, entry league_fetcher.LeagueEntry) bool {
	if lastRating == nil {
		return true // No previous rating, needs update
	}

	// Check if any important fields have changed
	if lastRating.LeaguePoints != entry.LeaguePoints ||
		lastRating.Wins != entry.Wins ||
		lastRating.Losses != entry.Losses ||
		(entry.Tier != nil && lastRating.Tier != *entry.Tier) ||
		(entry.Tier != nil && lastRating.Rank != *entry.Rank) {
		return true
	}

	return false
}
