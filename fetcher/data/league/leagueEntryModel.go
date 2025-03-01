package league_fetcher

// Define the type return by the simple league entries.
// The tier, rank and queue type doesn't need to be stored since they will be passed to fetch it.
type LeagueEntry struct {
	SummonerId   string `json:"summonerId"`
	Puuid        string `json:"puuid"`
	LeaguePoints uint16 `json:"leaguePoints"`
	Wins         uint16 `json:"wins"`
	Losses       uint16 `json:"losses"`
	FreshBlood   bool   `json:"freshBlood"`
	HotStreak    bool   `json:"hotStreak"`
}
