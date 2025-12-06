package data

import (
	leaguefetcher "goleague/fetcher/data/league"
	matchfetcher "goleague/fetcher/data/match"
	playerfetcher "goleague/fetcher/data/player"
	"goleague/fetcher/requests"
	"goleague/pkg/config"
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
func NewMainFetcher(config *config.Config, region string) *MainFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter(config.Limits)

	// Return the fetcher with it's player instance for queries.
	return &MainFetcher{
		Player: playerfetcher.NewPlayerFetcher(config.ApiKey, limiter, region),
		Match:  matchfetcher.NewMatchFetcher(config.ApiKey, limiter, region),
		League: leaguefetcher.NewLeagueFetcher(config.ApiKey, limiter, region),
	}
}

// NewSubFetcher instanciate the sub fetcher.
func NewSubFetcher(config *config.Config, region string) *SubFetcher {
	// Create the limiter for this region.
	limiter := requests.NewRateLimiter(config.Limits)

	// Return the fetcher with it's player instance for queries.
	return &SubFetcher{
		Player: playerfetcher.NewSubPlayerFetcher(config.ApiKey, limiter, region),
		Match:  matchfetcher.NewSubMatchFetcher(config.ApiKey, limiter, region),
		League: leaguefetcher.NewSubLeagueFetcher(config.ApiKey, limiter, region),
	}
}
