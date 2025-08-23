package filters

// Query parameters for the player search filters.
type PlayerSearchParams struct {
	Name   string `form:"name"`
	Tag    string `form:"tag"`
	Region string `form:"region"`
}

// Get the query parameters as a map.
func (q *PlayerSearchParams) AsMap() map[string]any {
	return map[string]any{
		"name":   q.Name,
		"tag":    q.Tag,
		"region": q.Region,
	}
}

// Query params for the player match history.
type PlayerMatchHistoryParams struct {
	Page  int `form:"page"`
	Queue int `form:"queue"`
}

// Get the query parameters as a map.
func (q *PlayerMatchHistoryParams) AsMap() map[string]any {
	return map[string]any{
		"page":  q.Page,
		"queue": q.Queue,
	}
}

// Query params for the player match history.
type PlayerStatsParams struct {
	Interval int `form:"interval"`
}

// Get the query parameters as a map.
func (q *PlayerStatsParams) AsMap() map[string]any {
	return map[string]any{
		"interval": q.Interval,
	}
}

// URI params for the player endpoitns.
type PlayerURIParams struct {
	GameName string `uri:"gameName" binding:"required"`
	GameTag  string `uri:"gameTag" binding:"required"`
	Region   string `uri:"region" binding:"required"`
}

// Get the path params as a map.
func (q *PlayerURIParams) AsMap() map[string]any {
	return map[string]any{
		"gameName": q.GameName,
		"gameTag":  q.GameTag,
		"region":   q.Region,
	}
}
