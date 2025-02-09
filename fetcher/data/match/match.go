package match_fetcher

import (
	"goleague/fetcher/requests"
)

type Match_fetcher struct {
	limiter *requests.RateLimiter
	region  string
}

func CreateMatchFetcher(limiter *requests.RateLimiter, region string) *Match_fetcher {
	return &Match_fetcher{
		limiter,
		region,
	}
}
