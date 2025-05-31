package main

import (
	"goleague/api/modules"
	"goleague/api/routes"
	"goleague/pkg/config"
	"log"
	"net/http"
	"os"
	"time"

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
	module, err := modules.NewModule(grpcClient)
	if err != nil {
		log.Fatalf("Error to start the modules: %v", err)
	}

	// Create a new router with the routes setup.
	router := routes.NewRouter(module.Router)
	router.SetupRoutes(
		module.TierlistHandler,
		module.PlayerHandler,
	)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router.Engine, // router is *gin.Engine which implements http.Handler
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
	}

}
