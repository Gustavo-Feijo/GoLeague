package league_fetcher

// Define the type return by the simple league entries.
type LeagueEntry struct {
	SummonerId   string  `json:"summonerId"`
	Puuid        string  `json:"puuid"`
	Tier         *string `json:"tier,omitempty"`
	Division     *string `json:"division,omitempty"`
	QueueType    *string `json:"queueType,omitempty"`
	LeaguePoints uint16  `json:"leaguePoints"`
	Wins         uint16  `json:"wins"`
	Losses       uint16  `json:"losses"`
	FreshBlood   bool    `json:"freshBlood"`
	HotStreak    bool    `json:"hotStreak"`
}

// The high elo league entries come in a very similar way.
// Only having some outer keys.
type HighEloLeagueEntry struct {
	Entries []LeagueEntry `json:"entries"`
	Tier    string        `json:"tier"`
}
