package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/api/dto"
	"goleague/pkg/redis"
	"time"
)

// Default key for the match previews.
const (
	matchPreviewCacheDuration = time.Hour
	matchPreviewKey           = "match:previews:%d"
)

// MatchCache is the public interface for accessing the player repository.
type MatchCache interface {
	GetMatchesPreviewByMatchIds(ctx context.Context, matchIds []uint) ([]dto.MatchPreview, []uint, error)
	SetMatchPreview(ctx context.Context, preview dto.MatchPreview) error
}

// Create a redis cache client.
type matchCache struct {
	redis *redis.RedisClient
}

// NewMatchCache creates a new  instance of the match redis client.
func NewMatchCache(redis *redis.RedisClient) MatchCache {
	mc := &matchCache{
		redis: redis,
	}

	return mc
}

// GetMatchesPreviewByMatchIds retrieves the match preview from a list of match ids.
func (mc *matchCache) GetMatchesPreviewByMatchIds(ctx context.Context, matchIds []uint) ([]dto.MatchPreview, []uint, error) {
	keys := make([]string, len(matchIds))
	for i, matchID := range matchIds {
		keys[i] = fmt.Sprintf(matchPreviewKey, matchID)
	}
	results, err := mc.redis.MGet(ctx, keys...).Result()
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
func (mc *matchCache) SetMatchPreview(ctx context.Context, preview dto.MatchPreview) error {
	j, err := json.Marshal(preview)
	if err == nil {
		key := fmt.Sprintf(matchPreviewKey, preview.Metadata.InternalId)
		mc.redis.Set(context.Background(), key, string(j), matchPreviewCacheDuration)
	}
	return err
}
