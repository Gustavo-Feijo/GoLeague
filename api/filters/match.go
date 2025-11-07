package filters

// URI params for the match endpoitns.
type MatchURIParams struct {
	MatchId string `uri:"matchId" binding:"required"`
}

type GetFullMatchDataFilter struct {
	MatchId string
}

func NewGetFullMatchDataFilter(pp *MatchURIParams) *GetFullMatchDataFilter {
	return &GetFullMatchDataFilter{
		MatchId: pp.MatchId,
	}
}
