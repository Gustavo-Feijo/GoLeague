package modules

import (
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/handlers"
	"goleague/pkg/models/champion"
	"goleague/pkg/redis"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Module containing the necessary handlers.
type Module struct {
	Router          *gin.Engine
	ChampionHandler *handlers.ChampionHandler
	MatchHandler    *handlers.MatchHandler
	PlayerHandler   *handlers.PlayerHandler
	TierlistHandler *handlers.TierlistHandler
}

// ModuleDependencies holds the necessary dependencies to the module start.
type ModuleDependencies struct {
	DB               *gorm.DB
	ChampionCache    cache.ChampionCache
	GrpcClient       *grpc.ClientConn
	ChampionMemCache cache.MemCache[*champion.Champion]
	TierlistMemCache cache.MemCache[[]*dto.TierlistResult]
	Redis            *redis.RedisClient
}

// Create a new module with all the necessary handlers initialized.
func NewModule(deps *ModuleDependencies) (*Module, error) {
	router := gin.Default()

	// Return the module with all handlers.
	return &Module{
		Router:          router,
		ChampionHandler: initializeChampionHandler(deps),
		MatchHandler:    initializeMatchHandler(deps),
		PlayerHandler:   initializePlayerHandler(deps),
		TierlistHandler: initializeTierlistHandler(deps),
	}, nil
}
