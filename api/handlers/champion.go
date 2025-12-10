package handlers

import (
	"goleague/api/filters"
	championservice "goleague/api/services/champion"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlayerHandler is the handler for the player endpoints.
type ChampionHandler struct {
	ChampionService *championservice.ChampionService
}

type ChampionHandlerDependencies struct {
	ChampionService *championservice.ChampionService
}

// NewChampionHandler  creates a new instance of the champion handler.
func NewChampionHandler(deps *ChampionHandlerDependencies) *ChampionHandler {
	return &ChampionHandler{
		ChampionService: deps.ChampionService,
	}
}

// Helper to bind the default URI params for championes.
func (h *ChampionHandler) bindURIParams(c *gin.Context) (*filters.ChampionURIParams, error) {
	var mp filters.ChampionURIParams
	if err := c.ShouldBindUri(&mp); err != nil {
		return nil, err
	}
	return &mp, nil
}

// GetChampionData is the handler to return all available data for a given champion.
func (h *ChampionHandler) GetChampionData(c *gin.Context) {

	pp, err := h.bindURIParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := filters.NewGetChampionDataFilter(pp)

	championData, err := h.ChampionService.GetChampionData(c, filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": championData})
}

// GetAllChampions is the handler to return all available data for all champions.
func (h *ChampionHandler) GetAllChampions(c *gin.Context) {
	championData, err := h.ChampionService.GetAllChampions(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": championData})
}
