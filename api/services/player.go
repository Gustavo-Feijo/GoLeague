package services

import (
	"errors"
	"fmt"
	"goleague/api/dto"
	"goleague/api/repositories"

	"google.golang.org/grpc"
)

// PlayerService service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type PlayerService struct {
	PlayerRepository repositories.PlayerRepository
	grpcClient       *grpc.ClientConn
}

// NewPlayerService creates a service for handling player services.
func NewPlayerService(grpcClient *grpc.ClientConn) (*PlayerService, error) {
	// Create the repository.
	repo, err := repositories.NewPlayerRepository()
	if err != nil {
		return nil, errors.New("failed to start the player repository")
	}

	return &PlayerService{
		grpcClient:       grpcClient,
		PlayerRepository: repo,
	}, nil
}

// GetPlayerSearch returns the result of a given search.
func (ps *PlayerService) GetPlayerSearch(filters map[string]any) ([]*dto.PlayerSearch, error) {
	return ps.PlayerRepository.SearchPlayer(filters)
}

// GetPlayerMatchHistory returns a player match list based on filters.
func (ps *PlayerService) GetPlayerMatchHistory(filters map[string]any) (error, error) {
	// Convert to string.
	// Received through path params.
	name := filters["gameName"].(string)
	tag := filters["gameTag"].(string)
	region := filters["region"].(string)

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf("couldn't find the playerId: %w", err)
	}

	filters["playerId"] = playerId.ID
	return ps.PlayerRepository.GetPlayerMatchHistory(filters), nil
}
