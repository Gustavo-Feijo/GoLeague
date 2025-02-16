package main

import (
	"goleague/fetcher/assets"
	"goleague/pkg/config"
	"log"
)

// Load the env and just revalidate the entire champion cache.
// Will be executed in a regular basis.
func main() {
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
