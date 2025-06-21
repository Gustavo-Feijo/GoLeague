package tiervalues

import (
	"fmt"
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

// Pre-sorted slices for better lookup.
var tierNames = []string{"IRON", "BRONZE", "SILVER", "GOLD", "PLATINUM", "EMERALD", "DIAMOND", "MASTER", "GRANDMASTER", "CHALLENGER"}
var rankNames = []string{"IV", "III", "II", "I"}

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

// CalculateInverseRank takes a numeric value and returns the closest tier and rank.
func CalculateInverseRank(value int) string {
	if value < 0 {
		return "IRON IV"
	}

	// Go through each tier, if the value is greater, then it's the right one.
	tierIndex := 0
	for i := len(tierNames) - 1; i >= 0; i-- {
		if value >= tierValues[tierNames[i]] {
			tierIndex = i
			break
		}
	}

	tier := tierNames[tierIndex]

	// Early return if it's a elo without division.
	if slices.Contains([]string{"MASTER", "GRANDMASTER", "CHALLENGER"}, tier) {
		return tier
	}

	remainingValue := value - tierValues[tier]

	// Apply same logic for ranks.
	rankIndex := 0
	for i := len(rankNames) - 1; i >= 0; i-- {
		if remainingValue >= rankValues[rankNames[i]] {
			rankIndex = i
			break
		}
	}

	return fmt.Sprintf("%s %s", tier, rankNames[rankIndex])
}
