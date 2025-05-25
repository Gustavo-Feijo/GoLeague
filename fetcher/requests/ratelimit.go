package requests

import (
	"goleague/pkg/config"

	"github.com/Gustavo-Feijo/gomultirate"
)

// NewRateLimiter creates a instance of the rate limiter.
func NewRateLimiter() *gomultirate.RateLimiter {
	limits := map[string]*gomultirate.Limit{
		"api": gomultirate.NewLimit(
			config.Limits.Lower.ResetInterval,
			config.Limits.Lower.Count,
		),
		"job": gomultirate.NewLimit(
			config.Limits.Higher.ResetInterval,
			config.Limits.Higher.Count,
		),
	}

	limiter, _ := gomultirate.NewRateLimiter(limits)
	return limiter
}
