package playerservice

import (
	"fmt"
	"goleague/fetcher/data"
	leaguefetcher "goleague/fetcher/data/league"
	playerfetcher "goleague/fetcher/data/player"

	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	"goleague/pkg/regions"
)

// PlayerService handles all player-related operations.
type PlayerService struct {
	fetcher    data.SubFetcher
	repository repositories.PlayerRepository
	subRegion  regions.SubRegion
}

// NewPlayerService creates a new player service.
func NewPlayerService(fetcher data.SubFetcher, repository repositories.PlayerRepository, subRegion regions.SubRegion) *PlayerService {
	return &PlayerService{
		fetcher:    fetcher,
		repository: repository,
		subRegion:  subRegion,
	}
}

// GetPlayerIDsFromMap extracts player IDs from a map of players.
func (s *PlayerService) GetPlayerIDsFromMap(players map[string]*models.PlayerInfo) []uint {
	playerIDs := make([]uint, 0, len(players))
	for _, player := range players {
		playerIDs = append(playerIDs, player.ID)
	}
	return playerIDs
}

// GetPlayersByPuuids fetches existing players by their PUUIDs.
func (s *PlayerService) GetPlayersByPuuids(puuids []string) (map[string]*models.PlayerInfo, error) {
	return s.repository.GetPlayersByPuuids(puuids)
}

// Wrapper for getting the summoner data from the fetcher.
func (s *PlayerService) GetSummonerData(puuid string, onDemand bool) (*playerfetcher.SummonerByPuuid, error) {
	return s.fetcher.Player.GetSummonerDataByPuuid(puuid, onDemand)
}

// ProcessPlayersFromEntries processes players from league entries, creating any that don't exist.
func (s *PlayerService) ProcessPlayersFromEntries(
	entries []leaguefetcher.LeagueEntry,
	existingPlayers map[string]*models.PlayerInfo,
) ([]*models.PlayerInfo, error) {
	var playersToCreate []*models.PlayerInfo

	// Loop through each entry and verify if the player exists.
	for _, entry := range entries {
		_, exists := existingPlayers[entry.Puuid]
		// The player doesn't exist.
		if !exists {
			playersToCreate = append(playersToCreate, &models.PlayerInfo{
				Puuid:  entry.Puuid,
				Region: s.subRegion,
			})
		}
	}

	// Creates the list of players.
	if len(playersToCreate) > 0 {
		if err := s.repository.CreatePlayersBatch(playersToCreate); err != nil {
			return nil, fmt.Errorf("error inserting %v new players: %v", len(playersToCreate), err)
		}

		// Add newly created players to the existing players map.
		for _, player := range playersToCreate {
			existingPlayers[player.Puuid] = player
		}
	}

	return playersToCreate, nil
}

// ProcessSummonerData gets a summoner info from the Riot API and upserts the entry in the database.
func (s *PlayerService) ProcessSummonerData(playeraccount *playerfetcher.Account, onDemand bool) (*models.PlayerInfo, error) {
	summonerData, err := s.GetSummonerData(playeraccount.Puuid, onDemand)
	if err != nil {
		return nil, fmt.Errorf("couldn't get summoner data: %w", err)
	}

	fullSummoner := &models.PlayerInfo{
		ProfileIcon:    summonerData.ProfileIconId,
		Puuid:          playeraccount.Puuid,
		RiotIdGameName: playeraccount.GameName,
		RiotIdTagline:  playeraccount.TagLine,
		SummonerLevel:  summonerData.SummonerLevel,
		Region:         s.subRegion,
	}

	// Make it as a array to reuse the upsert in batches, but using a batch of only one player.
	fullSummonerArray := []*models.PlayerInfo{
		fullSummoner,
	}

	err = s.repository.UpsertPlayerBatch(fullSummonerArray)
	if err != nil {
		return nil, fmt.Errorf("couldn't save player on database: %w", err)
	}

	return fullSummoner, nil
}
