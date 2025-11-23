package modules

import (
	"goleague/api/cache"
	grpcclient "goleague/api/grpc"
	"goleague/api/handlers"
	playerservice "goleague/api/services/player"
)

func initializePlayerHandler(deps *ModuleDependencies) *handlers.PlayerHandler {
	grpcClient := grpcclient.NewPlayerGRPCClient(deps.GrpcClient)
	matchCache := cache.NewMatchCache(deps.Redis)

	// Initialize the player service and handler.
	playerDeps := &playerservice.PlayerServiceDeps{
		DB:         deps.DB,
		GrpcClient: grpcClient,
		MatchCache: matchCache,
		Redis:      deps.Redis,
	}

	playerService := playerservice.NewPlayerService(playerDeps)

	playerHandlerDeps := &handlers.PlayerHandlerDependencies{
		PlayerService: playerService,
	}

	return handlers.NewPlayerHandler(playerHandlerDeps)
}
