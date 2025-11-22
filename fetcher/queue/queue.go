package queue

import (
	mainregionqueue "goleague/fetcher/queue/mainregion"
	subregionqueue "goleague/fetcher/queue/subregion"
	regionmanager "goleague/fetcher/regionmanager"
	"goleague/pkg/regions"
	"log"
	"sync"
)

// StartQueue is the main process of the fetcher.
// Initialize all subregions and main region queues.
func StartQueue(rm *regionmanager.RegionManager) {
	var wg sync.WaitGroup
	// Loop through each main region and start it's queue.
	for mainRegion, subRegions := range regions.RegionList {
		wg.Add(1)

		// Start the queue.
		go func(region regions.MainRegion) {
			defer wg.Done()
			// Create the main region queue instance.
			queue, err := mainregionqueue.NewMainRegionQueue(region, rm)
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
				queue, err := subregionqueue.NewSubRegionQueue(sr, rm)
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
