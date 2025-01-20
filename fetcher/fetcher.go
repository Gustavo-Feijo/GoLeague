package main

import (
	"context"
	"goleague/pkg/config"
	"goleague/pkg/redis"
	"time"
)

func main() {
	config.LoadEnv()
	client := redis.GetClient()
	client.Set(context.Background(), "teste", "batata", time.Hour)
}
