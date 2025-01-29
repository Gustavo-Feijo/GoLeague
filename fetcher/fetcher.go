package main

import (
	"fmt"
	"goleague/pkg/config"
	pb "goleague/pkg/grpc"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	config.LoadEnv()

	fmt.Println("Starting grpcServer...")
	// Start the gRPC server for
	startgrpcServer()
}

// Start the grpc server for handling cache on demand.
func startgrpcServer() {
	// Start a TPC listener.
	list, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Couldn't start the tcp server: %v", err)
	}

	// Create the server, register it and serve.
	grpcServer := grpc.NewServer()
	pb.RegisterAssetsServiceServer(grpcServer, &server{})

	// Register the health check.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the serving status as serving.
	healthServer.SetServingStatus("goleague.AssetService", grpc_health_v1.HealthCheckResponse_SERVING)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("Failed to server grpc: %v", err)
	}
	log.Println("Succsfully started the grpc server.")
}
