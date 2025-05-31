package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/services"
	"goleague/pkg/redis"
	tiervalues "goleague/pkg/riotvalues/tier"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Tier list handler.
type TierlistHandler struct {
	tierlistService services.TierlistService
}

// Create a new instance of the tierlist handler.
func NewTierlistHandler(service *services.TierlistService) *TierlistHandler {
	return &TierlistHandler{
		tierlistService: *service,
	}
}

// Handler for getting the tierlist.
func (h *TierlistHandler) GetTierlist(c *gin.Context) {
	var qp filters.TierlistQueryParams

	if err := c.ShouldBindQuery(&qp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filtersMap := qp.AsMap()

	if tier, exists := filtersMap["tier"]; exists {
		rank, exists := filtersMap["rank"]
		if !exists {
			rank = "I"
		}
		delete(filtersMap, "rank")
		filtersMap["tier"] = tiervalues.CalculateRank(tier.(string), rank.(string), 0)
	}

	key := getTierlistKey(filtersMap)

	// Get a instance of the memory cache and retrieve the key.
	memCache := cache.GetSimpleCache()
	memCachedData := memCache.Get(key)
	if memCachedData != nil {
		memCachedTierlist := memCachedData.([]*dto.FullTierlist)
		c.JSON(http.StatusOK, gin.H{"result": memCachedTierlist})
		return
	}

	// Create context for fast  redis lookup.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	// Try to get it on redis.
	redisCached, err := redis.GetClient().Get(ctx, key)
	if err == nil {
		// Unmarshal the value to save it as binary on the cache.
		var fulltierlist []*dto.FullTierlist
		json.Unmarshal([]byte(redisCached), &fulltierlist)
		memCache.Set(key, fulltierlist, 15*time.Minute)
		c.JSON(http.StatusOK, gin.H{"result": redisCached})
		return
	}

	result, err := h.tierlistService.GetTierlist(filtersMap)
	if err != nil {
		if err.Error() == "cache failed" {
			c.JSON(http.StatusOK, gin.H{"result": result})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the value in memory and redis.
	memCache.Set(key, result, 15*time.Minute)

	// Marshal it to set on Redis.
	j, err := json.Marshal(result)
	if err == nil {
		redis.GetClient().Set(context.Background(), key, string(j), time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// getTierList generates the cache key.
func getTierlistKey(filters map[string]any) string {
	var builder strings.Builder
	builder.WriteString("tierlist")

	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		filter := fmt.Sprintf(":%s_%v", key, filters[key])
		builder.WriteString(filter)
	}
	return builder.String()
}
