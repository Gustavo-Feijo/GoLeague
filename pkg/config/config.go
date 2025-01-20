package config

import (
	"os"
)

// Redis configuration struct.
type RedisConfiguration struct {
	Host     string
	Port     string
	Password string
}

type DatabaseConfiguration struct{}

var (
	Redis    RedisConfiguration
	Database DatabaseConfiguration
)

// Load the variables.
func LoadEnv() {
	// Load the Redis configuration.
	Redis.Host = os.Getenv("REDIS_HOST")
	Redis.Port = os.Getenv("REDIS_PORT")
	Redis.Password = os.Getenv("REDIS_PASSWORD")
}
