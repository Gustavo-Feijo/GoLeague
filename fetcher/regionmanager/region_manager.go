package regionmanager

import (
	"fmt"
	"goleague/fetcher/data"
	"goleague/fetcher/regions"
	mainregionservice "goleague/fetcher/services/mainregion"
	subregionservice "goleague/fetcher/services/subregion"
	"sync"

	"gorm.io/gorm"
)

type RegionManagerDependencies struct {
	DB *gorm.DB
}

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

	// Passed necessary dependencies.
	deps RegionManagerDependencies

	mu sync.RWMutex
}

// NewRegionManager creates a new region manager instance.
func NewRegionManager(deps RegionManagerDependencies) (*RegionManager, error) {
	rm := &RegionManager{
		subToMain:   make(map[regions.SubRegion]regions.MainRegion),
		mainToSub:   make(map[regions.MainRegion][]regions.SubRegion),
		mainService: make(map[regions.MainRegion]*mainregionservice.MainRegionService),
		subService:  make(map[regions.SubRegion]*subregionservice.SubRegionService),
		deps:        deps,
	}

	if err := rm.initialize(); err != nil {
		return nil, fmt.Errorf("couldn't initialize the region manager: %w", err)
	}

	return rm, nil
}

// initialize creates the instances for the main regions and subregions.
func (rm *RegionManager) initialize() error {
	for mainRegion, subRegions := range regions.RegionList {
		if err := rm.initializeMainRegion(mainRegion, subRegions); err != nil {
			return fmt.Errorf("failed to initialize main region %s: %w", mainRegion, err)
		}

		for _, subRegion := range subRegions {
			if err := rm.initializeSubRegion(subRegion, mainRegion); err != nil {
				return fmt.Errorf("failed to initialize sub region %s: %w", subRegion, err)
			}
		}
	}
	return nil
}

// initializeMainRegion creates the main fetcher and service for a main region.
func (rm *RegionManager) initializeMainRegion(mainRegion regions.MainRegion, subRegions []regions.SubRegion) error {
	// Create the relationship between main and sub regions
	rm.mainToSub[mainRegion] = subRegions

	// Create the main region fetcher
	fetcher := data.NewMainFetcher(string(mainRegion))

	// Create the service
	service, err := mainregionservice.NewMainRegionService(rm.deps.DB, fetcher, mainRegion)
	if err != nil {
		return fmt.Errorf("couldn't create service: %w", err)
	}

	rm.mainService[mainRegion] = service
	return nil
}

// initializeSubRegion creates the sub region fetcher and service.
func (rm *RegionManager) initializeSubRegion(subRegion regions.SubRegion, mainRegion regions.MainRegion) error {
	// Save the parent of this sub region
	rm.subToMain[subRegion] = mainRegion

	// Create the sub region fetcher
	fetcher := data.NewSubFetcher(string(subRegion))

	// Create the service
	service, err := subregionservice.NewSubRegionService(rm.deps.DB, fetcher, subRegion)
	if err != nil {
		return fmt.Errorf("couldn't create service: %w", err)
	}

	rm.subService[subRegion] = service
	return nil
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
