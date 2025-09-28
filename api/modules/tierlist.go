package modules

import (
	"goleague/api/cache"
	"goleague/api/handlers"
	"goleague/api/services"
)

func initializeTierlistHandler(deps *ModuleDependencies) *handlers.TierlistHandler {

	// Preload the cache.
	championCache := cache.NewChampionCache(deps.DB, deps.Redis, deps.MemCache)

	// Initialize the tierlist service and handler.
	tierlistDeps := &services.TierlistServiceDeps{
		DB:            deps.DB,
		ChampionCache: championCache,
		MemCache:      deps.MemCache,
		Redis:         deps.Redis,
	}

	tierlistService := services.NewTierlistService(tierlistDeps)

	tierlistHandlerDeps := &handlers.TierlistHandlerDependencies{
		MemCache:        deps.MemCache,
		Redis:           deps.Redis,
		TierlistService: tierlistService,
	}

	return handlers.NewTierlistHandler(tierlistHandlerDeps)
}
