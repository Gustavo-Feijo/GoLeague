package redis

import (
	"context"
	"goleague/pkg/config"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Type for the client.
type RedisClient struct {
	*redis.Client
}

var (
	once     sync.Once
	instance *RedisClient
)

// Return the only existing instance of the client.
func GetClient() *RedisClient {
	once.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:         config.Redis.Host + ":" + config.Redis.Port,
			Password:     config.Redis.Password,
			DB:           0,
			MaxRetries:   3,
			PoolSize:     100,
			MinIdleConns: 10,
			PoolTimeout:  30 * time.Second,
		})

		instance = &RedisClient{
			Client: client,
		}
	})
	return instance
}

// Close the client connection.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Wrapper to return the Result directly.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Wrapper to already return the .Err()
func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}
