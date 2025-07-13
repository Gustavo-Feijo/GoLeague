package main

import (
	"context"
	"fmt"
	regionmanager "goleague/fetcher/regionmanager"
	"goleague/fetcher/regions"
	pb "goleague/pkg/grpc"
	"strings"
)

// Server definition.
type server struct {
	pb.UnimplementedServiceServer
	regionManager *regionmanager.RegionManager
}

// GetRiotAccount utilizes the regions services to communicate with the Riot api
// Gets the data from the player account and summoner data to return it.
func (s *server) GetSummonerData(ctx context.Context, req *pb.SummonerRequest) (*pb.Summoner, error) {
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
