package dto

// PlayerSearch is the type of a player search result.
type PlayerSearch struct {
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	ProfileIcon   int    `json:"profileIconId"`
	Puuid         string `json:"puuid"`
	Region        string `json:"region"`
	SummonerLevel int    `json:"summonerLevel"`
	Tag           string `json:"tag"`
}

// StatsEntry is a single stat entry to be returned.
type StatsEntry struct {
	Matches        int     `json:"matches"`
	WinRate        float32 `json:"winRate"`
	AverageKills   float32 `json:"averageKills"`
	AverageDeaths  float32 `json:"averageDeaths"`
	AverageAssists float32 `json:"averageAssists"`
	CsPerMin       float32 `json:"csPerMin"`
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

// FullPlayerInfo is the the DTO of a given player info.
type FullPlayerInfo struct {
	Id            uint         `json:"id"`
	Name          string       `json:"name"`
	ProfileIcon   int          `json:"profileIconId"`
	Puuid         string       `json:"puuid"`
	Region        string       `json:"region"`
	SummonerLevel int          `json:"summonerLevel"`
	Tag           string       `json:"tag"`
	Rating        []RatingInfo `json:"rating"`
}

// RatingInfo contains a player rating information for a given queue at a given region.
type RatingInfo struct {
	Queue        string `json:"queue"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"lp"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Region       string `json:"region"`
}
