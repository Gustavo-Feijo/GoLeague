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
func CreateMainFetcher(region string) *MainFetcher {
	// Create the limiter for this region.
	limiter := requests.CreateRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &MainFetcher{
		Player: player_fetcher.CreatePlayerFetcher(limiter, region),
		Match:  match_fetcher.CreateMatchFetcher(limiter, region),
		League: league_fetcher.CreateLeagueFetcher(limiter, region),
	}
}

func CreateSubFetcher(region string) *SubFetcher {
	// Create the limiter for this region.
	limiter := requests.CreateRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &SubFetcher{
		Player: player_fetcher.CreateSubPlayerFetcher(limiter, region),
		Match:  match_fetcher.CreateSubMatchFetcher(limiter, region),
		League: league_fetcher.CreateSubLeagueFetcher(limiter, region),
	}
}
