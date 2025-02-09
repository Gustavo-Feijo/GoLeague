package data

import (
	match_fetcher "goleague/fetcher/data/match"
	player_fetcher "goleague/fetcher/data/player"
	"goleague/fetcher/requests"
)

// Define a main fetcher.
type MainFetcher struct {
	Player *player_fetcher.Player_fetcher
	Match  *match_fetcher.Match_fetcher
}

// Function to instanciate the main fetcher.
func CreateMainFetcher(region string) *MainFetcher {
	// Create the limiter for this region.
	limiter := requests.CreateRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &MainFetcher{
		Player: player_fetcher.CreatePlayerFetcher(limiter, region),
		Match:  match_fetcher.CreateMatchFetcher(limiter, region),
	}
}
