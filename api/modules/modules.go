package modules

import (
	"context"
	"goleague/api/cache"
	"goleague/api/handlers"
	"goleague/pkg/redis"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Module containing the necessary handlers.
type Module struct {
	Router          *gin.Engine
	MatchHandler    *handlers.MatchHandler
	PlayerHandler   *handlers.PlayerHandler
	TierlistHandler *handlers.TierlistHandler
}

// ModuleDependencies holds the necessary dependencies to the module start.
type ModuleDependencies struct {
	DB         *gorm.DB
	GrpcClient *grpc.ClientConn
	MemCache   cache.MemCache
	Redis      *redis.RedisClient
}

// Create a new module with all the necessary handlers initialized.
func NewModule(deps *ModuleDependencies) (*Module, error) {
	router := gin.Default()

	// Create a instance  just for initializing.
	championCache := cache.NewChampionCache(deps.DB, deps.Redis, deps.MemCache)
	championCache.Initialize(context.Background())

	// Return the module with all handlers.
	return &Module{
		Router:          router,
		MatchHandler:    initializeMatchHandler(deps),
		PlayerHandler:   initializePlayerHandler(deps),
		TierlistHandler: initializeTierlistHandler(deps),
	}, nil
}
