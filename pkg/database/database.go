package database

import (
	"fmt"
	"goleague/pkg/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateEnums create the enums for the queue, tier and rank.
func CreateEnums(db *gorm.DB) error {
	// Check and create ENUM types if they do not exist.
	err := db.Exec(`
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

// createFetchMatchesTrigger create the triggers for setting the player matches to be fetched when a new rating is inserted.
// New rating entries are only created when something changed, meaning the player has matches to fetch.
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

// CreateTriggers create all necessary triggers for the database.
func CreateTriggers(db *gorm.DB) error {
	err := createFetchMatchesTrigger(db)
	if err != nil {
		return fmt.Errorf("couldn't create the triggers: %v", err)
	}

	return nil
}

// CreateCustomIndexes creates any necessary custom index.
func CreateCustomIndexes(db *gorm.DB) error {
	// Creates a index for improving player searching time.
	searchIndex := `
		CREATE INDEX IF NOT EXISTS idx_player_search_all ON player_infos (
		  region, 
		  riot_id_game_name text_pattern_ops, 
		  riot_id_tagline text_pattern_ops
		) WHERE riot_id_game_name != '';`
	return db.Exec(searchIndex).Error
}

// GetConnections is a singleton implementation of the database.
// Return the connection pool.
func NewConnection() (*gorm.DB, error) {

	// Create the database instance.
	db, err := gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{
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
