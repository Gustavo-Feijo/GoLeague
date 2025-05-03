package jobs

import (
	"goleague/fetcher/assets"
	"log"
)

func RevalidateCache() {
	log.Println("Starting champion cache revalidation")
	_, err := assets.RevalidateChampionCache("en_US", "")
	if err != nil {
		log.Printf("Error revalidating champion cache: %v", err)
	} else {
		log.Println("Champion cache revalidation completed successfully")
	}

	log.Println("Starting item cache revalidation")
	_, err = assets.RevalidateItemCache("en_US", "")
	if err != nil {
		log.Printf("Error revalidating item cache: %v", err)
	} else {
		log.Println("Item cache revalidation completed successfully")
	}
}
