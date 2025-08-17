package dto

// PlayerSearch is the type of a player search result.
type PlayerSearch struct {
	Id            uint
	Name          string
	ProfileIcon   int
	Puuid         string
	Region        string
	SummonerLevel int
	Tag           string
}

// StatsEntry is a single stat entry to be returned.
type StatsEntry struct {
	Matches        int     `json:"matches"`
	WinRate        float32 `json:"win_rate"`
	AverageKills   float32 `json:"average_kills"`
	AverageDeaths  float32 `json:"average_deaths"`
	AverageAssists float32 `json:"average_assists"`
	CsPerMin       float32 `json:"cs_per_min"`
	KDA            float32 `json:"kda"`
}

// PlayerStatsQueue represents different metrics for a given player stats in a given queueId.
type PlayerStatsQueue struct {
	ChampionData map[string]*StatsEntry `json:"champions"`
	LaneData     map[string]*StatsEntry `json:"lanes"`
	Unfiltered   *StatsEntry            `json:"unfiltered"`
}

// FullPlayerStats is a map of QueueIds and the respective stats.
type FullPlayerStats map[string]*PlayerStatsQueue
