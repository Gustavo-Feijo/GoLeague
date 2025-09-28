package filters

// URI params for the match endpoitns.
type MatchURIParams struct {
	MatchId string `uri:"matchId" binding:"required"`
}
