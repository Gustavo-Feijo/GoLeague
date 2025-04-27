package handlers

import (
	"goleague/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Tier list handler.
type TierlistHandler struct {
	tierlistService services.TierlistService
}

// Create a new instance of the tierlist handler.
func NewTierlistHandler(service services.TierlistService) *TierlistHandler {
	return &TierlistHandler{
		tierlistService: service,
	}
}

// Handler for getting the tierlist.
func (h *TierlistHandler) GetTierlist(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "okay"})
}
