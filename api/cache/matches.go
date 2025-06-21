package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/api/dto"
	"goleague/pkg/redis"
	"sync"
	"time"
)

const matchPreviewKey = "match:previews:%d"

// Create a redis cache client.
type MatchCache struct {
	redis *redis.RedisClient
}

// Singleton.
var (
	matchInstance *MatchCache
	matchOnce     sync.Once
)

// GetMatchCache gets the instance of the match cache.
func GetMatchCache() *MatchCache {
	matchOnce.Do(func() {
		matchInstance = &MatchCache{
			redis: redis.GetClient(),
		}

	})

	return matchInstance
}

// GetMatchesPreviewByMatchIds retrieves the match preview from a list of match ids.
func (m *MatchCache) GetMatchesPreviewByMatchIds(ctx context.Context, matchIds []uint) ([]dto.MatchPreview, []uint, error) {
	keys := make([]string, len(matchIds))
	for i, matchID := range matchIds {
		keys[i] = fmt.Sprintf(matchPreviewKey, matchID)
	}

	results, err := m.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	var foundMatches []dto.MatchPreview
	var notFoundIds []uint

	for i, result := range results {
		matchID := matchIds[i]

		if result == nil {
			notFoundIds = append(notFoundIds, matchID)
		} else {
			var preview dto.MatchPreview
			if jsonStr, ok := result.(string); ok {
				if err := json.Unmarshal([]byte(jsonStr), &preview); err != nil {
					notFoundIds = append(notFoundIds, matchID)
					continue
				}

				foundMatches = append(foundMatches, preview)
			}
		}
	}

	return foundMatches, notFoundIds, nil
}

// SetMatchPreview saves a given match preview in cache.
func (m *MatchCache) SetMatchPreview(ctx context.Context, preview dto.MatchPreview) error {
	j, err := json.Marshal(preview)
	if err == nil {
		key := fmt.Sprintf(matchPreviewKey, preview.Metadata.InternalId)
		m.redis.Set(context.Background(), key, string(j), time.Hour)
	}
	return err
}
