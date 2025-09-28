package modules

import (
	"goleague/api/cache"
	grpcclient "goleague/api/grpc"
	"goleague/api/handlers"
	"goleague/api/services"
)

func initializePlayerHandler(deps *ModuleDependencies) *handlers.PlayerHandler {
	grpcClient := grpcclient.NewPlayerGRPCClient(deps.GrpcClient)
	matchCache := cache.NewMatchCache(deps.Redis)

	// Initialize the player service and handler.
	playerDeps := &services.PlayerServiceDeps{
		DB:         deps.DB,
		GrpcClient: grpcClient,
		MatchCache: matchCache,
		Redis:      deps.Redis,
	}

	playerService := services.NewPlayerService(playerDeps)

	playerHandlerDeps := &handlers.PlayerHandlerDependencies{
		PlayerService: playerService,
	}

	return handlers.NewPlayerHandler(playerHandlerDeps)
}
