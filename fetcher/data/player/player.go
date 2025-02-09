package player_fetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"goleague/fetcher/requests"
	"strconv"
	"time"
)

type Player_fetcher struct {
	limiter *requests.RateLimiter
	region  string
}

func CreatePlayerFetcher(limiter *requests.RateLimiter, region string) *Player_fetcher {
	return &Player_fetcher{
		limiter,
		region,
	}
}

func (p *Player_fetcher) GetMatchList(puuid string, lastFetch time.Time, offset int) ([]string, error) {
	// Format the URL and create the params.
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids", p.region, puuid)
	params := map[string]string{
		"startTime": strconv.FormatInt(lastFetch.Unix(), 10),
		"start":     strconv.Itoa(offset),
		"count":     "100", // 100 is the maximum allowed count.
	}

	resp, err := requests.AuthRequest(url, "GET", params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Parse the matches list.
	var matches []string
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, err
	}

	// Verify if any match was found.
	if len(matches) == 0 {
		return nil, errors.New("couldn't find any match")
	}

	// Return the matches.
	return matches, nil
}
