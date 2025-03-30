package main

import (
	"goleague/fetcher/assets"
	"goleague/pkg/config"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load the env and just revalidate the entire champion cache.
// Will be executed in a regular basis.
func main() {
	if os.Getenv("ENVIRONMENT") != "docker" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	config.LoadEnv()
	_, err := assets.RevalidateChampionCache("en_US", "")
	if err != nil {
		log.Fatal("Couldn't fetch the data from the ddragon to revalidate the champion cache")
	}

	// Revalidate all the items.
	_, err = assets.RevalidateItemCache("en_US", "")
	if err != nil {
		log.Fatal("Couldn't fetch the data from the ddragon to revalidate the item cache")
	}
}
