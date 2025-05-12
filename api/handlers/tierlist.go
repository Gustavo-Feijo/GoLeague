package handlers

import (
	"goleague/api/filters"
	"goleague/api/services"
	tiervalues "goleague/pkg/riotvalues/tier"
	"net/http"

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
	var qp filters.GetQueryParams

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
		filtersMap["tier"] = tiervalues.CalculateRank(tier.(string), rank.(string), 0)
	}

	result, err := h.tierlistService.GetTierlist(filtersMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Now you have all params in one place:
	c.JSON(http.StatusOK, gin.H{"result": result})
}
