package repositories

import (
	"fmt"
	"goleague/pkg/database"
	"goleague/pkg/database/models"

	"gorm.io/gorm"
)

// Public Interface.
type CacheRepository interface {
	GetKey(key string) (string, error)
	GetByPrefix(prefix string) ([]*models.CacheBackup, error)
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

// GetKey gets the given key value.
// Should be used as a Redis fallback.
func (cr *cacheRepository) GetKey(key string) (string, error) {
	cacheEntry := &models.CacheBackup{
		CacheKey: key,
	}

	// Upsert the cache  key.
	err := cr.db.Where(&cacheEntry).First(&cacheEntry).Error
	if err != nil {
		return "", err
	}

	return cacheEntry.CacheValue, nil
}

// SetKey sets the given key value.
// Should be used as a Redis fallback.
func (cr *cacheRepository) GetByPrefix(prefix string) ([]*models.CacheBackup, error) {
	var cacheEntries []*models.CacheBackup

	if err := cr.db.Where("cache_key LIKE ?", prefix+"%").Find(&cacheEntries).Error; err != nil {
		return nil, err
	}

	return cacheEntries, nil
}
