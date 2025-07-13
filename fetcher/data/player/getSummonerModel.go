package playerfetcher

// SummonerByPuuid is the return of the sub region endpoint.
type SummonerByPuuid struct {
	Id            string `json:"id"`
	Puuid         string `json:"puuid"`
	ProfileIconId int    `json:"profileIconId"`
	SummonerLevel int    `json:"summonerLevel"`
}

// Account is the return of a account search.
type Account struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}
