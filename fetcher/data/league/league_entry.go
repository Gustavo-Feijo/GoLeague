package leaguefetcher

// LeagueEntry defines the type returned by the simple league entries.
type LeagueEntry struct {
	FreshBlood   bool    `json:"freshBlood"`
	HotStreak    bool    `json:"hotStreak"`
	LeaguePoints int     `json:"leaguePoints"`
	Losses       int     `json:"losses"`
	Puuid        string  `json:"puuid"`
	QueueType    *string `json:"queueType,omitempty"`
	Rank         *string `json:"rank,omitempty"`
	Tier         *string `json:"tier,omitempty"`
	Wins         int     `json:"wins"`
}

// HighEloLeagueEntry come in a very similar way.
// Only having some outer keys.
type HighEloLeagueEntry struct {
	Entries []LeagueEntry `json:"entries"`
	Tier    string        `json:"tier"`
}
