package database

import (
	"fmt"
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

// Create the enums for the queue, tier and rank.
func CreateEnums(db *gorm.DB) error {
	// Check and create ENUM types if they do not exist
	err = db.Exec(`
		DO $$ 
		BEGIN
		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'queue_type') THEN
		        CREATE TYPE queue_type AS ENUM ('RANKED_SOLO_5x5', 'RANKED_FLEX_SR');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_type') THEN
		        CREATE TYPE tier_type AS ENUM ('IRON', 'BRONZE', 'SILVER', 'GOLD', 'PLATINUM', 'EMERALD', 'DIAMOND', 'MASTER', 'GRANDMASTER', 'CHALLENGER');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'rank_type') THEN
		        CREATE TYPE rank_type AS ENUM ('IV', 'III', 'II', 'I');
		    END IF;
		END $$;
	`).Error

	return err
}

// Create any necessary triggers.
func CreateTriggers(db *gorm.DB) error {
	err := createFetchMatchesTrigger(db)
	if err != nil {
		return fmt.Errorf("couldn't craete the triggers: %v", err)
	}

	return nil
}

// Create the triggers for setting the player matches to be fetched when a new rating is inserted.
// Since rating entries are only created when something changed.
func createFetchMatchesTrigger(db *gorm.DB) error {
	return db.Exec(`
      CREATE OR REPLACE FUNCTION update_unfetched_match()
      RETURNS TRIGGER AS $$
      BEGIN
        UPDATE player_infos
        SET unfetched_match = true 
        WHERE id = NEW.player_id;

        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;

      DROP TRIGGER IF EXISTS update_fetch_match_rating_insert ON rating_entries;
      
      CREATE TRIGGER update_fetch_match_rating_insert
      AFTER INSERT ON rating_entries
      FOR EACH ROW
      EXECUTE FUNCTION update_unfetched_match();
    `).Error
}
