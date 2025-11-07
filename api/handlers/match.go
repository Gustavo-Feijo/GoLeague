package handlers

import (
	"goleague/api/filters"
	"goleague/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlayerHandler is the handler for the player endpoints.
type MatchHandler struct {
	MatchService *services.MatchService
}

type MatchHandlerDependencies struct {
	MatchService *services.MatchService
}

// NewMatchHandler  creates a new instance of the match handler.
func NewMatchHandler(deps *MatchHandlerDependencies) *MatchHandler {
	return &MatchHandler{
		MatchService: deps.MatchService,
	}
}

// Helper to bind the default URI params for matches.
func (h *MatchHandler) bindURIParams(c *gin.Context) (*filters.MatchURIParams, error) {
	var mp filters.MatchURIParams
	if err := c.ShouldBindUri(&mp); err != nil {
		return nil, err
	}
	return &mp, nil
}

// GetFullMatchData is the handler to return all available data for a given match.
func (h *MatchHandler) GetFullMatchData(c *gin.Context) {

	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewGetFullMatchDataFilter(pp)

	matchData, err := h.MatchService.GetFullMatchData(filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": matchData})
}
