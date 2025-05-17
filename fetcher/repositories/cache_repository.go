package repositories

import (
	"fmt"
	"goleague/pkg/database"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Public Interface.
type CacheRepository interface {
	Setkey(key string, value string) error
}

// Match repository structure.
type cacheRepository struct {
	db *gorm.DB
}

// Create a match repository.
func NewCacheRepository() (CacheRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &cacheRepository{db: db}, nil
}

// SetKey sets the given key value.
// Should be used as a Redis fallback.
func (cr *cacheRepository) Setkey(key string, value string) error {
	cacheEntry := &models.CacheBackup{
		CacheKey:   key,
		CacheValue: value,
	}

	// Upsert the cache  key.
	return cr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cache_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"cache_value"}),
	}).Create(cacheEntry).Error
}
