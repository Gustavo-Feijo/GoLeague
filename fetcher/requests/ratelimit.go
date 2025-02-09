package requests

import (
	"goleague/pkg/config"
	"sync"
	"time"
)

// Single riot rate limiting.
type RiotLimit struct {
	limit         int
	resetInterval time.Duration
	count         int
	lastReset     time.Time
}

// Full riot rate limit, containing all the constraints.
type RateLimiter struct {
	windows []*RiotLimit

	// Fetch interval for the background job.
	// Will be the slowest interval that let all requests be consumed before reseting.
	fetchInterval time.Duration

	// Last fetch and the mutex.
	lastFetch time.Time
	mu        sync.Mutex
}

// Create a instance of the rate limiter.
func CreateRateLimiter() *RateLimiter {
	return &RateLimiter{
		// Hardcoded values for now.
		windows: []*RiotLimit{
			{
				limit:         config.Limits.Lower.Count,
				resetInterval: config.Limits.Lower.ResetInterval,
				lastReset:     time.Now(),
			},
			{
				limit:         config.Limits.Higher.Count,
				resetInterval: config.Limits.Higher.ResetInterval,
				lastReset:     time.Now(),
			},
		},
		fetchInterval: config.Limits.SlowInterval,
		lastFetch:     time.Now(),
	}
}

// Reset the count.
func (r *RateLimiter) resetCounts() {
	// Get the current time.
	now := time.Now()
	// Loop through each window and verify if can reset.
	for _, window := range r.windows {
		if now.Sub(window.lastReset) >= window.resetInterval {
			window.count = 0
			window.lastReset = now
		}
	}
}

// Check if the window is on it's limits.
func (r *RateLimiter) checkLimits() bool {
	// Loop through each window.
	for _, window := range r.windows {
		if window.count >= window.limit {
			return false
		}
	}
	return true
}

// Loop through each window and increment the counter.
func (r *RateLimiter) incrementCounts() {
	// Loop through each window and increment each count.
	for _, window := range r.windows {
		window.count++
	}
}

// Wait until the next refresh.
func (r *RateLimiter) WaitApi() {
	// Verify if can run the API.
	if r.canRunApi() {
		return
	}

	// Verify if the windows limit wasn't reached.
	r.waitWindowsReset()

	// Verify again the API.
	r.WaitApi()
}

// Wait until next job refresh.
func (r *RateLimiter) WaitJob() {
	// Verify if can run the job.
	if r.canRunJob() {
		return
	}

	// Verify if the elapsed time until the next job fetch was reached.
	if time.Since(r.lastFetch) > r.fetchInterval {
		waitTill := r.fetchInterval - time.Since(r.lastFetch)
		time.Sleep(waitTill)
	}

	// Verify if the general limit wasn't already reached.
	r.waitWindowsReset()
	// Verify again for the job.
	r.WaitJob()
}

// Wait until all the rate limit windows are met.
func (r *RateLimiter) waitWindowsReset() {
	// If can't run, see how many time must wait.
	var waitTime time.Duration
	waitTime = 0
	for _, window := range r.windows {
		// If it's not this window that is limited, just continue.
		if window.count < window.limit {
			continue
		}

		// See how many time has elapsed since the last reset.
		elapsed := time.Since(window.lastReset)
		// See how many time till the next reset.
		waitTill := window.resetInterval - elapsed
		if waitTill > waitTime {
			waitTime = waitTill
		}
	}
	// Wait till next reset.
	time.Sleep(waitTime)
}

// Verify if can run the job/background request.
func (r *RateLimiter) canRunJob() bool {
	// Locks the limiter.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resetCounts()

	// Verify if it's not to early.
	if time.Since(r.lastFetch) < r.fetchInterval {
		return false
	}

	// Verify the limit.
	if !r.checkLimits() {
		return false
	}

	// Increment the  count.
	r.incrementCounts()
	r.lastFetch = time.Now()
	return true
}

// Verify if can run the API.
func (r *RateLimiter) canRunApi() bool {
	// Locks the limiter.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resetCounts()

	// Check the limits.
	if !r.checkLimits() {
		return false
	}

	r.incrementCounts()
	return true
}
