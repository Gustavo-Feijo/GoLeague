package leaguefetcher

// Define the type return by the simple league entries.
type LeagueEntry struct {
	FreshBlood   bool    `json:"freshBlood"`
	HotStreak    bool    `json:"hotStreak"`
	LeaguePoints int     `json:"leaguePoints"`
	Losses       int     `json:"losses"`
	Puuid        string  `json:"puuid"`
	QueueType    *string `json:"queueType,omitempty"`
	Rank         *string `json:"rank,omitempty"`
	SummonerId   string  `json:"summonerId"`
	Tier         *string `json:"tier,omitempty"`
	Wins         int     `json:"wins"`
}

// The high elo league entries come in a very similar way.
// Only having some outer keys.
type HighEloLeagueEntry struct {
	Entries []LeagueEntry `json:"entries"`
	Tier    string        `json:"tier"`
}
