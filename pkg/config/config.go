package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ApiKey    string
	Bucket    BucketConfig
	Database  DatabaseConfig
	Grpc      GRPCConfig
	Limits    RiotLimiterConfig
	PrintLogs bool
	Redis     RedisConfig
}

type BucketConfig struct {
	AccessKey    string
	AccessSecret string
	Endpoint     string
	LogBucket    string
	Region       string
}

type DatabaseConfig struct {
	Database       string
	Host           string
	MigrationsPath string
	Password       string
	Port           string
	User           string
	DSN            string
}

type GRPCConfig struct {
	Host string
	Port string
}

type RedisConfig struct {
	Host     string
	Password string
	Port     string
}

type riotLimits struct {
	Count         int
	ResetInterval time.Duration
}

type RiotLimiterConfig struct {
	Lower        riotLimits
	Higher       riotLimits
	SlowInterval time.Duration
}

// Constant valuues based on the personal/development Riot key.
const (
	defaultLowerCount  = 20
	defaultLowerReset  = 1 // Seconds
	defaultHigherCount = 100
	defaultHigherReset = 120 // Seconds
)

func Load() (*Config, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY is required")
	}

	// Load higher limit settings.
	higherCount := getEnvInt("LIMIT_HIGHER_COUNT", defaultHigherCount)
	higherReset := getEnvInt("LIMIT_HIGHER_RESET", defaultHigherReset)

	// Get the API Limits.
	// Load lower limit settings.
	lowerCount := getEnvInt("LIMIT_LOWER_COUNT", defaultLowerCount)
	lowerReset := getEnvInt("LIMIT_LOWER_RESET", defaultLowerReset)

	printLogs, _ := strconv.ParseBool(os.Getenv("ENABLE_CONSOLE_LOG"))

	jobInterval := (float64(higherReset) / float64(higherCount)) * 1000

	dbConfig := DatabaseConfig{
		Database:       os.Getenv("POSTGRES_DATABASE"),
		Host:           os.Getenv("POSTGRES_HOST"),
		Password:       os.Getenv("POSTGRES_PASSWORD"),
		Port:           os.Getenv("POSTGRES_PORT"),
		User:           os.Getenv("POSTGRES_USER"),
		MigrationsPath: os.Getenv("MIGRATIONS_PATH"),
	}

	dbConfig.DSN = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Database, dbConfig.Port,
	)

	// Validate required database fields
	if dbConfig.Host == "" || dbConfig.User == "" || dbConfig.Database == "" {
		return nil, fmt.Errorf("required database configuration missing")
	}

	return &Config{
		ApiKey: apiKey,
		Bucket: BucketConfig{
			AccessKey:    os.Getenv("BUCKET_ACCESS_KEY"),
			AccessSecret: os.Getenv("BUCKET_ACCESS_SECRET"),
			Endpoint:     os.Getenv("BUCKET_ENDPOINT"),
			LogBucket:    os.Getenv("BUCKET_LOGGER_NAME"),
			Region:       os.Getenv("BUCKET_REGION"),
		},
		Database: dbConfig,
		Grpc: GRPCConfig{
			Host: os.Getenv("GRPC_HOST"),
			Port: os.Getenv("GRPC_PORT"),
		},
		Limits: RiotLimiterConfig{
			Lower: riotLimits{
				Count:         lowerCount,
				ResetInterval: time.Duration(lowerReset) * time.Second,
			},
			Higher: riotLimits{
				Count:         higherCount,
				ResetInterval: time.Duration(higherReset) * time.Second,
			},
			SlowInterval: time.Duration(jobInterval) * time.Millisecond,
		},
		PrintLogs: printLogs,
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Password: os.Getenv("REDIS_PASSWORD"),
			Port:     os.Getenv("REDIS_PORT"),
		},
	}, nil
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
