package queue

import (
	"goleague/fetcher/regions"
	"goleague/pkg/database/models"
	"log"
	"time"
)

// Run the sub region queue.
// Mainly responsible for getting the ratings for each player on the region.
func runSubRegionQueue(region regions.SubRegion, rm *regions.RegionManager) {
	fetcher, err := rm.GetSubFetcher(region)
	if err != nil {
		log.Printf("Failed to get main region fetcher for %v: %v", region, err)
		return
	}

	ratingService, err := models.CreateRatingService()
	if err != nil {
		log.Printf("Failed to start the rating service for the %v: %v", region, err)
		return
	}

	playerService, err := models.CreatePlayerService()
	if err != nil {
		log.Printf("Failed to start the player service for the %v: %v", region, err)
		return
	}

	// Run for each queue.
	for _, queue := range queues {
		// First, get the high elo players.
		for _, highElo := range highElos {
			highRating, err := fetcher.League.GetHighEloLeagueEntries(highElo, queue)
			if err != nil {
				log.Printf("Couldn't get the high elo league %v: %v", highElo, err)
				continue
			}

			for _, entry := range highRating.Entries {
				// For high elo we don't have the tier inside the entries array, so we set manually.
				entry.Tier = &highRating.Tier

				// Verify if the player exists in the database.
				player, err := playerService.GetPlayerByPuuid(entry.Puuid)
				if err != nil {
					log.Printf("Couldn't get the player with PUUID: %v", entry.Puuid)
					continue
				}

				// The player doesn't exist.
				if player == nil {
					// Reassign the player to the newly created one.
					player, err = playerService.CreatePlayerFromRating(entry, region)
					if err != nil {
						log.Printf("Couldn't create the player with PUUID %v on the region %v: %v", entry.Puuid, region, err)
						continue
					}

					log.Printf("Created player with id %v and PUUID %v in region %v", player.ID, player.Puuid, region)
				}

				// Get the last rating of this player.
				// Used to verify if the player has played a match recently.
				lastRating, err := ratingService.GetLastRatingEntryByPlayerIdAndQueue(player.ID, queue)
				if err != nil {
					log.Printf("Couldn't get the last rating for the player %v: %v", player.ID, err)
					continue
				}

				// Finally create the rating entry.
				newRating, err := ratingService.CreateRatingEntry(entry, player.ID, region, queue, lastRating)
				if err != nil {
					log.Printf("Couldn't insert the new rating entry for the player %v: %v", player.ID, err)
					continue
				}

				// Verify if something changed.
				if newRating == nil {
					log.Printf("Nothing changed the rating of the player %v on the region %v, last rating: %v", player.ID, region, lastRating.ID)
				} else {
					log.Printf("Created rating entry for the player %v on the region %v", player.ID, region)
				}
			}
		}
	}

	log.Printf("Sleeping for 10 minutes...")

	// Sleep for 10 minutes to wait new matches to happen.
	time.Sleep(10 * time.Minute)
}
