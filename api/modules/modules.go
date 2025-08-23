package modules

import (
	"context"
	"fmt"
	"goleague/api/cache"
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
		GrpcClient:    deps.GrpcClient,
		ChampionCache: championCache,
		MemCache:      deps.MemCache,
		Redis:         deps.Redis,
	}

	tierlistService, err := services.NewTierlistService(tierlistDeps)
	if err != nil {
		return nil, fmt.Errorf("couldn't start the tierlist service: %v", err)
	}

	tierlistHandlerDeps := &handlers.TierlistHandlerDependencies{
		MemCache:        deps.MemCache,
		Redis:           deps.Redis,
		TierlistService: tierlistService,
	}
	tierlistHandler := handlers.NewTierlistHandler(tierlistHandlerDeps)

	// Initialize the player service and handler.
	playerDeps := &services.PlayerServiceDeps{
		DB:         deps.DB,
		GrpcClient: deps.GrpcClient,
		MatchCache: matchCache,
		Redis:      deps.Redis,
	}

	playerService, err := services.NewPlayerService(playerDeps)
	if err != nil {
		return nil, fmt.Errorf("couldn't start the player service: %v", err)
	}

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
