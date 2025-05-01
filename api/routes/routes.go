package routes

import (
	"goleague/api/handlers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
	api    *gin.RouterGroup
}

func NewRouter(engine *gin.Engine) *Router {
	return &Router{
		api:    engine.Group("/api/v1"),
		engine: engine,
	}
}

func (r *Router) SetupRoutes(handlerList ...any) {
	for _, h := range handlerList {
		switch handler := h.(type) {
		case *handlers.TierlistHandler:
			r.registerTierlistHandler(handler)
		}
	}
}

// Register the tierlist handler.
func (r *Router) registerTierlistHandler(handler *handlers.TierlistHandler) {
	tierlist := r.api.Group("/tierlist")
	{
		tierlist.GET("", handler.GetTierlist)
	}
}

// Start the router.
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
