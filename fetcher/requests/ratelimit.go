package requests

import (
	"goleague/pkg/config"

	"github.com/Gustavo-Feijo/gomultirate"
)

// NewRateLimiter creates a instance of the rate limiter.
func NewRateLimiter(config config.RiotLimiterConfig) *gomultirate.RateLimiter {
	limits := map[string]*gomultirate.Limit{
		"api": gomultirate.NewLimit(
			config.Lower.ResetInterval,
			config.Lower.Count,
		),
		"job": gomultirate.NewLimit(
			config.Higher.ResetInterval,
			config.Higher.Count,
		),
	}

	limiter, _ := gomultirate.NewRateLimiter(limits)
	return limiter
}
