package models

// Database model for saving the cache keys.
// Used as fallback in case the Redis is down.
// Should be use only as a last resort to fetch and store in memory, since it will be too slow.
type CacheBackup struct {
	CacheKey   string `gorm:"primaryKey;autoIncrement:false"`
	CacheValue string `gorm:"type:jsonb"`
}
