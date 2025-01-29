package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/pkg/config"
	pb "goleague/pkg/grpc"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Simple testing for the Redis.
func main() {
	config.LoadEnv()

	// Connect to the fetcher grpc.
	conn, err := grpc.NewClient("fetcher:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error to connect to the gRPC server: %v", err)
	}

	defer conn.Close()

	// Create the context.
	c := pb.NewAssetsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()

	// Get a given champion for testing.
	champion, err := c.RevalidateChampionCache(ctx, &pb.ChampionId{Id: "62"})
	if err != nil {
		log.Fatalf("Error at revalidation: %v", err)
	}

	jsonChamp, err := json.Marshal(champion)
	if err != nil {
		log.Fatalf("Error formating the json: %v", err)
	}
	fmt.Println(string(jsonChamp))

	// Get a given champion for testing.
	item, err := c.RevalidateItemCache(ctx, &pb.ItemId{Id: "3158"})
	if err != nil {
		log.Fatalf("Error at revalidation: %v", err)
	}

	jsonItem, err := json.Marshal(item)
	if err != nil {
		log.Fatalf("Error formating the json: %v", err)
	}
	fmt.Println(string(jsonItem))
}
