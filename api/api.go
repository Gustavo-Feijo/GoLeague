package main

import (
	"goleague/api/modules"
	"goleague/api/routes"
	"goleague/pkg/config"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Simple testing for the Redis.
func main() {
	// Load the environment variables if not running on Docker.
	if os.Getenv("ENVIRONMENT") != "docker" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	config.LoadEnv()

	// Connect to the fetcher grpc.
	grpcClient, err := grpc.NewClient("fetcher:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error to connect to the gRPC server: %v", err)
	}

	defer grpcClient.Close()

	// Create a module with all necessary handlers.
	module := modules.NewModule()

	// Create a new router with the routes setup.
	router := routes.NewRouter(module.Router)
	router.SetupRoutes(
		module.TierlistHandler,
	)

	// Start the server.
	router.Run(":8080")
}
