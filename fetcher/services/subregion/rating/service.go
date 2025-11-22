package ratingservice

import (
	"fmt"
	leaguefetcher "goleague/fetcher/data/league"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	"goleague/pkg/regions"
	tiervalues "goleague/pkg/riotvalues/tier"
)

// RatingService handles all rating-related operations.
type RatingService struct {
	repository repositories.RatingRepository
	subRegion  regions.SubRegion
}

// NewRatingService creates a new rating service.
func NewRatingService(repository repositories.RatingRepository, subRegion regions.SubRegion) *RatingService {
	return &RatingService{
		repository: repository,
		subRegion:  subRegion,
	}
}

// createRatingFromEntry creates a rating entry from a league entry.
func (s *RatingService) createRatingFromEntry(
	player *models.PlayerInfo,
	entry leaguefetcher.LeagueEntry,
	queue string,
) models.RatingEntry {
	newRating := models.RatingEntry{
		PlayerId:     player.ID,
		Region:       s.subRegion,
		Queue:        queue,
		LeaguePoints: entry.LeaguePoints,
		Wins:         entry.Wins,
		Losses:       entry.Losses,
	}

	// Handle Tier and Rank if they are not nil.
	if entry.Tier != nil {
		newRating.Tier = *entry.Tier
	}

	if entry.Rank != nil {
		newRating.Rank = *entry.Rank
	} else {
		// If it's high elo, it will be nil, just set the ranking as I.
		newRating.Rank = "I"
	}

	// Calculate the numeric score before saving.
	newRating.NumericScore = tiervalues.CalculateRank(newRating.Tier, newRating.Rank, newRating.LeaguePoints)

	return newRating
}

// GetLastRatingsByPlayerIdsAndQueue fetches the last ratings for a list of players.
func (s *RatingService) GetLastRatingsByPlayerIdsAndQueue(playerIDs []uint, queue string) (map[uint]*models.RatingEntry, error) {
	return s.repository.GetLastRatingEntryByPlayerIdsAndQueue(playerIDs, queue)
}

// ProcessRatings processes ratings for players, creating new ones when needed.
func (s *RatingService) ProcessRatings(
	existingPlayers map[string]*models.PlayerInfo,
	entryByPuuid map[string]leaguefetcher.LeagueEntry,
	lastRatings map[uint]*models.RatingEntry,
	queue string,
) (
	[]models.RatingEntry,
	error,
) {
	var ratingsToCreate []models.RatingEntry

	for _, player := range existingPlayers {
		// Get the corresponding entry for the player, as well as the last rating.
		entry := entryByPuuid[player.Puuid]
		lastRating, exists := lastRatings[player.ID]

		// If the last rating doesn't exist or it changed, then create a new rating.
		if !exists || s.repository.RatingNeedsUpdate(lastRating, entry) {
			newRating := s.createRatingFromEntry(player, entry, queue)
			ratingsToCreate = append(ratingsToCreate, newRating)
		}
	}

	// Create the ratings.
	if len(ratingsToCreate) > 0 {
		if err := s.repository.CreateBatchRating(ratingsToCreate); err != nil {
			return nil, fmt.Errorf("error creating rating entries: %v", err)
		}
	}

	return ratingsToCreate, nil
}
