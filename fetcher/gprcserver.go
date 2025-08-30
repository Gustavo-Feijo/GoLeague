package main

import (
	"context"
	"fmt"
	regionmanager "goleague/fetcher/regionmanager"
	"goleague/fetcher/regions"
	pb "goleague/pkg/grpc"
	"goleague/pkg/logger"
	"strings"
	"time"
)

const (
	MAX_CONCURRENCY = 10
	MAX_GRPC_LOGS   = 1000
)

// Server definition.
type server struct {
	pb.UnimplementedServiceServer
	regionManager *regionmanager.RegionManager
	logger        *logger.NewLogger
}

// FetchMatchHistory forces the player match history processing.
func (s *server) FetchMatchHistory(ctx context.Context, req *pb.SummonerRequest) (*pb.MatchHistoryFetchNotification, error) {
	subRegion := regions.SubRegion(strings.ToUpper(req.Region))
	mainRegion, err := s.regionManager.GetMainRegion(subRegion)
	if err != nil {
		return nil, err
	}

	mainRegionService, err := s.regionManager.GetMainService(mainRegion)
	if err != nil {
		return nil, err
	}

	player, err := mainRegionService.GetPlayerByNameTagRegion(req.GameName, req.TagLine, string(subRegion))
	if err != nil {
		return nil, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		mainRegionService.ProcessPlayerHistory(ctx, player, subRegion, s.logger, MAX_CONCURRENCY, true)
		if s.logger.GetNumberOfWrites() > MAX_GRPC_LOGS {
			objectKey := fmt.Sprintf("grpc/%s.log", time.Now().Format("2006-01-02-15-04-05"))

			s.logger.UploadToS3Bucket(objectKey)
		}
	}()

	response := &pb.MatchHistoryFetchNotification{
		Message:     "Fetching player has started...",
		WillProcess: true,
	}

	return response, nil
}

// GetRiotAccount utilizes the regions services to communicate with the Riot api
// Gets the data from the player account and summoner data to return it.
func (s *server) FetchSummonerData(ctx context.Context, req *pb.SummonerRequest) (*pb.Summoner, error) {
	subRegion := regions.SubRegion(strings.ToUpper(req.Region))
	mainRegion, err := s.regionManager.GetMainRegion(subRegion)
	if err != nil {
		return nil, err
	}

	mainRegionService, err := s.regionManager.GetMainService(mainRegion)
	if err != nil {
		return nil, err
	}

	account, err := mainRegionService.GetAccount(req.GameName, req.TagLine)
	if err != nil {
		return nil, err
	}

	subRegionService, err := s.regionManager.GetSubService(subRegion)
	if err != nil {
		return nil, err
	}

	summoner, err := subRegionService.ProcessSummonerData(account, true)
	if err != nil {
		return nil, fmt.Errorf("couldn't process summoner: %w", err)
	}

	_ = subRegionService.ProcessPlayerLeagueEntries(summoner.Puuid, true)

	// Convert fetcher response to gRPC response
	response := &pb.Summoner{
		Puuid:         summoner.Puuid,
		GameName:      summoner.RiotIdGameName,
		TagLine:       summoner.RiotIdTagline,
		Region:        string(summoner.Region),
		SummonerLevel: int32(summoner.SummonerLevel),
		ProfileIconId: int32(summoner.ProfileIcon),
	}

	return response, nil
}
