package modules

import (
	"goleague/api/handlers"
	matchservice "goleague/api/services/match"
)

func initializeMatchHandler(deps *ModuleDependencies) *handlers.MatchHandler {
	matchDeps := &matchservice.MatchServiceDeps{
		DB:       deps.DB,
		MemCache: deps.MemCache,
	}

	matchService := matchservice.NewMatchService(matchDeps)

	matchHandlerDeps := &handlers.MatchHandlerDependencies{
		MatchService: matchService,
	}

	return handlers.NewMatchHandler(matchHandlerDeps)
}
