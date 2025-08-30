package filters

// URI params for the player endpoitns.
type PlayerURIParams struct {
	GameName string `uri:"gameName" binding:"required"`
	GameTag  string `uri:"gameTag" binding:"required"`
	Region   string `uri:"region" binding:"required"`
}

// Query parameters for the player search filters.
type PlayerSearchParams struct {
	Name   string `form:"name"`
	Tag    string `form:"tag"`
	Region string `form:"region"`
}

// PlayerSearchFilter is a enforced type for the player search filters.
type PlayerSearchFilter struct {
	Name   string
	Tag    string
	Region string
}

// NewPlayerSearchFilter creates simple search filter.
func NewPlayerSearchFilter(qp PlayerSearchParams) *PlayerSearchFilter {
	return &PlayerSearchFilter{
		Name:   qp.Name,
		Tag:    qp.Tag,
		Region: qp.Region,
	}
}

// Query params for the player match history.
type PlayerMatchHistoryParams struct {
	Page  int `form:"page"`
	Queue int `form:"queue"`
}

// PlayerMatchHistoryFilter is the full structure for a player matchHistoryFilter.
type PlayerMatchHistoryFilter struct {
	GameName string
	GameTag  string
	Region   string
	Page     int
	PlayerId *uint
	Queue    int
}

func NewPlayerMatchHistoryFilter(qp PlayerMatchHistoryParams, pp *PlayerURIParams) *PlayerMatchHistoryFilter {
	return &PlayerMatchHistoryFilter{
		GameName: pp.GameName,
		GameTag:  pp.GameTag,
		Page:     qp.Page,
		Queue:    qp.Queue,
		Region:   pp.Region,
	}
}

// Query params for the player match history.
type PlayerStatsParams struct {
	Interval int `form:"interval"`
}

// PlayerStatsFilter is the full stats filtering for player stats.
type PlayerStatsFilter struct {
	GameName string
	GameTag  string
	Interval int
	PlayerId *uint
	Region   string
}

func NewPlayerStatsFilter(qp PlayerStatsParams, pp *PlayerURIParams) *PlayerStatsFilter {
	return &PlayerStatsFilter{
		GameName: pp.GameName,
		GameTag:  pp.GameTag,
		Interval: qp.Interval,
		Region:   pp.Region,
	}
}

// PlayerForceFetchMatchListFilter is the struct for a force fetch match list call.
type PlayerForceFetchMatchListFilter struct {
	GameName string
	GameTag  string
	Region   string
}

func NewForceFetchMatchHistoryFilter(pp *PlayerURIParams) *PlayerForceFetchMatchListFilter {
	return &PlayerForceFetchMatchListFilter{
		GameName: pp.GameName,
		GameTag:  pp.GameTag,
		Region:   pp.Region,
	}
}

// PlayerForceFetchFilter is the struct for a force fetch player data call.
type PlayerForceFetchFilter struct {
	GameName string
	GameTag  string
	Region   string
}

func NewForceFetchPlayerFilter(pp *PlayerURIParams) *PlayerForceFetchFilter {
	return &PlayerForceFetchFilter{
		GameName: pp.GameName,
		GameTag:  pp.GameTag,
		Region:   pp.Region,
	}
}

// PlayerInfoFilter is the simple struct for holding player info filters.
type PlayerInfoFilter struct {
	GameName string
	GameTag  string
	Region   string
	PlayerId *uint
}

func NewPlayerInfoFilter(pp *PlayerURIParams) *PlayerInfoFilter {
	return &PlayerInfoFilter{
		GameName: pp.GameName,
		GameTag:  pp.GameTag,
		Region:   pp.Region,
	}
}
