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

// GetClient is a singleton that returns a redis connection pool.
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

// Close clopses the client connection.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Get is a simple wrapper to the Get, returning the result directly.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Set is a wrapper to already return the .Err() on the Set.
func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// GetKeysByPrefix return all keys that match a given prefix.
func (r *RedisClient) GetKeysByPrefix(ctx context.Context, prefix string) ([]string, error) {
	var cursor uint64
	var keys []string

	for {
		var result []string
		var err error

		result, cursor, err = r.Client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, result...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
