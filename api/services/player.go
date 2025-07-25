package services

import (
	"context"
	"errors"
	"fmt"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"time"

	pb "goleague/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// PlayerService service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type PlayerService struct {
	db         *gorm.DB
	grpcClient *grpc.ClientConn
	matchCache *cache.MatchCache

	MatchRepository  repositories.MatchRepository
	PlayerRepository repositories.PlayerRepository
}

type PlayerServiceDeps struct {
	DB         *gorm.DB
	GrpcClient *grpc.ClientConn
	MatchCache *cache.MatchCache
}

// NewPlayerService creates a service for handling player services.
func NewPlayerService(deps *PlayerServiceDeps) (*PlayerService, error) {
	// Create the repository.
	repo, err := repositories.NewPlayerRepository(deps.DB)
	if err != nil {
		return nil, errors.New("failed to start the player repository")
	}

	matchRepo, err := repositories.NewMatchRepository(deps.DB)
	if err != nil {
		return nil, errors.New("failed to start the match repository")
	}

	return &PlayerService{
		db:               deps.DB,
		grpcClient:       deps.GrpcClient,
		matchCache:       deps.MatchCache,
		MatchRepository:  matchRepo,
		PlayerRepository: repo,
	}, nil
}

// GetPlayerSearch returns the result of a given search.
func (ps *PlayerService) GetPlayerSearch(filters map[string]any) ([]*dto.PlayerSearch, error) {
	return ps.PlayerRepository.SearchPlayer(filters)
}

// GetPlayerMatchHistory returns a player match list based on filters.
func (ps *PlayerService) GetPlayerMatchHistory(filters map[string]any) (dto.MatchPreviewList, error) {
	// Convert to string.
	// Received through path params.
	name := filters["gameName"].(string)
	tag := filters["gameTag"].(string)
	region := filters["region"].(string)

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf("couldn't find the playerId: %w", err)
	}

	filters["playerId"] = playerId
	matchesIds, err := ps.PlayerRepository.GetPlayerMatchHistoryIds(filters)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match ids: %w", err)
	}

	if len(matchesIds) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Get all the cached matches previews.
	// Match previews shouldn't change, using cache to reduce load into the database.
	cachedMatches, missingMatches, err := ps.matchCache.GetMatchesPreviewByMatchIds(ctx, matchesIds)
	if err == nil {
		// All matches in cache.
		if len(missingMatches) == 0 {
			matchPreviews := make(dto.MatchPreviewList)
			handleCachedMatches(cachedMatches, matchPreviews)
			return matchPreviews, nil
		}

		matchesIds = missingMatches
	}

	// Get the non cached matches from the database.
	matchPreviews, err := ps.MatchRepository.GetMatchPreviews(matchesIds)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match history for the player: %w", err)
	}

	// Some matches came from cache, others from db.
	if len(missingMatches) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for _, match := range matchPreviews {
			ps.matchCache.SetMatchPreview(ctx, *match)
		}
		handleCachedMatches(cachedMatches, matchPreviews)
	}

	return matchPreviews, nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayer(filters filters.PlayerForceFetchParams) (*pb.Summoner, error) {
	client := pb.NewServiceClient(ps.grpcClient)

	request := &pb.SummonerRequest{
		GameName: filters.GameName,
		TagLine:  filters.GameTag,
		Region:   filters.Region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Make the request
	resp, err := client.GetSummonerData(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't force fetch the player: %w", errors.New(st.Message()))
		}
		return nil, fmt.Errorf("couldn't force fetch the player: %w", err)
	}

	return resp, nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayerMatchHistory(filters filters.PlayerForceFetchMatchHistoryParams) (*pb.MatchHistoryFetchNotification, error) {
	client := pb.NewServiceClient(ps.grpcClient)

	request := &pb.SummonerRequest{
		GameName: filters.GameName,
		TagLine:  filters.GameTag,
		Region:   filters.Region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Make the request
	resp, err := client.FetchMatchHistory(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't force fetch the player match history: %w", st.Err())
		}
		return nil, fmt.Errorf("couldn't force fetch the player match history: %w", err)
	}

	return resp, nil
}

func handleCachedMatches(cachedMatches []dto.MatchPreview, matchesDto dto.MatchPreviewList) {
	for _, match := range cachedMatches {
		m := match
		matchesDto[match.Metadata.MatchId] = &m
	}
}
