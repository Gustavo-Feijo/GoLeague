package repositories

import (
	"fmt"
	"goleague/pkg/database"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Public Interface.
type MatchRepository interface {
	CreateMatchBans(bans []*models.MatchBans) error
	CreateMatchInfo(match *models.MatchInfo) error
	CreateMatchStats(stats []*models.MatchStats) error
	GetAlreadyFetchedMatches(riotMatchIDs []string) ([]models.MatchInfo, error)
	SetFrameInterval(matchID uint, interval int64) error
	SetFullyFetched(matchID uint) error
	SetMatchWinner(matchID uint, winner int) error
}

// Match repository structure.
type matchRepository struct {
	db *gorm.DB
}

// Create a match repository.
func NewMatchRepository() (MatchRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &matchRepository{db: db}, nil
}

// Simply create the bans of a given match.
func (mr *matchRepository) CreateMatchBans(bans []*models.MatchBans) error {
	return mr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "pick_turn"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&bans).Error
}

// Simply create a match information.
func (mr *matchRepository) CreateMatchInfo(match *models.MatchInfo) error {
	return mr.db.Create(&match).Error
}

// Simply create the stats of a given match.
func (mr *matchRepository) CreateMatchStats(stats []*models.MatchStats) error {
	return mr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "player_id"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&stats).Error
}

// Get all the already existing matches.
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

// Set the frame interval for the match timeline.
func (mr *matchRepository) SetFrameInterval(matchID uint, interval int64) error {
	return mr.db.Model(&models.MatchInfo{}).
		Where("id = ?", matchID).
		Update("frame_interval", interval).Error
}

// Set a match as fully fetched.
func (mr *matchRepository) SetFullyFetched(matchID uint) error {
	return mr.db.Model(&models.MatchInfo{}).
		Where("id = ?", matchID).
		Update("fully_fetched", true).Error
}

// Set the match winner team id.
func (mr *matchRepository) SetMatchWinner(matchID uint, winner int) error {
	return mr.db.Model(&models.MatchInfo{}).
		Where("id = ?", matchID).
		Update("match_winner", winner).Error
}
