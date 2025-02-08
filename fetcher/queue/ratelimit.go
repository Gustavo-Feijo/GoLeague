package queue

import (
	"sync"
	"time"
)

type RiotLimit struct {
	limit         int
	resetInterval time.Duration
	count         int
	lastReset     time.Time
}

type RateLimiter struct {
	windows       []*RiotLimit
	fetchInterval time.Duration

	lastFetch time.Time

	mu sync.Mutex
}

// Create a instance of the rate limiter.
func CreateRateLimiter(fetchInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		// Hardcoded values for now.
		windows: []*RiotLimit{
			{
				limit:         20,
				resetInterval: time.Second,
				lastReset:     time.Now(),
			},
			{
				limit:         100,
				resetInterval: 2 * time.Minute,
				lastReset:     time.Now(),
			},
		},
		fetchInterval: fetchInterval,
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

// Verify if can run the job/background request.
func (r *RateLimiter) CanRunJob() bool {
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
func (r *RateLimiter) CanRunApi() bool {
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
