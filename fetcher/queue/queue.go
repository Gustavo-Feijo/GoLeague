package queue

import (
	mainregion_queue "goleague/fetcher/queue/mainregion"
	subregion_queue "goleague/fetcher/queue/subregion"
	"goleague/fetcher/regionmanager"
	"goleague/fetcher/regions"
	"log"
	"sync"
)

func StartQueue(rm *regionmanager.RegionManager) {
	var wg sync.WaitGroup
	// Loop through each main region and start it's queue.
	for mainRegion, subRegions := range regions.RegionList {
		wg.Add(1)

		// Start the queue.
		go func(region regions.MainRegion) {
			defer wg.Done()
			// Create the main region queue instance.
			queue, err := mainregion_queue.CreateMainRegionQueue(region, rm)
			if err != nil {
				log.Printf("Something went wrong at queue start for region %s: %v", region, err)
				return
			}

			queue.Run()
		}(mainRegion)

		// Loop through each associated subregion and start it's queue.
		for _, subRegion := range subRegions {
			wg.Add(1)

			// Start the sub region queue.
			go func(sr regions.SubRegion) {
				defer wg.Done()

				// Create the subregion queue instance.
				queue, err := subregion_queue.CreateSubRegionQueue(sr, rm)
				if err != nil {
					log.Printf("Something went wrong at queue start for subregion %s: %v", sr, err)
					return
				}

				queue.Run()
			}(subRegion)
		}
	}
	wg.Wait()
}
