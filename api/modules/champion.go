package modules

import (
	"goleague/api/handlers"
	championservice "goleague/api/services/champion"
)

func initializeChampionHandler(deps *ModuleDependencies) *handlers.ChampionHandler {
	championDeps := &championservice.ChampionServiceDeps{
		DB:            deps.DB,
		ChampionCache: deps.ChampionCache,
		MemCache:      deps.ChampionMemCache,
	}

	championService := championservice.NewChampionService(championDeps)

	championHandlerDeps := &handlers.ChampionHandlerDependencies{
		ChampionService: championService,
	}

	return handlers.NewChampionHandler(championHandlerDeps)
}
