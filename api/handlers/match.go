package handlers

import (
	"goleague/api/filters"
	"goleague/api/services"

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
