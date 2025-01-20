package main

import (
	"context"
	"fmt"
	"goleague/pkg/config"
	"goleague/pkg/redis"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()
	client := redis.GetClient()
	client.Set(context.Background(), "teste", "batata", time.Hour)
	fmt.Println(config.Database.Port)
	_, err := gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	}
	fmt.Println("Connected")
}
