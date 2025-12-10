package filters

// URI params for the champion endpoitns.
type ChampionURIParams struct {
	ChampionId string `uri:"championId" binding:"required"`
}

type GetChampionDataFilter struct {
	ChampionId string
}

func NewGetChampionDataFilter(pp *ChampionURIParams) *GetChampionDataFilter {
	return &GetChampionDataFilter{
		ChampionId: pp.ChampionId,
	}
}
