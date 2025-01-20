package main

import (
	"context"
	"fmt"
	"goleague/pkg/config"
	"goleague/pkg/redis"
	"time"
)

// Simple testing for the Redis.
func main() {
	config.LoadEnv()
	client := redis.GetClient()
	time.Sleep(time.Second * 5)
	data, err := client.Get(context.Background(), "teste")
	if err == nil {
		fmt.Println(data)
	}
}
