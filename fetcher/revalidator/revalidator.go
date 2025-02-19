package main

import (
	"fmt"
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
		log.Fatal("Couldn't fetch the data from the ddragon to revalidate the cache")
	}

	item, err := assets.RevalidateItemCache("en_US", "1001")
	fmt.Println(item)
}
