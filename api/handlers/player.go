package handlers

import (
	"goleague/api/filters"
	"goleague/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlayerHandler is the handler for the player endpoints.
type PlayerHandler struct {
	playerService *services.PlayerService
}

type PlayerHandlerDependencies struct {
	PlayerService *services.PlayerService
}

// NewPlayerHandler creates a new instance of the player handler.
func NewPlayerHandler(deps *PlayerHandlerDependencies) *PlayerHandler {
	return &PlayerHandler{
		playerService: deps.PlayerService,
	}
}

// ForceFetchPlayer calls the Fetcher service via gRPC to save a given player in the database.
func (h *PlayerHandler) ForceFetchPlayer(c *gin.Context) {
	// Path params.
	var pp filters.PlayerForceFetchParams
	if err := c.ShouldBindUri(&pp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	summoner, err := h.playerService.ForceFetchPlayer(pp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": summoner})
}

// ForceFetchPlayerMatchHistory forcefully fetches a given player match history.
// Don't return the match history, only a confirmation, since the fetching can take some time.
func (h *PlayerHandler) ForceFetchPlayerMatchHistory(c *gin.Context) {
	// Path params.
	var pp filters.PlayerForceFetchMatchHistoryParams
	if err := c.ShouldBindUri(&pp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	confirm, err := h.playerService.ForceFetchPlayerMatchHistory(pp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": confirm})
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

	matchList, err := h.playerService.GetPlayerMatchHistory(filtersMap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": matchList})
}

// GetPlayerStats handles requests for retrieving a player average status.
func (h *PlayerHandler) GetPlayerStats(c *gin.Context) {
	var qp filters.PlayerStatsParams

	if err := c.ShouldBindQuery(&qp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filtersMap := qp.AsMap()
	filtersMap["gameName"] = c.Params.ByName("gameName")
	filtersMap["gameTag"] = c.Params.ByName("gameTag")
	filtersMap["region"] = c.Params.ByName("region")

	playerStats, err := h.playerService.GetPlayerStats(filtersMap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": playerStats})
}

// GetPlayerElo handles rating related data from a player.
func (h *PlayerHandler) GetPlayerElo(c *gin.Context) {
}
