package matchfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"time"

	"github.com/Gustavo-Feijo/gomultirate"
)

// MatchFetcher with it's limiter and region.
type MatchFetcher struct {
	limiter *gomultirate.RateLimiter
	region  string
}

// SubMatchFetcher with it's limiter and region.
type SubMatchFetcher struct {
	limiter *gomultirate.RateLimiter
	region  string
}

// NewMatchFetcher creates a instance of the match fetcher.
func NewMatchFetcher(limiter *gomultirate.RateLimiter, region string) *MatchFetcher {
	return &MatchFetcher{
		limiter,
		region,
	}
}

// NewSubMatchFetcher creates a instance of the match fetcher.
func NewSubMatchFetcher(limiter *gomultirate.RateLimiter, region string) *SubMatchFetcher {
	return &SubMatchFetcher{
		limiter,
		region,
	}
}

// Handle the conversion of the int timestamps from riot.
type RiotTime time.Time

// Add the riot time UnmarshalJSON.
func (rt *RiotTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}

	// Convert milliseconds to time.Time.
	*rt = RiotTime(time.UnixMilli(timestamp))
	return nil
}

// Get the true time.
func (rt RiotTime) Time() time.Time {
	return time.Time(rt)
}

// MatchData is the return type from the match_v5 endpoint.
type MatchData struct {
	Info MatchInfo `json:"info"`
}

// GetMatchData returns a given match data.
func (m *MatchFetcher) GetMatchData(matchId string, onDemand bool) (*MatchData, error) {
	ctx := context.Background()
	// Verify if it's onDemand.
	if onDemand {
		m.limiter.Wait(ctx)
	} else {
		m.limiter.WaitEvenly(ctx, "job")
	}

	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", m.region, matchId)

	return requests.HandleAuthRequest[*MatchData](url, "GET", map[string]string{})
}

// GetMatchTimelineData returns a given match timeline.
func (m *MatchFetcher) GetMatchTimelineData(matchId string, onDemand bool) (*MatchTimeline, error) {
	ctx := context.Background()
	if onDemand {
		m.limiter.Wait(ctx)
	} else {
		m.limiter.WaitEvenly(ctx, "job")
	}

	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s/timeline", m.region, matchId)

	return requests.HandleAuthRequest[*MatchTimeline](url, "GET", map[string]string{})
}
