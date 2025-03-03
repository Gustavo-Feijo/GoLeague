package player_fetcher

import (
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"net/http"
	"strconv"
	"time"
)

// The player fetcher with it's limit and region.
type Player_fetcher struct {
	limiter *requests.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// The player fetcher with it's limit and region.
type Sub_player_fetcher struct {
	limiter *requests.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// Create a player fetcher.
func CreatePlayerFetcher(limiter *requests.RateLimiter, region string) *Player_fetcher {
	return &Player_fetcher{
		limiter,
		region,
	}
}

// Create a player fetcher.
func CreateSubPlayerFetcher(limiter *requests.RateLimiter, region string) *Sub_player_fetcher {
	return &Sub_player_fetcher{
		limiter,
		region,
	}
}

// Get a players match list.
func (p *Player_fetcher) GetMatchList(puuid string, lastFetch time.Time, offset int, onDemand bool) ([]string, error) {
	if onDemand {
		p.limiter.WaitApi()
	} else {
		p.limiter.WaitJob()
	}
	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids", p.region, puuid)
	params := map[string]string{
		"startTime": strconv.FormatInt(lastFetch.Unix(), 10),
		"start":     strconv.Itoa(offset),
		"count":     "100", // 100 is the maximum allowed count.
	}

	resp, err := requests.AuthRequest(url, "GET", params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the matches list.
	var matches []string
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the matches.
	return matches, nil
}

// Get a players summoner data.
func (p *Sub_player_fetcher) GetSummonerData(puuid string, onDemand bool) (*SummonerByPuuid, error) {
	if onDemand {
		p.limiter.WaitApi()
	} else {
		p.limiter.WaitJob()
	}
	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-puuid/%s", p.region, puuid)

	// Make the request with proper auth.
	resp, err := requests.AuthRequest(url, "GET", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the matches list.
	var summonerData SummonerByPuuid
	if err := json.NewDecoder(resp.Body).Decode(&summonerData); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the matches.
	return &summonerData, nil
}
