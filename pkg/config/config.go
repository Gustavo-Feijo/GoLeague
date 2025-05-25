package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Bucket configuration struct.
type BucketConfig struct {
	AccessKey    string
	AccessSecret string
	Endpoint     string
	LogBucket    string
	Region       string
}

// Database configuration struct.
type DatabaseConfiguration struct {
	Database string
	Host     string
	Password string
	Port     string
	User     string
	URL      string
}

type riotLimits struct {
	Count         int
	ResetInterval time.Duration
}

type RiotLimiterConfiguration struct {
	Lower        riotLimits
	Higher       riotLimits
	SlowInterval time.Duration
}

// Redis configuration struct.
type RedisConfiguration struct {
	Host     string
	Password string
	Port     string
}

// Constant valuues based on the personal/development Riot key.
const (
	defaultLowerCount  = 20
	defaultLowerReset  = 1 // Seconds
	defaultHigherCount = 100
	defaultHigherReset = 120 // Seconds
)

var (
	ApiKey   string
	Bucket   BucketConfig
	Database DatabaseConfiguration
	Redis    RedisConfiguration
	Limits   RiotLimiterConfiguration
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

	// Get the Riot API Key.
	ApiKey = os.Getenv("API_KEY")

	// Configure the bucket.
	Bucket.AccessKey = os.Getenv("BUCKET_ACCESS_KEY")
	Bucket.AccessSecret = os.Getenv("BUCKET_ACCESS_SECRET")
	Bucket.Endpoint = os.Getenv("BUCKET_ENDPOINT")
	Bucket.LogBucket = os.Getenv("BUCKET_LOGGER_NAME")
	Bucket.Region = os.Getenv("BUCKET_REGION")

	// Get the API Limits.
	// Load lower limit settings.
	lowerCount := getEnvInt("LIMIT_LOWER_COUNT", defaultLowerCount)
	lowerReset := getEnvInt("LIMIT_LOWER_RESET", defaultLowerReset)

	// Load higher limit settings.
	higherCount := getEnvInt("LIMIT_HIGHER_COUNT", defaultHigherCount)
	higherReset := getEnvInt("LIMIT_HIGHER_RESET", defaultHigherReset)

	// The job interval is how much queries you can run during the higher reset at a consistent rate.
	// Multiply by 1000 to get in milliseconds.
	jobInterval := (float64(higherReset) / float64(higherCount)) * 1000

	Limits = RiotLimiterConfiguration{
		Lower: riotLimits{
			Count:         lowerCount,
			ResetInterval: time.Duration(lowerReset) * time.Second,
		},
		Higher: riotLimits{
			Count:         higherCount,
			ResetInterval: time.Duration(higherReset) * time.Second,
		},
		SlowInterval: time.Duration(jobInterval) * time.Millisecond,
	}
}

// Convert a env key to int or return the default value.
func getEnvInt(key string, defaultVal int) int {
	// Find the env key.
	val := os.Getenv(key)
	if val == "" {
		// Return default if empty.
		return defaultVal
	}

	// Handle the conversion to int.
	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Warning: Couldn't parse the value %s=%s as int. Using default %d: %v", key, val, defaultVal, err)
		return defaultVal
	}

	return intVal
}
