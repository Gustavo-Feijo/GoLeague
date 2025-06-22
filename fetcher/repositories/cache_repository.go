package repositories

import (
	"goleague/pkg/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Public Interface.
type CacheRepository interface {
	Setkey(key string, value string) error
}

// cacheRepository structure.
type cacheRepository struct {
	db *gorm.DB
}

// NewCacheRepository creates a new cache repository and return it.
func NewCacheRepository(db *gorm.DB) (CacheRepository, error) {
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
