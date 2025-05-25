package filters

import (
	"strings"
)

// Query parameters for the tierlist filters.
type GetQueryParams struct {
	Tier  string `form:"tier"`
	Rank  string `form:"rank"`
	Queue int    `form:"queue"`
}

// Get the query parameters as a map.
func (q *GetQueryParams) AsMap() map[string]any {
	filters := make(map[string]any)

	// Only add non-empty filters.
	if q.Tier != "" {
		tierValue := q.Tier

		// Verify if it's getting a specific tier or all tiers above it.
		if strings.HasSuffix(tierValue, "+") {
			tierValue = strings.TrimSuffix(tierValue, "+")
			filters["tier_higher"] = true
		}

		filters["tier"] = tierValue
	}

	if q.Rank != "" {
		filters["rank"] = q.Rank
	}

	if q.Queue != 0 {
		filters["queue"] = q.Queue
	}

	return filters
}
