package handlers

import (
	"goleague/api/cache"
	"goleague/api/filters"
	"goleague/api/services"
	"goleague/pkg/redis"
	tiervalues "goleague/pkg/riotvalues/tier"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Tier list handler.
type TierlistHandler struct {
	memCache        *cache.MemCache
	redis           *redis.RedisClient
	tierlistService *services.TierlistService
}

type TierlistHandlerDependencies struct {
	MemCache        *cache.MemCache
	Redis           *redis.RedisClient
	TierlistService *services.TierlistService
}

// Create a new instance of the tierlist handler.
func NewTierlistHandler(deps *TierlistHandlerDependencies) *TierlistHandler {
	return &TierlistHandler{
		memCache:        deps.MemCache,
		redis:           deps.Redis,
		tierlistService: deps.TierlistService,
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

	result, err := h.tierlistService.GetTierlist(filtersMap)
	if err != nil {
		if err.Error() == "cache failed" {
			c.JSON(http.StatusOK, gin.H{"result": result})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}
