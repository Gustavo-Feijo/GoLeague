package queue

import (
	mainregion_queue "goleague/fetcher/queue/mainregion"
	subregion_queue "goleague/fetcher/queue/subregion"
	"goleague/fetcher/regions"
	"sync"
)

func StartQueue(rm *regions.RegionManager) {
	var wg sync.WaitGroup
	// Loop through each main region and start it's queue.
	for mainRegion, subRegions := range regions.RegionList {
		wg.Add(1)

		// Start the queue.
		go func(region regions.MainRegion) {
			defer wg.Done()
			mainregion_queue.RunMainRegionQueue(region, rm)
		}(mainRegion)

		// Loop through each associated subregion and start it's queue.
		for _, subRegion := range subRegions {
			wg.Add(1)

			// Start the sub region queue.
			go func(sr regions.SubRegion) {
				defer wg.Done()
				subregion_queue.RunSubRegionQueue(sr, rm)
			}(subRegion)
		}
	}
	wg.Wait()
}
