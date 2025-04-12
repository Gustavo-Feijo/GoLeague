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
	MatchWinner    int
	MatchSurrender bool
	MatchRemake    bool
	FrameInterval  int64
	FullyFetched   bool
	QueueId        int `gorm:"index"`
}

// Database model for saving a player perfomance in a given match.
type MatchStats struct {
	// Ids and identifiers for the match stats.
	ID       uint64 `gorm:"primaryKey"`
	MatchId  uint   `gorm:"not null;index:idx_match_player,unique"`
	PlayerId uint   `gorm:"not null;index:idx_match_player,unique"`

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

// Simply create the bans of a given match.
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

// Simply create the stats of a given match.
func (ms *MatchService) CreateMatchStats(stats []*MatchStats) error {
	return ms.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "player_id"}}, // Use the composite key columns
		DoNothing: true,
	}).Create(&stats).Error
}

// Get all the already existing matches.
func (ms *MatchService) GetAlreadyFetchedMatches(riot_match_ids []string) ([]MatchInfo, error) {
	const batchSize = 1000
	var allMatches []MatchInfo

	for i := 0; i < len(riot_match_ids); i += batchSize {
		end := min(i+batchSize, len(riot_match_ids))

		var batchMatches []MatchInfo
		result := ms.db.Where("match_id IN (?)", riot_match_ids[i:end]).Find(&batchMatches)
		if result.Error != nil {
			return nil, result.Error
		}

		allMatches = append(allMatches, batchMatches...)
	}

	return allMatches, nil
}

// Set the frame interval for the match timeline.
func (ms *MatchService) SetFrameInterval(match_id uint, interval int64) error {
	return ms.db.Model(&MatchInfo{}).
		Where("id = ?", match_id).
		Update("frame_interval", interval).Error
}

// Set a match as fully fetched.
func (ms *MatchService) SetFullyFetched(match_id uint) error {
	return ms.db.Model(&MatchInfo{}).
		Where("id = ?", match_id).
		Update("fully_fetched", true).Error
}

// Set the match winner team id.
func (ms *MatchService) SetMatchWinner(match_id uint, winner int) error {
	return ms.db.Model(&MatchInfo{}).
		Where("id = ?", match_id).
		Update("match_winner", winner).Error
}
