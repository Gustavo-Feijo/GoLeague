package regions

import (
	"fmt"
	"goleague/fetcher/data"
	"sync"
)

var (
	regionManagerInstance *RegionManager
	regionManagerOnce     sync.Once
)

// Create the types for clarity.
type (
	MainRegion string
	SubRegion  string
)

// List of regions.
var RegionList = map[MainRegion][]SubRegion{
	"AMERICAS": {"BR1", "LA1", "LA2", "NA1"},
	"EUROPE":   {"EUN1", "EUW1", "TR1", "ME1", "RU"},
	"ASIA":     {"KR", "JP1"},
	"SEA":      {"OC1", "SG2", "TW2", "VN2"},
}

// Define the region manager.
type RegionManager struct {
	// Get which main region is the parent of the sub region.
	subToMain map[SubRegion]MainRegion

	// Get which sub regions are children of the main region.
	mainToSub map[MainRegion][]SubRegion

	// List of fetchers the main regions.
	mainFetcher map[MainRegion]*data.MainFetcher

	// List of fetchers the sub regions.
	subFetcher map[SubRegion]*data.SubFetcher

	mu sync.RWMutex
}

// Create a region manager for the RiotAPI and populate it.
func GetRegionManager() *RegionManager {
	// Singleton for creating only one region manager.
	regionManagerOnce.Do(func() {
		// Create the region manager.
		regionManagerInstance = &RegionManager{
			subToMain:   make(map[SubRegion]MainRegion),
			mainToSub:   make(map[MainRegion][]SubRegion),
			mainFetcher: make(map[MainRegion]*data.MainFetcher),
			subFetcher:  make(map[SubRegion]*data.SubFetcher),
		}

		// Loop through each region and populate it.
		for mainRegion, subRegions := range RegionList {
			// Create the relationship between the main regions and the subregions.
			regionManagerInstance.mainToSub[mainRegion] = subRegions

			// Create each main region fetcher.
			regionManagerInstance.mainFetcher[mainRegion] = data.CreateMainFetcher(string(mainRegion))

			for _, subRegion := range subRegions {
				// Save the parent of this sub regions.
				regionManagerInstance.subToMain[subRegion] = mainRegion

				// Create the sub regions fetchers.
				regionManagerInstance.subFetcher[subRegion] = data.CreateSubFetcher(string(subRegion))
			}
		}
	})

	return regionManagerInstance
}

// Get the fetcher for a given region
func (m *RegionManager) GetMainFetcher(region MainRegion) (*data.MainFetcher, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the fetcher exists and get it.
	fetcher, exists := m.mainFetcher[region]
	if !exists {
		return nil, fmt.Errorf("the main region %s doesn't exist", region)
	}

	return fetcher, nil
}

// Get the fetcher for a given region
func (m *RegionManager) GetSubFetcher(region SubRegion) (*data.SubFetcher, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the fetcher exists and get it.
	fetcher, exists := m.subFetcher[region]
	if !exists {
		return nil, fmt.Errorf("the sub region %s doesn't exist", region)
	}

	return fetcher, nil
}

// Get the parent of a sub region.
func (m *RegionManager) GetMainRegion(subRegion SubRegion) (MainRegion, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the sub region exists and get it.
	mainRegion, exists := m.subToMain[subRegion]
	if !exists {
		return "", fmt.Errorf("the region %s doesn't exist or isn't a sub region", subRegion)
	}

	return mainRegion, nil
}

// Get all child sub regions to a given main region.
func (m *RegionManager) GetSubRegions(mainRegion MainRegion) ([]SubRegion, error) {
	// Lock for reading.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify if the sub region exists and get it.
	subRegions, exists := m.mainToSub[mainRegion]
	if !exists {
		return nil, fmt.Errorf("the region %s doesn't exist or isn't a main region", mainRegion)
	}

	return subRegions, nil
}
