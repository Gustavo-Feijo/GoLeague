package filters

// Query parameters for the playere search filters.
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
