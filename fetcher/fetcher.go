package main

import (
	"context"
	"goleague/fetcher/queue"
	"goleague/fetcher/regions"
	"goleague/pkg/config"
	"goleague/pkg/database"
	"goleague/pkg/database/models"
	pb "goleague/pkg/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	config.LoadEnv()

	_, stop := context.WithCancel(context.Background())

	defer stop()
	log.Println("Starting grpcServer...")

	// Create the manager that will be used to handle all the regions fetching.
	// Will be passed by reference to any place that make requests.
	manager := regions.GetRegionManager()

	// Migrate all necessary models.
	db, err := database.GetConnection()
	if err != nil {
		log.Fatal(err)
	}

	// Check and create ENUM types if they do not exist
	err = db.Exec(`
		DO $$ 
		BEGIN
		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'queue_type') THEN
		        CREATE TYPE queue_type AS ENUM ('RANKED_SOLO_5x5', 'RANKED_FLEX_SR');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_type') THEN
		        CREATE TYPE tier_type AS ENUM ('IRON', 'BRONZE', 'SILVER', 'GOLD', 'PLATINUM', 'EMERALD', 'DIAMOND', 'MASTER', 'GRANDMASTER', 'CHALLENGER');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'rank_type') THEN
		        CREATE TYPE rank_type AS ENUM ('IV', 'III', 'II', 'I');
		    END IF;
		END $$;
	`).Error

	if err != nil {
		log.Fatal(err)
	}

	// Automigrate the models.
	err = db.AutoMigrate(
		&models.MatchInfo{},
		&models.MatchBans{},
		&models.PlayerInfo{},
		&models.RatingEntry{},
		&models.MatchStats{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start the queue.
	go queue.StartQueue(manager)

	// Start the gRPC server.
	grpcServer, healthServer := startGRPCServer(manager)

	// Shutdown everything.
	handleShutdown(grpcServer, healthServer, stop)
}

// Start the grpc server for handling cache on demand.
func startGRPCServer(regionManager *regions.RegionManager) (*grpc.Server, *health.Server) {
	// Start a TPC listener.
	list, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Couldn't start the tcp server: %v", err)
	}

	// Create the server, register it and serve.
	grpcServer := grpc.NewServer()
	srv := &server{regionManager: regionManager}
	pb.RegisterAssetsServiceServer(grpcServer, srv)

	// Register the health check.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the serving status as serving.
	healthServer.SetServingStatus("goleague.AssetService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Run a go routine for the grpc server.
	go func() {
		log.Println("Running gRPC server.")
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("Failed to server grpc: %v", err)
		}
	}()

	//  Return the grpc and health server.
	return grpcServer, healthServer
}

// Handle the shutdown of the whole server.
func handleShutdown(grpcServer *grpc.Server, healthServer *health.Server, cancel context.CancelFunc) {
	// Create the signal channel.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	<-signalChannel

	// Set it to not serving.
	healthServer.SetServingStatus("goleague.AssetService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		grpcServer.GracefulStop()
	}()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	cancel()
}
