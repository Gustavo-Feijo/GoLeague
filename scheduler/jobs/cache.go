package jobs

import (
	"fmt"
	"goleague/fetcher/assets"
	"goleague/pkg/database"
	"goleague/pkg/redis"
	"log"
)

func RevalidateCache() error {
	log.Println("Starting champion cache revalidation")
	// Create a new connection pool.
	db, err := database.NewConnection()
	if err != nil {
		return fmt.Errorf("couldn't get database connection: %w", err)
	}

	redis, err := redis.NewClient()
	if err != nil {
		return fmt.Errorf("couldn't get redis connection: %w", err)
	}

	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		redis.Close()
	}()

	err = assets.RevalidateChampionCache(redis, db, "en_US")
	if err != nil {
		log.Printf("Error revalidating champion cache: %v", err)
	} else {
		log.Println("Champion cache revalidation completed successfully")
	}

	log.Println("Starting item cache revalidation")
	err = assets.RevalidateItemCache(redis, db, "en_US")
	if err != nil {
		log.Printf("Error revalidating item cache: %v", err)
	} else {
		log.Println("Item cache revalidation completed successfully")
	}

	return nil
}
