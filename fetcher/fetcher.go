package main

import (
	"context"
	"goleague/fetcher/queue"
	regionmanager "goleague/fetcher/regionmanager"
	"goleague/pkg/config"
	"goleague/pkg/database"
	pb "goleague/pkg/grpc"
	"goleague/pkg/logger"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "docker" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Couldn't initialize the configuration: %v", err)
	}

	_, stop := context.WithCancel(context.Background())

	defer stop()

	log.Println("Running migration and creating triggers/enums...")

	// Creates the database connection.
	db, err := database.NewConnection(cfg.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// Runs the migrations.
	rawDb, err := db.DB()
	if err != nil {
		log.Fatalf("Couldn't get raw db connection: %v", err)
	}

	if err := database.RunMigrations(cfg.Database, rawDb); err != nil {
		log.Fatal(err)
	}

	log.Println("Instanciating Region Managers...")

	// Pass down the necessary dependencies.
	deps := regionmanager.RegionManagerDependencies{
		DB: db,
	}

	// Create the manager that will be used to handle all the regions fetching.
	manager, err := regionmanager.NewRegionManager(cfg, deps)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Region Managers created...")

	log.Println("Starting the queues...")
	// Start the queue.
	go queue.StartQueue(manager)

	// Start the gRPC server.
	grpcServer, healthServer := startGRPCServer(cfg, manager)

	// Shutdown everything.
	handleShutdown(grpcServer, healthServer, stop)
}

// Start the grpc server for handling cache on demand.
func startGRPCServer(config *config.Config, regionManager *regionmanager.RegionManager) (*grpc.Server, *health.Server) {
	// Start a TPC listener.
	list, err := net.Listen("tcp", ":"+config.Grpc.Port)
	if err != nil {
		log.Fatalf("Couldn't start the tcp server: %v", err)
	}

	// Create the server, register it and serve.
	grpcServer := grpc.NewServer()

	// Create a logger for thge gRPC server requests.
	logger, err := logger.CreateLogger(config)
	if err != nil {
		log.Fatalf("Couldn't start the gRPC server logger: %v", err)
	}

	srv := &server{
		logger:        logger,
		regionManager: regionManager,
	}

	pb.RegisterServiceServer(grpcServer, srv)

	// Register the health check.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the serving status as serving.
	healthServer.SetServingStatus("goleague.AssetService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Run a go routine for the grpc server.
	go func() {
		log.Println("Starting gRPC server...")
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
