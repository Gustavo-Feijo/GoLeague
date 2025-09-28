package modules

import (
	"goleague/api/cache"
	"goleague/api/handlers"
	"goleague/api/services"
)

func initializeMatchHandler(deps *ModuleDependencies) *handlers.MatchHandler {

	// Preload the cache.
	championCache := cache.NewChampionCache(deps.DB, deps.Redis, deps.MemCache)

	matchDeps := &services.MatchServiceDeps{
		DB:            deps.DB,
		ChampionCache: championCache,
		MemCache:      deps.MemCache,
		Redis:         deps.Redis,
	}

	matchService := services.NewMatchService(matchDeps)

	matchHandlerDeps := &handlers.MatchHandlerDependencies{
		MatchService: matchService,
	}

	return handlers.NewMatchHandler(matchHandlerDeps)
}
