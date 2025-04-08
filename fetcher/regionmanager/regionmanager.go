package regionmanager

import (
	"fmt"
	"goleague/fetcher/data"
	mainregion_processor "goleague/fetcher/data/processors/mainregion"
	subregion_processor "goleague/fetcher/data/processors/subregion"
	"goleague/fetcher/regions"
	"log"
	"sync"
)

var (
	regionManagerInstance *RegionManager
	regionManagerOnce     sync.Once
)

// Define the region manager.
type RegionManager struct {
	// Get which main region is the parent of the sub region.
	subToMain map[regions.SubRegion]regions.MainRegion

	// Get which sub regions are children of the main region.
	mainToSub map[regions.MainRegion][]regions.SubRegion

	// List of fetchers the main regions.
	mainProcessor map[regions.MainRegion]*mainregion_processor.MainRegionProcessor

	// List of fetchers the sub regions.
	subProcessor map[regions.SubRegion]*subregion_processor.SubRegionProcessor

	mu sync.RWMutex
}

// Create a region manager for the RiotAPI and populate it.
func GetRegionManager() *RegionManager {
	// Singleton for creating only one region manager.
	regionManagerOnce.Do(func() {
		// Create the region manager.
		regionManagerInstance = &RegionManager{
			subToMain:     make(map[regions.SubRegion]regions.MainRegion),
			mainToSub:     make(map[regions.MainRegion][]regions.SubRegion),
			mainProcessor: make(map[regions.MainRegion]*mainregion_processor.MainRegionProcessor),
			subProcessor:  make(map[regions.SubRegion]*subregion_processor.SubRegionProcessor),
		}

		// Loop through each region and populate it.
		for MainRegion, SubRegions := range regions.RegionList {
			// Create the relationship between the main regions and the regions.SubRegions.
			regionManagerInstance.mainToSub[MainRegion] = SubRegions

			// Create  the main region fetcher.
			fetcher := data.CreateMainFetcher(string(MainRegion))

			// Create the processor.
			processor, err := mainregion_processor.CreateMainRegionProcessor(fetcher, MainRegion)
			if err != nil {
				log.Fatalf("Couldn't create the processor for region %s: %v", MainRegion, err)
			}

			regionManagerInstance.mainProcessor[MainRegion] = processor

			for _, SubRegion := range SubRegions {
				// Save the parent of this sub regions.
				regionManagerInstance.subToMain[SubRegion] = MainRegion

				// Create the sub region fetcher.
				fetcher := data.CreateSubFetcher(string(SubRegion))

				// Create the processor.
				processor, err := subregion_processor.CreateSubRegionProcessor(fetcher, SubRegion)
				if err != nil {
					log.Fatalf("Couldn't create the processor for region %s: %v", SubRegion, err)
				}

				regionManagerInstance.subProcessor[SubRegion] = processor
			}
		}
	})

	return regionManagerInstance
}

// Get the fetcher for a given region
func (m *RegionManager) GetMainProcessor(region regions.MainRegion) (*mainregion_processor.MainRegionProcessor, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the processor exists and get it.
	processor, exists := m.mainProcessor[region]
	if !exists {
		return nil, fmt.Errorf("the main region %s doesn't exist", region)
	}

	return processor, nil
}

// Get the processor for a given region
func (m *RegionManager) GetSubProcessor(region regions.SubRegion) (*subregion_processor.SubRegionProcessor, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the processor exists and get it.
	processor, exists := m.subProcessor[region]
	if !exists {
		return nil, fmt.Errorf("the sub region %s doesn't exist", region)
	}

	return processor, nil
}

// Get the parent of a sub region.
func (m *RegionManager) GetMainRegion(SubRegion regions.SubRegion) (regions.MainRegion, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the sub region exists and get it.
	MainRegion, exists := m.subToMain[SubRegion]
	if !exists {
		return "", fmt.Errorf("the region %s doesn't exist or isn't a sub region", SubRegion)
	}

	return MainRegion, nil
}

// Get all child sub regions to a given main region.
func (m *RegionManager) GetSubRegions(MainRegion regions.MainRegion) ([]regions.SubRegion, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the sub region exists and get it.
	SubRegions, exists := m.mainToSub[MainRegion]
	if !exists {
		return nil, fmt.Errorf("the region %s doesn't exist or isn't a main region", MainRegion)
	}

	return SubRegions, nil
}
