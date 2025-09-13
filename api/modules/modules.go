package modules

import (
	"context"
	"goleague/api/cache"
	grpcclient "goleague/api/grpc"
	"goleague/api/handlers"
	"goleague/api/services"
	"goleague/pkg/redis"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Module containing the necessary handlers.
type Module struct {
	Router          *gin.Engine
	PlayerHandler   *handlers.PlayerHandler
	TierlistHandler *handlers.TierlistHandler
}

// ModuleDependencies holds the necessary dependencies to the module start.
type ModuleDependencies struct {
	DB         *gorm.DB
	GrpcClient *grpc.ClientConn
	MemCache   *cache.MemCache
	Redis      *redis.RedisClient
}

// Create a new module with all the necessary handlers initialized.
func NewModule(deps *ModuleDependencies) (*Module, error) {
	router := gin.Default()

	// Preload the cache.
	championCache := cache.NewChampionCache(deps.DB, deps.Redis, deps.MemCache)
	championCache.Initialize(context.Background())

	matchCache := cache.NewMatchCache(deps.Redis)

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
	tierlistHandler := handlers.NewTierlistHandler(tierlistHandlerDeps)

	grpcClient := grpcclient.NewPlayerGRPCClient(deps.GrpcClient)

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

	playerHandler := handlers.NewPlayerHandler(playerHandlerDeps)

	// Return the module with all handlers.
	return &Module{
		Router:          router,
		PlayerHandler:   playerHandler,
		TierlistHandler: tierlistHandler,
	}, nil
}
