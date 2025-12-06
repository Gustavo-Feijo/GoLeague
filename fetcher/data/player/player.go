package playerfetcher

import (
	"context"
	"fmt"
	"goleague/fetcher/requests"
	"strconv"
	"time"

	"github.com/Gustavo-Feijo/gomultirate"
)

// PlayerFetcher with it's limit and region.
type PlayerFetcher struct {
	apiKey  string
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// SubPlayerFetcher with it's limit and region.
type SubPlayerFetcher struct {
	apiKey  string
	limiter *gomultirate.RateLimiter // Pointer to the fetcher, since it's shared.
	region  string
}

// NewPlayerFetcher creates a player fetcher.
func NewPlayerFetcher(apiKey string, limiter *gomultirate.RateLimiter, region string) *PlayerFetcher {
	return &PlayerFetcher{
		apiKey,
		limiter,
		region,
	}
}

// NewSubPlayerFetcher creates a player fetcher.
func NewSubPlayerFetcher(apiKey string, limiter *gomultirate.RateLimiter, region string) *SubPlayerFetcher {
	return &SubPlayerFetcher{
		apiKey,
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

	return requests.HandleAuthRequest[[]string](p.apiKey, url, "GET", params)
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

	account, err := requests.HandleAuthRequest[Account](p.apiKey, url, "GET", params)
	return &account, err
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

	params := map[string]string{}

	summoner, err := requests.HandleAuthRequest[SummonerByPuuid](p.apiKey, url, "GET", params)
	return &summoner, err
}
