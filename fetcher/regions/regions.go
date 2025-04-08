package regions

// Simple package containing the region list.
// Separated from the region manager to avoid import cycles.
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
