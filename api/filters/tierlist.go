package filters

import (
	tiervalues "goleague/pkg/riotvalues/tier"
)

// Query parameters for the tierlist filters.
type TierlistQueryParams struct {
	Tier      string `form:"tier"`
	Rank      string `form:"rank"`
	Queue     int    `form:"queue"`
	AboveTier bool   `form:"above_tiers"`
}

type TierlistFilter struct {
	Tier          string
	GetTiersAbove bool
	NumericTier   int
	Rank          *string
	Queue         int
}

func NewTierlistFilter(params TierlistQueryParams) *TierlistFilter {
	filters := &TierlistFilter{
		GetTiersAbove: false,
		NumericTier:   0,
	}

	// Only add non-empty filters.
	if params.Tier != "" {
		filters.Tier = params.Tier
	}

	if params.Rank != "" {
		filters.Rank = &params.Rank
	}

	if params.Queue != 0 {
		filters.Queue = params.Queue
	}

	if filters.Tier != "" {
		var rank string
		if filters.Rank == nil {
			rank = "I"
		} else {
			rank = *filters.Rank
		}

		filters.NumericTier = tiervalues.CalculateRank(filters.Tier, rank, 0)
	}

	filters.GetTiersAbove = params.AboveTier

	return filters
}
