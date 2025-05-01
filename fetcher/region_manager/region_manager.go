package regionmanager

import (
	"fmt"
	"goleague/fetcher/data"
	"goleague/fetcher/regions"
	mainregion_service "goleague/fetcher/services/main_region"
	subregion_service "goleague/fetcher/services/sub_region"
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
	mainService map[regions.MainRegion]*mainregion_service.MainRegionService

	// List of fetchers the sub regions.
	subService map[regions.SubRegion]*subregion_service.SubRegionService

	mu sync.RWMutex
}

// Create a region manager for the RiotAPI and populate it.
func GetRegionManager() *RegionManager {
	// Singleton for creating only one region manager.
	regionManagerOnce.Do(func() {
		// Create the region manager.
		regionManagerInstance = &RegionManager{
			subToMain:   make(map[regions.SubRegion]regions.MainRegion),
			mainToSub:   make(map[regions.MainRegion][]regions.SubRegion),
			mainService: make(map[regions.MainRegion]*mainregion_service.MainRegionService),
			subService:  make(map[regions.SubRegion]*subregion_service.SubRegionService),
		}

		// Loop through each region and populate it.
		for MainRegion, SubRegions := range regions.RegionList {
			// Create the relationship between the main regions and the regions.SubRegions.
			regionManagerInstance.mainToSub[MainRegion] = SubRegions

			// Create  the main region fetcher.
			fetcher := data.CreateMainFetcher(string(MainRegion))

			// Create the service.
			service, err := mainregion_service.NewMainRegionService(fetcher, MainRegion)
			if err != nil {
				log.Fatalf("Couldn't create the service for region %s: %v", MainRegion, err)
			}

			regionManagerInstance.mainService[MainRegion] = service

			for _, SubRegion := range SubRegions {
				// Save the parent of this sub regions.
				regionManagerInstance.subToMain[SubRegion] = MainRegion

				// Create the sub region fetcher.
				fetcher := data.CreateSubFetcher(string(SubRegion))

				// Create the service.
				service, err := subregion_service.CreateSubRegionService(fetcher, SubRegion)
				if err != nil {
					log.Fatalf("Couldn't create the service for region %s: %v", SubRegion, err)
				}

				regionManagerInstance.subService[SubRegion] = service
			}
		}
	})

	return regionManagerInstance
}

// Get the fetcher for a given region
func (m *RegionManager) GetMainService(region regions.MainRegion) (*mainregion_service.MainRegionService, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the service exists and get it.
	service, exists := m.mainService[region]
	if !exists {
		return nil, fmt.Errorf("the main region %s doesn't exist", region)
	}

	return service, nil
}

// Get the service for a given region
func (m *RegionManager) GetSubService(region regions.SubRegion) (*subregion_service.SubRegionService, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the service exists and get it.
	service, exists := m.subService[region]
	if !exists {
		return nil, fmt.Errorf("the sub region %s doesn't exist", region)
	}

	return service, nil
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
