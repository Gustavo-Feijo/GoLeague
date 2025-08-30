package leaguefetcher

import (
	"context"
	"fmt"
	"goleague/fetcher/requests"
	"strings"

	"github.com/Gustavo-Feijo/gomultirate"
)

// LeagueFetcher contains the fetcher with it's limit and region.
type LeagueFetcher struct {
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// SubLeagueFetcher is another fetcher instance, used only to diferenciate methods.
type SubLeagueFetcher struct {
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// NewLeagueFetcher creates a new instance of the league fetcher.
func NewLeagueFetcher(limiter *gomultirate.RateLimiter, region string) *LeagueFetcher {
	return &LeagueFetcher{
		limiter,
		region,
	}
}

// Create a league fetcher.
func NewSubLeagueFetcher(limiter *gomultirate.RateLimiter, region string) *SubLeagueFetcher {
	return &SubLeagueFetcher{
		limiter,
		region,
	}
}

// Get a given league page.
// Used only for  job  requests, since it would not be necessary to get a given page at demand.
func (l *SubLeagueFetcher) GetLeagueEntries(tier string, rank string, queue string, page int) ([]LeagueEntry, error) {
	// Wait for job.
	ctx := context.Background()
	l.limiter.WaitEvenly(ctx, "job")

	// Format the URL and create the params.
	// Riot only accept upper case on this entries.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/league-exp/v4/entries/%s/%s/%s?page=%d",
		l.region, queue, strings.ToUpper(tier), strings.ToUpper(rank), page)

	return requests.HandleAuthRequest[[]LeagueEntry](url, "GET", map[string]string{})
}
