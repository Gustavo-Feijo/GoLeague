package data

import (
	league_fetcher "goleague/fetcher/data/league"
	match_fetcher "goleague/fetcher/data/match"
	player_fetcher "goleague/fetcher/data/player"
	"goleague/fetcher/requests"
)

// Define a main fetcher.
type MainFetcher struct {
	Player *player_fetcher.Player_fetcher
	Match  *match_fetcher.Match_fetcher
	League *league_fetcher.League_Fetcher
}

// Define a sub region fetcher.
type SubFetcher struct {
	Player *player_fetcher.Sub_player_fetcher
	Match  *match_fetcher.Sub_match_fetcher
	League *league_fetcher.Sub_league_Fetcher
}

// Function to instanciate the main fetcher.
func NewMainFetcher(region string) *MainFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &MainFetcher{
		Player: player_fetcher.NewPlayerFetcher(limiter, region),
		Match:  match_fetcher.NewMatchFetcher(limiter, region),
		League: league_fetcher.NewLeagueFetcher(limiter, region),
	}
}

func NewSubFetcher(region string) *SubFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &SubFetcher{
		Player: player_fetcher.NewSubPlayerFetcher(limiter, region),
		Match:  match_fetcher.NewSubMatchFetcher(limiter, region),
		League: league_fetcher.NewSubLeagueFetcher(limiter, region),
	}
}
