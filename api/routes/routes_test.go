package routes

import (
	"testing"

	"goleague/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *Router {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	return NewRouter(engine)
}

func TestNewRouter(t *testing.T) {
	router := setupTestRouter()

	assert.NotNil(t, router)
	assert.NotNil(t, router.Engine)
	assert.NotNil(t, router.api)
}

func TestSetupRoutes(t *testing.T) {
	router := setupTestRouter()

	// Create mock handlers (you'll need to implement these)
	tierlistHandler := &handlers.TierlistHandler{}
	playerHandler := &handlers.PlayerHandler{}
	matchHandler := &handlers.MatchHandler{}

	router.SetupRoutes(tierlistHandler, playerHandler, matchHandler)

	routes := router.Engine.Routes()
	assert.Greater(t, len(routes), 0)
}
