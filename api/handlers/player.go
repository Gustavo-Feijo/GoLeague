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

// GetPlayerSearch handles requests for player searching.
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

// GetPlayerMatchHistory handles requests for retrieving a player match history.
func (h *PlayerHandler) GetPlayerMatchHistory(c *gin.Context) {
	var qp filters.PlayerMatchHistoryParams

	if err := c.ShouldBindQuery(&qp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filtersMap := qp.AsMap()
	filtersMap["gameName"] = c.Params.ByName("gameName")
	filtersMap["gameTag"] = c.Params.ByName("gameTag")
	filtersMap["region"] = c.Params.ByName("region")

	h.playerService.GetPlayerMatchHistory(filtersMap)
	c.JSON(http.StatusOK, gin.H{"teste": c.Params.ByName("gameName")})
}

// GetPlayerStats handles requests for retrieving a player average status.
func (h *PlayerHandler) GetPlayerStats(c *gin.Context) {
}

// GetPlayerElo handles rating related data from a player.
func (h *PlayerHandler) GetPlayerElo(c *gin.Context) {
}
