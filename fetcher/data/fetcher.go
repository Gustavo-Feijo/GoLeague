package data

import (
	leaguefetcher "goleague/fetcher/data/league"
	matchfetcher "goleague/fetcher/data/match"
	playerfetcher "goleague/fetcher/data/player"
	"goleague/fetcher/requests"
)

// MainFetcher with it's dependencies.
type MainFetcher struct {
	Player *playerfetcher.PlayerFetcher
	Match  *matchfetcher.MatchFetcher
	League *leaguefetcher.LeagueFetcher
}

// SubFetcher with it's dependencies.
type SubFetcher struct {
	Player *playerfetcher.SubPlayerFetcher
	Match  *matchfetcher.SubMatchFetcher
	League *leaguefetcher.SubLeagueFetcher
}

// NewMainFetcher instanciate the main fetcher.
func NewMainFetcher(region string) *MainFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &MainFetcher{
		Player: playerfetcher.NewPlayerFetcher(limiter, region),
		Match:  matchfetcher.NewMatchFetcher(limiter, region),
		League: leaguefetcher.NewLeagueFetcher(limiter, region),
	}
}

// NewSubFetcher instanciate the sub fetcher.
func NewSubFetcher(region string) *SubFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter()

	// Return the fetcher with it's player instance for queries.
	return &SubFetcher{
		Player: playerfetcher.NewSubPlayerFetcher(limiter, region),
		Match:  matchfetcher.NewSubMatchFetcher(limiter, region),
		League: leaguefetcher.NewSubLeagueFetcher(limiter, region),
	}
}
