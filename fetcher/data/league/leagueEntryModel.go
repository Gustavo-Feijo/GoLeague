package league_fetcher

// Define the type return by the simple league entries.
type LeagueEntry struct {
	SummonerId   string  `json:"summonerId"`
	Puuid        string  `json:"puuid"`
	Tier         *string `json:"tier,omitempty"`
	Rank         *string `json:"rank,omitempty"`
	QueueType    *string `json:"queueType,omitempty"`
	LeaguePoints int     `json:"leaguePoints"`
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
	FreshBlood   bool    `json:"freshBlood"`
	HotStreak    bool    `json:"hotStreak"`
}

// The high elo league entries come in a very similar way.
// Only having some outer keys.
type HighEloLeagueEntry struct {
	Entries []LeagueEntry `json:"entries"`
	Tier    string        `json:"tier"`
}
