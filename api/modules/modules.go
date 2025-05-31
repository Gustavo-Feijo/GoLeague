package modules

import (
	"context"
	"fmt"
	"goleague/api/cache"
	"goleague/api/handlers"
	"goleague/api/services"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// Module containing the necessary handlers.
type Module struct {
	Router          *gin.Engine
	PlayerHandler   *handlers.PlayerHandler
	TierlistHandler *handlers.TierlistHandler
}

// Create a new module with all the necessary handlers initialized.
func NewModule(grpcClient *grpc.ClientConn) (*Module, error) {
	router := gin.Default()

	// Preload the cache.
	cache.GetChampionCache().Initialize(context.Background())

	// Initialize the services.
	tierlistService, err := services.NewTierlistService(grpcClient)
	if err != nil {
		return nil, fmt.Errorf("couldn't start the tierlist service: %v", err)
	}
	// Initialize the handlers.
	tierlistHandler := handlers.NewTierlistHandler(tierlistService)

	// Initialize the player service.
	playerService, err := services.NewPlayerService(grpcClient)
	if err != nil {
		return nil, fmt.Errorf("couldn't start the player service: %v", err)
	}
	// Initialize the handlers.
	playerHandler := handlers.NewPlayerHandler(playerService)

	// Return the module with all handlers.
	return &Module{
		Router:          router,
		PlayerHandler:   playerHandler,
		TierlistHandler: tierlistHandler,
	}, nil
}
