package config

import (
	"fmt"
	"os"
)

// Redis configuration struct.
type RedisConfiguration struct {
	Host     string
	Password string
	Port     string
}

type DatabaseConfiguration struct {
	Database string
	Host     string
	Password string
	Port     string
	User     string
	URL      string
}

var (
	Redis    RedisConfiguration
	Database DatabaseConfiguration
)

// Load the variables.
func LoadEnv() {
	// Load the Redis configuration.
	Redis.Host = os.Getenv("REDIS_HOST")
	Redis.Password = os.Getenv("REDIS_PASSWORD")
	Redis.Port = os.Getenv("REDIS_PORT")

	// Load the database configuration.
	Database.Database = os.Getenv("POSTGRES_DATABASE")
	Database.Host = os.Getenv("POSTGRES_HOST")
	Database.Password = os.Getenv("POSTGRES_PASSWORD")
	Database.Port = os.Getenv("POSTGRES_PORT")
	Database.User = os.Getenv("POSTGRES_USER")
	Database.URL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", Database.Host, Database.User, Database.Password, Database.Database, Database.Port)
}
