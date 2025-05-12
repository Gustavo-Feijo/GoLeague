package filters

import (
	"strings"
)

// Query parameters for the tierlist filters.
type GetQueryParams struct {
	Tier      string `form:"tier"`
	Rank      string `form:"rank"`
	Queue     int    `form:"queue"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=championId winRate pickRate banRate"`
	Direction string `form:"direction,default=ascending"`
	Limit     int    `form:"limit,default=100" binding:"omitempty,min=1"`
	Offset    int    `form:"offset,default=0" binding:"omitempty,min=0"`
}

// Get the query parameters as a map.
func (q *GetQueryParams) AsMap() map[string]any {
	filters := make(map[string]any)

	// Set to the default maximum.
	// Could use max on the form, but that would return a error.
	if q.Limit > 100 {
		q.Limit = 100
	}

	// Set pagination data.
	filters["direction"] = q.Direction
	filters["limit"] = q.Limit
	filters["offset"] = q.Offset
	filters["sort"] = q.SortBy

	// Only add non-empty filters
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
