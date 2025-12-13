package testutil

import (
	"context"
	"fmt"
	"goleague/pkg/config"
	"goleague/pkg/database"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetConnections is a singleton implementaudo ssion of the database.
// Return the connection pool.
func NewTestConnection(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "postgres:16",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)

	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=testdb sslmode=disable TimeZone=UTC",
		host, port.Port(),
	)

	fmt.Println(dsn)

	// Create the database instance.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}

	// Get the SQL database itself.
	sqlDB, sqlErr := db.DB()

	// Verify if could get the connection.
	if sqlErr != nil {
		t.Fatalf("Failed to get SQL DB: %v", sqlErr)
	}

	// Set the pool values.
	sqlDB.SetMaxOpenConns(400)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		t.Fatalf("failed ping: %v", err)
	}

	// Run the migrations to replicate the full schema.
	database.RunMigrations(cfg, sqlDB)

	cleanup := func() {
		sqlDB.Close()
		testcontainers.CleanupContainer(t, container)
	}

	return db, cleanup
}
