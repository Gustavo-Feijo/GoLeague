package routes

import (
	"goleague/api/handlers"

	"github.com/gin-gonic/gin"
)

// Router wrapper.
type Router struct {
	Engine *gin.Engine
	api    *gin.RouterGroup
}

// NewRouter creates a router engine with a API group.
func NewRouter(engine *gin.Engine) *Router {
	return &Router{
		api:    engine.Group("/api/v1"),
		Engine: engine,
	}
}

// SetupRoutes go through each passed handler and register it.
func (r *Router) SetupRoutes(handlerList ...any) {
	for _, h := range handlerList {
		switch handler := h.(type) {
		case *handlers.TierlistHandler:
			r.registerTierlistHandler(handler)
		case *handlers.PlayerHandler:
			r.registerPlayerHandler(handler)
		}
	}
}

// registerTierlistHandler implements the tierlist routes.
func (r *Router) registerTierlistHandler(handler *handlers.TierlistHandler) {
	tierlist := r.api.Group("/tierlist")
	{
		tierlist.GET("", handler.GetTierlist)
	}
}

// registerPlayerHandler implements the player routes.
func (r *Router) registerPlayerHandler(handler *handlers.PlayerHandler) {
	player := r.api.Group("/player")
	{
		player.GET("search", handler.GetPlayerSearch)
	}
}

// Start the router.
func (r *Router) Run(addr string) error {
	return r.Engine.Run(addr)
}
