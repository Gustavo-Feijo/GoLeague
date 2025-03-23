package models

import (
	"errors"
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/pkg/database"
	"time"

	"gorm.io/gorm"
)

// Create a rating entry for a given player.
type RatingEntry struct {
	ID uint `gorm:"primaryKey"`

	// Reference to the player that has the rating.
	PlayerId uint       `gorm:"primaryKey;autoIncrement:false"`
	Player   PlayerInfo `gorm:"PlayerId"`

	Queue        string `gorm:"type:queue_type"`
	Tier         string `gorm:"type:tier_type"`
	Rank         string `gorm:"type:rank_type"`
	LeaguePoints int
	Wins         int
	Losses       int
	Region       regions.SubRegion `gorm:"type:varchar(5)"`
	FetchTime    time.Time         `gorm:"autoCreateTime"`
}

type RatingService struct {
	db *gorm.DB
}

// Create a rating service.
func CreateRatingService() (*RatingService, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &RatingService{db: db}, nil
}

// Create a rating entry to be saved.
func (rs *RatingService) CreateRatingEntry(
	entry league_fetcher.LeagueEntry,
	playerId uint,
	region regions.SubRegion,
	queue string,
	lastRating *RatingEntry,
) (*RatingEntry, error) {
	insertEntry := &RatingEntry{
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
	rs.db.Model(&PlayerInfo{}).Where("id = ?", playerId).Update("unfetched_match", true)

	// Create the entry.
	err := rs.db.Create(insertEntry).Error
	if err != nil {
		return nil, fmt.Errorf("couldn't create the rating entry for the player %d: %v", playerId, err)
	}

	return insertEntry, nil
}

// Create a list of entries.
func (rs *RatingService) CreateBatchRating(
	entries []RatingEntry,
) error {
	if len(entries) == 0 {
		return nil
	}

	return rs.db.CreateInBatches(&entries, 1000).Error
}

// Get the last rating entry for a given player id.
func (rs *RatingService) GetLastRatingEntryByPlayerIdAndQueue(playerId uint, queue string) (*RatingEntry, error) {
	// Retrieve the latest rating entry for the player
	var rating RatingEntry
	if err := rs.db.Where("player_id = ? AND queue = ?", playerId, queue).Last(&rating).Error; err != nil {
		// If the record was not found, doesn't need to return a error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("no rating entry found: %v", err)
	}

	return &rating, nil
}

// Return a list of ratings.
func (rs *RatingService) GetLastRatingEntryByPlayerIdsAndQueue(ids []uint, queue string) (map[uint]*RatingEntry, error) {

	// Empty list, just return nil.
	if len(ids) == 0 {
		return nil, nil
	}

	// Get the ratings.
	var ratings []RatingEntry
	result := rs.db.Where("player_id IN (?)", ids).Find(&ratings)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to make it faster.
	ratingMap := make(map[uint]*RatingEntry, len(ratings))
	for i := range ratings {
		ratingMap[ratings[i].PlayerId] = &ratings[i]
	}

	return ratingMap, nil
}

// Helper function to determine if a rating needs to be updated
func (rs *RatingService) RatingNeedsUpdate(lastRating *RatingEntry, entry league_fetcher.LeagueEntry) bool {
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
