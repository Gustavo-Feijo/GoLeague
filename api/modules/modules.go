package modules

import (
	"goleague/api/handlers"
	"goleague/api/repositories"
	"goleague/api/services"
	"log"

	"github.com/gin-gonic/gin"
)

// Module containing the necessary handlers.
type Module struct {
	Router          *gin.Engine
	TierlistHandler *handlers.TierlistHandler
}

// Create a new module with all the necessary handlers initialized.
func NewModule() *Module {
	router := gin.Default()

	// Initialize necessary repositories.
	tierlistRepo, err := repositories.NewTierlistRepository()
	if err != nil {
		log.Fatalf("Couldn't start the tierlist repository: %v", err)
	}

	// Initialize the services.
	tierlistService := services.NewTierlistService(tierlistRepo)

	// Initialize the handlers.
	tierlistHandler := handlers.NewTierlistHandler(tierlistService)

	// Return the module with all handlers.
	return &Module{
		Router:          router,
		TierlistHandler: tierlistHandler,
	}
}
