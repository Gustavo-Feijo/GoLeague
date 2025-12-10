package handlers

import (
	"goleague/api/cache"
	"goleague/api/filters"
	tierlistservice "goleague/api/services/tierlist"
	"goleague/pkg/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Tier list handler.
type TierlistHandler struct {
	memCache        cache.MemCache
	redis           *redis.RedisClient
	tierlistService *tierlistservice.TierlistService
}

type TierlistHandlerDependencies struct {
	TierlistService *tierlistservice.TierlistService
}

// Create a new instance of the tierlist handler.
func NewTierlistHandler(deps *TierlistHandlerDependencies) *TierlistHandler {
	return &TierlistHandler{
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

	filters := filters.NewTierlistFilter(qp)

	result, err := h.tierlistService.GetTierlist(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}
