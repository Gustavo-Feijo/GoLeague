package league_fetcher

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"net/http"
	"strings"
)

// The league fetcher with it's limit and region.
type League_Fetcher struct {
	limiter *requests.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// Create a league fetcher.
func CreateLeagueFetcher(limiter *requests.RateLimiter, region string) *League_Fetcher {
	return &League_Fetcher{
		limiter,
		region,
	}
}

// Get a given high elo league page.
// Used only for  job  requests, since it would not be necessary to get a given page at demand.
func (l *League_Fetcher) GetHighEloLeagueEntries(division string, queue string) (*HighEloLeagueEntry, error) {
	// Wait for job.
	l.limiter.WaitJob()

	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/league/v4/%s/by-queue/%s",
		l.region, strings.ToLower(division), queue)

	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the league entries.
	var entries HighEloLeagueEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the entries.
	return &entries, nil
}

// Get a given player entries for each queue.
func (l *League_Fetcher) GetLeagueByPuuid(puuid string, onDemand bool) ([]LeagueEntry, error) {
	// Verify the type of request.
	if onDemand {
		l.limiter.WaitApi()
	} else {
		l.limiter.WaitJob()
	}
	// Format the URL and create the params.
	// Riot only accept upper case on this entries.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/league/v4/entries/by-puuid/%s",
		l.region, puuid)

	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the league entries.
	var entries []LeagueEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the entries.
	return entries, nil
}

// Get a given league page.
// Used only for  job  requests, since it would not be necessary to get a given page at demand.
func (l *League_Fetcher) GetLeagueEntries(tier string, division string, queue string, page int) ([]LeagueEntry, error) {
	// Wait for job.
	l.limiter.WaitJob()

	// Format the URL and create the params.
	// Riot only accept upper case on this entries.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/league/v4/entries/%s/%s/%s?page=%d",
		l.region, queue, strings.ToUpper(tier), strings.ToUpper(division), page)

	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the league entries.
	var entries []LeagueEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the entries.
	return entries, nil
}
