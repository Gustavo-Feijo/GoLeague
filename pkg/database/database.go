package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetConnections is a singleton implementaudo ssion of the database.
// Return the connection pool.
func NewConnection(dsn string) (*gorm.DB, error) {

	// Create the database instance.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get the SQL database itself.
	sqlDb, sqlErr := db.DB()

	// Verify if could get the connection.
	if sqlErr != nil {
		return nil, fmt.Errorf("failed to get the sql connection: %v", err)
	}

	// Set the pool values.
	sqlDb.SetMaxOpenConns(400)
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetConnMaxLifetime(time.Hour)
	sqlDb.SetConnMaxIdleTime(time.Hour)

	// Test the connection
	if err := sqlDb.Ping(); err != nil {
		sqlDb.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, err
}
