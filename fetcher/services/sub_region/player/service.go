package playerservice

import (
	"fmt"
	league_fetcher "goleague/fetcher/data/league"
	"goleague/fetcher/regions"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
)

// PlayerService handles all player-related operations
type PlayerService struct {
	repository repositories.PlayerRepository
	subRegion  regions.SubRegion
}

// NewPlayerService creates a new player service
func NewPlayerService(repository repositories.PlayerRepository, subRegion regions.SubRegion) *PlayerService {
	return &PlayerService{
		repository: repository,
		subRegion:  subRegion,
	}
}

// GetPlayerIDsFromMap extracts player IDs from a map of players
func (s *PlayerService) GetPlayerIDsFromMap(players map[string]*models.PlayerInfo) []uint {
	playerIDs := make([]uint, 0, len(players))
	for _, player := range players {
		playerIDs = append(playerIDs, player.ID)
	}
	return playerIDs
}

// GetPlayersByPuuids fetches existing players by their PUUIDs
func (s *PlayerService) GetPlayersByPuuids(puuids []string) (map[string]*models.PlayerInfo, error) {
	return s.repository.GetPlayersByPuuids(puuids)
}

// ProcessPlayersFromEntries processes players from league entries, creating any that don't exist
func (s *PlayerService) ProcessPlayersFromEntries(
	entries []league_fetcher.LeagueEntry,
	existingPlayers map[string]*models.PlayerInfo,
) ([]*models.PlayerInfo, error) {
	var playersToCreate []*models.PlayerInfo

	// Loop through each entry and verify if the player exists
	for _, entry := range entries {
		_, exists := existingPlayers[entry.Puuid]
		// The player doesn't exist
		if !exists {
			playersToCreate = append(playersToCreate, &models.PlayerInfo{
				SummonerId: entry.SummonerId,
				Puuid:      entry.Puuid,
				Region:     s.subRegion,
			})
		}
	}

	// Creates the list of players
	if len(playersToCreate) > 0 {
		if err := s.repository.CreatePlayersBatch(playersToCreate); err != nil {
			return nil, fmt.Errorf("error inserting %v new players: %v", len(playersToCreate), err)
		}

		// Add newly created players to the existing players map
		for _, player := range playersToCreate {
			existingPlayers[player.Puuid] = player
		}
	}

	return playersToCreate, nil
}
