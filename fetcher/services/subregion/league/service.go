package leagueservice

import (
	"fmt"
	"goleague/fetcher/data"
	leaguefetcher "goleague/fetcher/data/league"
)

// LeagueService handles all interactions with the League API.
type LeagueService struct {
	fetcher    data.SubFetcher
	maxRetries int
}

// LeagueServiceConfig is the configuration for the league service.
type LeagueServiceConfig struct {
	MaxRetries int
}

// DefaultConfig provides default configuration.
func DefaultConfig() LeagueServiceConfig {
	return LeagueServiceConfig{
		MaxRetries: 3,
	}
}

// NewLeagueService creates a new league service.
func NewLeagueService(fetcher data.SubFetcher, config LeagueServiceConfig) *LeagueService {
	return &LeagueService{
		fetcher:    fetcher,
		maxRetries: config.MaxRetries,
	}
}

// ExtractPuuidsFromEntries retrieves PUUIDs from league entries.
func (s *LeagueService) ExtractPuuidsFromEntries(entries []leaguefetcher.LeagueEntry) ([]string, map[string]leaguefetcher.LeagueEntry) {
	puuids := make([]string, len(entries))
	entryByPuuid := make(map[string]leaguefetcher.LeagueEntry)

	for i, entry := range entries {
		puuids[i] = entry.Puuid
		entryByPuuid[entry.Puuid] = entry
	}

	return puuids, entryByPuuid
}

// GetHighEloLeagueEntries fetches high elo league entries with retry.
func (s *LeagueService) GetHighEloLeagueEntries(highElo string, queue string) ([]leaguefetcher.LeagueEntry, error) {
	var highRating *leaguefetcher.HighEloLeagueEntry
	var err error

	// Try to get the entries with retry.
	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		highRating, err = s.fetcher.League.GetHighEloLeagueEntries(highElo, queue)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch high elo entries after %d attempts: %v", s.maxRetries, err)
	}

	// Process each rating entry.
	entries := make([]leaguefetcher.LeagueEntry, len(highRating.Entries))
	for i, entry := range highRating.Entries {
		// For high elo we don't have the tier inside the entries array, so we set manually.
		entry.Tier = &highRating.Tier
		entry.QueueType = &queue
		entries[i] = entry
	}

	return entries, nil
}

// GetLeagueEntries fetches the league entries.
func (s *LeagueService) GetLeagueEntries(tier string, rank string, queue string, page int) ([]leaguefetcher.LeagueEntry, error) {
	var entries []leaguefetcher.LeagueEntry
	var err error

	// Try to get the entries with retry.
	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		entries, err = s.fetcher.League.GetLeagueEntries(tier, rank, queue, page)
		if err == nil {
			break
		}
	}

	// If all retries failed.
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league entries after %d attempts: %v", s.maxRetries, err)
	}

	// Set the queue type for each entry.
	for i := range entries {
		entries[i].QueueType = &queue
	}

	return entries, nil
}
