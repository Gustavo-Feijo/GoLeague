package playerfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"goleague/fetcher/requests"
	"net/http"
	"strconv"
	"time"

	"github.com/Gustavo-Feijo/gomultirate"
)

// PlayerFetcher with it's limit and region.
type PlayerFetcher struct {
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// SubPlayerFetcher with it's limit and region.
type SubPlayerFetcher struct {
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// NewPlayerFetcher creates a player fetcher.
func NewPlayerFetcher(limiter *gomultirate.RateLimiter, region string) *PlayerFetcher {
	return &PlayerFetcher{
		limiter,
		region,
	}
}

// NewSubPlayerFetcher creates a player fetcher.
func NewSubPlayerFetcher(limiter *gomultirate.RateLimiter, region string) *SubPlayerFetcher {
	return &SubPlayerFetcher{
		limiter,
		region,
	}
}

// GetMatchList returns a players match list.
func (p *PlayerFetcher) GetMatchList(puuid string, lastFetch time.Time, offset int, onDemand bool) ([]string, error) {
	ctx := context.Background()
	if onDemand {
		p.limiter.Wait(ctx)
	} else {
		p.limiter.WaitEvenly(ctx, "job")
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

// GetPlayerAccount returns a given player account info.
func (p *PlayerFetcher) GetPlayerAccount(gameName string, tagLine string, onDemand bool) (*Account, error) {
	ctx := context.Background()
	if onDemand {
		p.limiter.Wait(ctx)
	} else {
		p.limiter.WaitEvenly(ctx, "job")
	}
	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s", p.region, gameName, tagLine)

	params := map[string]string{}

	resp, err := requests.AuthRequest(url, "GET", params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Parse the account data.
	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Return the account data.
	return &account, nil
}

// GetSummonerData returns a players summoner data.
func (p *SubPlayerFetcher) GetSummonerDataByPuuid(puuid string, onDemand bool) (*SummonerByPuuid, error) {
	ctx := context.Background()
	if onDemand {
		p.limiter.Wait(ctx)
	} else {
		p.limiter.WaitEvenly(ctx, "job")
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
