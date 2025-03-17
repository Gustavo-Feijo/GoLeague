package database

import (
	"goleague/pkg/config"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	err  error
	once sync.Once
	db   *gorm.DB
)

// Get or create a database connection as a singleton and return the connection.
func GetConnection() (*gorm.DB, error) {
	// Create the database if doesn't exist, else just return it.
	once.Do(
		func() {
			// Create the database instance.
			db, err = gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
			if err != nil {
				// Can't run without connection to the database.
				log.Fatalf("Failed to connect to database: %v", err)
			}

			// Get the SQL database itself.
			sqlDb, sqlErr := db.DB()

			// Verify if could get the connection.
			if sqlErr != nil {
				log.Fatalf("Faild to get the SQL connection: %v", err)
			}

			// Set the pool values.
			sqlDb.SetMaxOpenConns(400)
			sqlDb.SetMaxIdleConns(10)
			sqlDb.SetConnMaxLifetime(time.Hour)
			sqlDb.SetConnMaxIdleTime(time.Hour)
		})

	return db, err
}
