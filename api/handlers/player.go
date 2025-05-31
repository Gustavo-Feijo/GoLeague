package handlers

import (
	"goleague/api/filters"
	"goleague/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlayerHandler is the handler for the player endpoints.
type PlayerHandler struct {
	playerService services.PlayerService
}

// NewPlayerHandler creates a new instance of the player handler.
func NewPlayerHandler(service *services.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService: *service,
	}
}

// GetPlayerSearch handle requests for player searching.
func (h *PlayerHandler) GetPlayerSearch(c *gin.Context) {
	var qp filters.PlayerSearchParams

	if err := c.ShouldBindQuery(&qp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filtersMap := qp.AsMap()

	result, err := h.playerService.GetPlayerSearch(filtersMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}
