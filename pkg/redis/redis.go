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
	client *redis.Client
}

var (
	once     sync.Once
	instance *RedisClient
)

// Return the only existing instance of the client.
func GetClient() *RedisClient {
	once.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:     config.Redis.Host + ":" + config.Redis.Port,
			Password: config.Redis.Password,
			DB:       0,
		})

		instance = &RedisClient{
			client: client,
		}
	})
	return instance
}

// Close the client connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Wrapper to return the Result directly.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Wrapper to already return the .Err()
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}
