package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"goleague/api/filters"
	"time"

	pb "goleague/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// PlayerGRPCClient is a interface for any player related gRPC client fetching.
type PlayerGRPCClient interface {
	ForceFetchPlayer(filters *filters.PlayerForceFetchFilter, operation string) (*pb.Summoner, error)
	ForceFetchPlayerMatchHistory(filters *filters.PlayerForceFetchMatchListFilter, operation string) (*pb.MatchHistoryFetchNotification, error)
}

type playerGRPCClient struct {
	*grpc.ClientConn
}

// NewPlayerGRPCClient creates a new player gRPC client.
func NewPlayerGRPCClient(grpcConn *grpc.ClientConn) PlayerGRPCClient {
	return &playerGRPCClient{ClientConn: grpcConn}
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (pgc *playerGRPCClient) ForceFetchPlayer(filters *filters.PlayerForceFetchFilter, operation string) (*pb.Summoner, error) {
	client := pb.NewServiceClient(pgc.ClientConn)

	grpcCall := func(ctx context.Context, req *pb.SummonerRequest) (any, error) {
		return client.FetchSummonerData(ctx, req)
	}

	resp, err := pgc.executeSummonerGRPCCall(filters.GameName, filters.GameTag, filters.Region, operation, grpcCall)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.Summoner), nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (pgc *playerGRPCClient) ForceFetchPlayerMatchHistory(filters *filters.PlayerForceFetchMatchListFilter, operation string) (*pb.MatchHistoryFetchNotification, error) {
	client := pb.NewServiceClient(pgc.ClientConn)

	grpcCall := func(ctx context.Context, req *pb.SummonerRequest) (any, error) {
		return client.FetchMatchHistory(ctx, req)
	}

	resp, err := pgc.executeSummonerGRPCCall(filters.GameName, filters.GameTag, filters.Region, operation, grpcCall)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.MatchHistoryFetchNotification), nil
}

// executeSummonerGRPCCall is a helper to execute any gRPC call for summoner requests.
func (pgc *playerGRPCClient) executeSummonerGRPCCall(
	gameName string,
	gameTag string,
	region string,
	operation string,
	grpcCall func(context.Context, *pb.SummonerRequest) (any, error),
) (any, error) {
	request := &pb.SummonerRequest{
		GameName: gameName,
		TagLine:  gameTag,
		Region:   region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := grpcCall(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't execute %s: %w", operation, errors.New(st.Message()))
		}
		return nil, fmt.Errorf("couldn't execute %s: %w", operation, err)
	}

	return resp, nil
}
