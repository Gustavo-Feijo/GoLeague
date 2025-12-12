package modules

import (
	"goleague/api/handlers"
	tierlistservice "goleague/api/services/tierlist"
)

func initializeTierlistHandler(deps *ModuleDependencies) *handlers.TierlistHandler {
	// Initialize the tierlist service and handler.
	tierlistDeps := &tierlistservice.TierlistServiceDeps{
		DB:       deps.DB,
		MemCache: deps.TierlistMemCache,
		Redis:    deps.Redis,
	}

	tierlistService := tierlistservice.NewTierlistService(tierlistDeps)

	tierlistHandlerDeps := &handlers.TierlistHandlerDependencies{
		TierlistService: tierlistService,
	}

	return handlers.NewTierlistHandler(tierlistHandlerDeps)
}
