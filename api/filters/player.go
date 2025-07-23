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

// Path params for the player force fetch.
type PlayerForceFetchParams struct {
	GameName string `uri:"gameName" binding:"required"`
	GameTag  string `uri:"gameTag" binding:"required"`
	Region   string `uri:"region" binding:"required"`
}

// Get the path params as a map.
func (q *PlayerForceFetchParams) AsMap() map[string]any {
	return map[string]any{
		"gameName": q.GameName,
		"gameTag":  q.GameTag,
		"region":   q.Region,
	}
}

// Path params for the player force fetch.
type PlayerForceFetchMatchHistoryParams struct {
	GameName string `uri:"gameName" binding:"required"`
	GameTag  string `uri:"gameTag" binding:"required"`
	Region   string `uri:"region" binding:"required"`
}

// Get the path params as a map.
func (q *PlayerForceFetchMatchHistoryParams) AsMap() map[string]any {
	return map[string]any{
		"gameName": q.GameName,
		"gameTag":  q.GameTag,
		"region":   q.Region,
	}
}
