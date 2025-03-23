package models

import (
	"fmt"
	match_fetcher "goleague/fetcher/data/match"
	"goleague/pkg/database"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Database model for the match information.
type MatchInfo struct {
	ID             uint   `gorm:"primaryKey"`
	GameVersion    string `gorm:"type:varchar(20)"`
	MatchId        string `gorm:"type:varchar(20);uniqueIndex"`
	MatchStart     time.Time
	MatchDuration  int
	MatchWinner    bool
	MatchSurrender bool
	MatchRemake    bool
	FullyFetched   bool
	QueueId        int `gorm:"index"`
}

// Database model for saving a player perfomance in a given match.
type MatchStats struct {
	// Composite primary key, since the same player can't be twice on the same match.
	MatchId  uint `gorm:"primaryKey;autoIncrement:false"`
	PlayerId uint `gorm:"primaryKey;autoIncrement:false"`

	// Foreign keys.
	Match  MatchInfo  `gorm:"MatchId"`
	Player PlayerInfo `gorm:"PlayerId"`

	// Embedded match stats.
	PlayerData match_fetcher.MatchPlayer `gorm:"embedded"`
}

// Database model for saving the match bans.
type MatchBans struct {
	MatchId    uint `gorm:"primaryKey;autoIncrement:false"`
	PickTurn   int  `gorm:"primaryKey;autoIncrement:false"`
	ChampionId int
}

// Match service structure.
type MatchService struct {
	db *gorm.DB
}

// Create a match service.
func CreateMatchService() (*MatchService, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &MatchService{db: db}, nil
}

// Simply craete the bans of a given match.
func (ms *MatchService) CreateMatchBans(bans []*MatchBans) error {
	return ms.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "pick_turn"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&bans).Error
}

// Simply create a match information.
func (ms *MatchService) CreateMatchInfo(match *MatchInfo) error {
	return ms.db.Create(&match).Error
}

// Simply craete the stats of a given match.
func (ms *MatchService) CreateMatchStats(stats []*MatchStats) error {
	return ms.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "player_id"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&stats).Error
}

// Get all the already existing matches.
func (ms *MatchService) GetAlreadyFetchedMatches(ids []string) ([]MatchInfo, error) {
	const batchSize = 1000
	var allMatches []MatchInfo

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		var batchMatches []MatchInfo
		result := ms.db.Where("match_id IN (?)", ids[i:end]).Find(&batchMatches)
		if result.Error != nil {
			return nil, result.Error
		}

		allMatches = append(allMatches, batchMatches...)
	}

	return allMatches, nil
}

// Set a match as fully fetched.
func (ms *MatchService) SetFullyFetched(match_id uint) error {
	return ms.db.Model(&MatchInfo{}).
		Where("id = ?", match_id).
		Update("fully_fetched", true).Error
}
