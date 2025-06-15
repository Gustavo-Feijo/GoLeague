package filters

// Query parameters for the player search filters.
type PlayerSearchParams struct {
	Name   string `form:"name"`
	Tag    string `form:"tag"`
	Region string `form:"region"`
}

// Get the query parameters as a map.
func (q *PlayerSearchParams) AsMap() map[string]any {
	filters := make(map[string]any)

	filters["name"] = q.Name
	filters["tag"] = q.Tag
	filters["region"] = q.Region

	return filters
}

// Query params for the player match history.
type PlayerMatchHistoryParams struct {
	Page  int    `form:"page"`
	Queue string `form:"queue"`
}

// Get the query parameters as a map.
func (q *PlayerMatchHistoryParams) AsMap() map[string]any {
	filters := make(map[string]any)

	filters["page"] = q.Page
	filters["queue"] = q.Queue

	return filters
}
