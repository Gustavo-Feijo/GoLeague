package playerfetcher

// SummonerByPuuid is the return of the sub region endpoint.
type SummonerByPuuid struct {
	Id            string `json:"id"`
	AccountId     string `json:"accountId"`
	Puuid         string `json:"puuid"`
	ProfileIconId int    `json:"profileIconId"`
	SummonerLevel int    `json:"summonerLevel"`
}
