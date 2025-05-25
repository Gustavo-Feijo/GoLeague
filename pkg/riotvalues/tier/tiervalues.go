package tiervalues

import (
	"slices"
	"strings"
)

var tierValues = map[string]int{
	"IRON":        0,
	"BRONZE":      10000,
	"SILVER":      20000,
	"GOLD":        30000,
	"PLATINUM":    40000,
	"EMERALD":     50000,
	"DIAMOND":     60000,
	"MASTER":      70000,
	"GRANDMASTER": 80000,
	"CHALLENGER":  90000,
}

var rankValues = map[string]int{
	"IV":  0,
	"III": 2500,
	"II":  5000,
	"I":   7500,
}

// Calculate numeric rank from tier and division.
func CalculateRank(tier string, rank string, lp int) int {
	// Normalize the tier entry.
	tier = strings.ToUpper(tier)
	tier = strings.TrimSpace(tier)

	baseValue, exists := tierValues[tier]
	if !exists {
		return 0 // Unknown tier
	}

	// Normalize the rank entry.
	rank = strings.ToUpper(rank)
	rank = strings.TrimSpace(rank)

	// Division 4 is lowest, 1 is highest, each worth 100 points.
	divisionValue, exists := rankValues[rank]
	if !exists {
		return baseValue
	}

	// Don't add the division value if it's a highelo.
	if slices.Contains([]string{"MASTER", "GRANDMASTER", "CHALLENGER"}, tier) {
		divisionValue = 0
	}

	// Return the sum of the ratings and lp.
	return baseValue + divisionValue + lp
}
