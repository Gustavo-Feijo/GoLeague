package models

// CacheBackup saves keys and values that will be in cache.
// Used as fallback in case the Redis is down.
// Should be use only as a last resort to fetch and store in memory, since it will be too slow.
type CacheBackup struct {
	CacheKey   string `gorm:"primaryKey;autoIncrement:false"`
	CacheValue string `gorm:"type:jsonb"`
}
