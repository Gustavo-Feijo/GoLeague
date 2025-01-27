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
	time.Sleep(time.Second * 10)
	data, err := client.Get(context.Background(), "ddragon:champion:62")
	if err == nil {
		fmt.Println(data)
	} else {
		fmt.Println(err)
	}
}
