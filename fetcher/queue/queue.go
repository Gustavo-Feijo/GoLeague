package queue

import (
	"goleague/fetcher/regions"
	"log"
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
			runMainRegionQueue(region, rm)
		}(mainRegion)

		// Loop through each associated subregion and start it's queue.
		for _, subRegion := range subRegions {
			wg.Add(1)

			// Start the sub region queue.
			go func(sr regions.SubRegion) {
				defer wg.Done()
				runSubRegionQueue(sr, rm)
			}(subRegion)
		}
	}
	wg.Wait()
}

func runMainRegionQueue(region regions.MainRegion, rm *regions.RegionManager) {
	_, err := rm.GetMainFetcher(region)
	if err != nil {
		log.Printf("Failed to get main region fetcher for %v: %v", region, err)
		return
	}
}
