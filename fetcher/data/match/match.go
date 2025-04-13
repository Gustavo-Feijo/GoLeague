package matchfetcher

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"net/http"
	"time"
)

// The match fetcher with it's limiter and region.
type Match_fetcher struct {
	limiter *requests.RateLimiter
	region  string
}

// The match fetcher with it's limiter and region.
type Sub_match_fetcher struct {
	limiter *requests.RateLimiter
	region  string
}

// Create a instance of the match fetcher.
func CreateMatchFetcher(limiter *requests.RateLimiter, region string) *Match_fetcher {
	return &Match_fetcher{
		limiter,
		region,
	}
}

// Create a instance of the match fetcher.
func CreateSubMatchFetcher(limiter *requests.RateLimiter, region string) *Sub_match_fetcher {
	return &Sub_match_fetcher{
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

	// Convert milliseconds to time.Time
	*rt = RiotTime(time.UnixMilli(timestamp))
	return nil
}

// Get the true time.
func (rt RiotTime) Time() time.Time {
	return time.Time(rt)
}

// Return type from the match_v5 endpoint.
type MatchData struct {
	Info MatchInfo `json:"info"`
}

// Get a given match data.
func (m *Match_fetcher) GetMatchData(matchId string, onDemand bool) (*MatchData, error) {
	// Verify if it's onDemand.
	if onDemand {
		m.limiter.WaitApi()
	} else {
		m.limiter.WaitJob()
	}

	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", m.region, matchId)

	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}
	// Parse the matches data.
	var matchData MatchData
	if err := json.NewDecoder(resp.Body).Decode(&matchData); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the matches.
	return &matchData, nil
}

// Get a given match timeline.
func (m *Match_fetcher) GetMatchTimelineData(matchId string, onDemand bool) (*MatchTimeline, error) {
	// Verify if it's onDemand.
	if onDemand {
		m.limiter.WaitApi()
	} else {
		m.limiter.WaitJob()
	}

	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s/timeline", m.region, matchId)

	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}
	// Parse the match timeline.
	var matchTimeline MatchTimeline
	if err := json.NewDecoder(resp.Body).Decode(&matchTimeline); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the timeline.
	return &matchTimeline, nil
}
