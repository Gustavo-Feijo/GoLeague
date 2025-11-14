package database

import (
	"database/sql"
	"fmt"
	"goleague/pkg/config"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending migrations to the database.
func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.Database.MigrationsPath),
		config.Database.Database,
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	// Acquire an advisory lock to prevent concurrent migrations between services.
	var lockAcquired bool
	lockKey := "goleague_migrations_lock"
	err = db.QueryRow("SELECT pg_try_advisory_lock(hashtext($1))", lockKey).Scan(&lockAcquired)
	if err != nil {
		return err
	}

	if !lockAcquired {
		log.Println("Another process is already running migrations, skipping...")
		return nil
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	var lockReleased bool
	err = db.QueryRow("SELECT pg_advisory_unlock(hashtext($1))", lockKey).Scan(&lockReleased)
	if err != nil || !lockReleased {
		return fmt.Errorf("could not release advisory lock: %w", err)
	}

	return nil
}
