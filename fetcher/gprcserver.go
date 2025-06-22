package main

import (
	regionmanager "goleague/fetcher/regionmanager"
	pb "goleague/pkg/grpc"
)

// Server definition.
type server struct {
	pb.UnimplementedAssetsServiceServer
	regionManager *regionmanager.RegionManager
}
