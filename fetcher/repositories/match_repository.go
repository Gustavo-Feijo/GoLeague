package repositories

import (
	"goleague/pkg/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MatchRepository defines the public interface to interact with match data.
type MatchRepository interface {
	CreateMatchBans(bans []*models.MatchBans) error
	CreateMatchInfo(match *models.MatchInfo) error
	CreateMatchStats(stats []*models.MatchStats) error
	GetAlreadyFetchedMatches(riotMatchIDs []string) ([]models.MatchInfo, error)
	SetAverageRating(matchID uint, rating float64) error
	SetFrameInterval(matchID uint, interval int64) error
	SetFullyFetched(matchID uint) error
	SetMatchWinner(matchID uint, winner int) error
}

// matchRepository is the repository instance.
type matchRepository struct {
	db *gorm.DB
}

// NewMatchRepository creates and returns a match repository.
func NewMatchRepository(db *gorm.DB) (MatchRepository, error) {
	return &matchRepository{db: db}, nil
}

// CreateMatchBans inserts the bans in the database. Ignore duplicate picks for a given match.
func (mr *matchRepository) CreateMatchBans(bans []*models.MatchBans) error {
	return mr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "pick_turn"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&bans).Error
}

// CreateMatchInfo creates a match metadata into the database and return the returned error.
func (mr *matchRepository) CreateMatchInfo(match *models.MatchInfo) error {
	return mr.db.Create(&match).Error
}

// CreateMatchStats insert stats entries in the database. Ignores duplicate entries for a player in a given match.
func (mr *matchRepository) CreateMatchStats(stats []*models.MatchStats) error {
	return mr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "player_id"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&stats).Error
}

// GetAlreadyFetchedMatches returns which matches from the received array are already fetched.
func (mr *matchRepository) GetAlreadyFetchedMatches(riotMatchIDs []string) ([]models.MatchInfo, error) {
	const batchSize = 1000
	var allMatches []models.MatchInfo

	for i := 0; i < len(riotMatchIDs); i += batchSize {
		end := min(i+batchSize, len(riotMatchIDs))

		var batchMatches []models.MatchInfo
		result := mr.db.Where("match_id IN (?)", riotMatchIDs[i:end]).Find(&batchMatches)
		if result.Error != nil {
			return nil, result.Error
		}

		allMatches = append(allMatches, batchMatches...)
	}

	return allMatches, nil
}

// SetAverageRating set the average rating for a given match, used for calculating tier data.
func (mr *matchRepository) SetAverageRating(matchID uint, rating float64) error {
	return mr.updateMatchField(matchID, "average_rating", rating)
}

// SetFrameInterval set the frame interval for the match timeline.
func (mr *matchRepository) SetFrameInterval(matchID uint, interval int64) error {
	return mr.updateMatchField(matchID, "frame_interval", interval)
}

// SetFullyFetched set the match as fetched, meaning it doesn't need to be fetched again.
func (mr *matchRepository) SetFullyFetched(matchID uint) error {
	return mr.updateMatchField(matchID, "fully_fetched", true)
}

// SetMatchWinner sets which team has won a given metch.
func (mr *matchRepository) SetMatchWinner(matchID uint, winner int) error {
	return mr.updateMatchField(matchID, "match_winner", winner)
}

// updateMatchField is a generic update helper for a single field in MatchInfo.
func (mr *matchRepository) updateMatchField(matchID uint, field string, value any) error {
	return mr.db.Model(&models.MatchInfo{}).
		Where("id = ?", matchID).
		Update(field, value).Error
}
