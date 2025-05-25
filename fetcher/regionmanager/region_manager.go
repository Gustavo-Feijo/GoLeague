package regionmanager

import (
	"fmt"
	"goleague/fetcher/data"
	"goleague/fetcher/regions"
	mainregionservice "goleague/fetcher/services/mainregion"
	subregionservice "goleague/fetcher/services/subregion"
	"log"
	"sync"
)

var (
	regionManagerInstance *RegionManager
	regionManagerOnce     sync.Once
)

// RegionManager is the centralized region manager, with all embedded services.
type RegionManager struct {
	// Get which main region is the parent of the sub region.
	subToMain map[regions.SubRegion]regions.MainRegion

	// Get which sub regions are children of the main region.
	mainToSub map[regions.MainRegion][]regions.SubRegion

	// List of fetchers the main regions.
	mainService map[regions.MainRegion]*mainregionservice.MainRegionService

	// List of fetchers the sub regions.
	subService map[regions.SubRegion]*subregionservice.SubRegionService

	mu sync.RWMutex
}

// GetRegionManager is a sigleton for the region manager for the RiotAPI.
// Creates and populates it if not created yet.
func GetRegionManager() *RegionManager {
	// Singleton for creating only one region manager.
	regionManagerOnce.Do(func() {
		// Create the region manager.
		regionManagerInstance = &RegionManager{
			subToMain:   make(map[regions.SubRegion]regions.MainRegion),
			mainToSub:   make(map[regions.MainRegion][]regions.SubRegion),
			mainService: make(map[regions.MainRegion]*mainregionservice.MainRegionService),
			subService:  make(map[regions.SubRegion]*subregionservice.SubRegionService),
		}

		// Loop through each region and populate it.
		for MainRegion, SubRegions := range regions.RegionList {
			// Create the relationship between the main regions and the regions.SubRegions.
			regionManagerInstance.mainToSub[MainRegion] = SubRegions

			// Create  the main region fetcher.
			fetcher := data.NewMainFetcher(string(MainRegion))

			// Create the service.
			service, err := mainregionservice.NewMainRegionService(fetcher, MainRegion)
			if err != nil {
				log.Fatalf("Couldn't create the service for region %s: %v", MainRegion, err)
			}

			regionManagerInstance.mainService[MainRegion] = service

			for _, SubRegion := range SubRegions {
				// Save the parent of this sub regions.
				regionManagerInstance.subToMain[SubRegion] = MainRegion

				// Create the sub region fetcher.
				fetcher := data.NewSubFetcher(string(SubRegion))

				// Create the service.
				service, err := subregionservice.NewSubRegionService(fetcher, SubRegion)
				if err != nil {
					log.Fatalf("Couldn't create the service for region %s: %v", SubRegion, err)
				}

				regionManagerInstance.subService[SubRegion] = service
			}
		}
	})

	return regionManagerInstance
}

// GetMainService returns the service for a Main Region.
func (m *RegionManager) GetMainService(region regions.MainRegion) (*mainregionservice.MainRegionService, error) {
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

// GetSubService returns the service for a Sub Region.
func (m *RegionManager) GetSubService(region regions.SubRegion) (*subregionservice.SubRegionService, error) {
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

// GetMainRegion returns the parent of a Sub Region.
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

// GetSubRegions returns all Sub Regions that are 'children' of a Main Region.
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
