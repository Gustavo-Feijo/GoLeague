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

// Helper to bind the default URI params for players.
func (h *PlayerHandler) bindURIParams(c *gin.Context) (*filters.PlayerURIParams, error) {
	var pp filters.PlayerURIParams
	if err := c.ShouldBindUri(&pp); err != nil {
		return nil, err
	}
	return &pp, nil
}

// ForceFetchPlayer calls the Fetcher service via gRPC to save a given player in the database.
func (h *PlayerHandler) ForceFetchPlayer(c *gin.Context) {
	// Path params.
	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewForceFetchPlayerFilter(pp)

	summoner, err := h.playerService.ForceFetchPlayer(filters)
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
	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewForceFetchMatchHistoryFilter(pp)

	confirm, err := h.playerService.ForceFetchPlayerMatchHistory(filters)
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

	filters := filters.NewPlayerSearchFilter(qp)

	result, err := h.playerService.GetPlayerSearch(filters)
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

	// Path params.
	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewPlayerMatchHistoryFilter(qp, pp)

	matchList, err := h.playerService.GetPlayerMatchHistory(filters)
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

	// Path params.
	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewPlayerStatsFilter(qp, pp)

	playerStats, err := h.playerService.GetPlayerStats(filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": playerStats})
}

// GetPlayerInfo handles getting all the player related data.
func (h *PlayerHandler) GetPlayerInfo(c *gin.Context) {
	// Path params.
	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewPlayerInfoFilter(pp)

	playerInfo, err := h.playerService.GetPlayerInfo(filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": playerInfo})
}
