package playerservice

import (
	"context"
	"errors"
	"goleague/api/filters"
	"goleague/api/services/testutil"
	pb "goleague/pkg/grpc"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Default type for a gRPC test.
type gRPCTestCase[TFilter any, TResponse any] struct {
	name           string
	filters        *TFilter
	rateLimitError error
	grpcResponse   *TResponse
	grpcError      error
	expectedError  string
	shouldCallGRPC bool
}

// Execute a test on the gRPC calls.
func runForceFetch[TFilter any, TResponse any](
	t *testing.T,
	testCase gRPCTestCase[TFilter, TResponse],
	mockPlayerRedisClient *testutil.MockPlayerRedisClient,
	mockPlayerGRPCClient *testutil.MockPlayerGRPCClient,
	functionName string,
	serviceCall func(ctx context.Context, T *TFilter) (*TResponse, error),
	operation string,
) {
	t.Run(testCase.name, func(t *testing.T) {
		mockBoolCmd := &redis.BoolCmd{}
		if testCase.rateLimitError != nil {
			mockBoolCmd.SetVal(false)
			mockDurationCmd := &redis.DurationCmd{}
			mockDurationCmd.SetVal(time.Minute)
			mockPlayerRedisClient.On("SetNX", mock.AnythingOfType(testutil.DefaultTimerCtx), mock.AnythingOfType("string"), "processing", time.Minute*5).
				Return(mockBoolCmd).Once()
			mockPlayerRedisClient.On("TTL", mock.AnythingOfType(testutil.DefaultTimerCtx), mock.AnythingOfType("string")).
				Return(mockDurationCmd).Once()
		} else {
			mockBoolCmd.SetVal(true)
			mockPlayerRedisClient.On("SetNX", mock.AnythingOfType(testutil.DefaultTimerCtx), mock.AnythingOfType("string"), "processing", time.Minute*5).
				Return(mockBoolCmd).Once()
		}

		if testCase.shouldCallGRPC {
			mockPlayerGRPCClient.On(functionName, mock.Anything, testCase.filters, operation).
				Return(testCase.grpcResponse, testCase.grpcError).Once()
		}

		result, err := serviceCall(context.Background(), testCase.filters)

		if testCase.expectedError != "" {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
			assert.Nil(t, result)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, testCase.grpcResponse, result)
		}

		mockPlayerRedisClient.AssertExpectations(t)
		mockPlayerGRPCClient.AssertExpectations(t)
	})
}

// Test gRPC call used for forcing a player fetch.
func TestForceFetchPlayer(t *testing.T) {
	service, _, _, _, mockPlayerGRPCClient, mockPlayerRedisClient := setupTestService()

	tests := []gRPCTestCase[filters.PlayerForceFetchFilter, pb.Summoner]{
		{
			name: "successful force fetch",
			filters: &filters.PlayerForceFetchFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: nil,
			grpcResponse: &pb.Summoner{
				Puuid:         "test-puuid",
				GameName:      "TestPlayer",
				TagLine:       "TestTag",
				Region:        "TestRegion",
				SummonerLevel: 12,
				ProfileIconId: 123,
			},
			grpcError:      nil,
			expectedError:  "",
			shouldCallGRPC: true,
		},
		{
			name: "rate limit blocked",
			filters: &filters.PlayerForceFetchFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: errors.New("operation already in progress, try again in 60 seconds"),
			grpcResponse:   nil,
			grpcError:      nil,
			expectedError:  "operation already in progress",
			shouldCallGRPC: false,
		},
		{
			name: "grpc client error",
			filters: &filters.PlayerForceFetchFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: nil,
			grpcResponse:   nil,
			grpcError:      errors.New(testutil.GrpcConnectionFailedMessage),
			expectedError:  testutil.GrpcConnectionFailedMessage,
			shouldCallGRPC: true,
		},
	}

	for _, tt := range tests {
		runForceFetch(
			t,
			tt,
			mockPlayerRedisClient,
			mockPlayerGRPCClient,
			"ForceFetchPlayer",
			service.ForceFetchPlayer,
			FORCE_FETCH_OPERATION,
		)
	}
}

// Almost the same as the ForceFetchPlayer, force a player match history to be fetched.
func TestForceFetchPlayerMatchHistory(t *testing.T) {
	service, _, _, _, mockPlayerGRPCClient, mockPlayerRedisClient := setupTestService()

	tests := []gRPCTestCase[filters.PlayerForceFetchMatchListFilter, pb.MatchHistoryFetchNotification]{
		{
			name: "successful force fetch",
			filters: &filters.PlayerForceFetchMatchListFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: nil,
			grpcResponse:   &pb.MatchHistoryFetchNotification{Message: "Started fetching", WillProcess: true},
			grpcError:      nil,
			expectedError:  "",
			shouldCallGRPC: true,
		},
		{
			name: "rate limit blocked",
			filters: &filters.PlayerForceFetchMatchListFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: errors.New("operation already in progress, try again in 60 seconds"),
			grpcResponse:   nil,
			grpcError:      nil,
			expectedError:  "operation already in progress",
			shouldCallGRPC: false,
		},
		{
			name: "grpc client error",
			filters: &filters.PlayerForceFetchMatchListFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			rateLimitError: nil,
			grpcResponse:   nil,
			grpcError:      errors.New(testutil.GrpcConnectionFailedMessage),
			expectedError:  testutil.GrpcConnectionFailedMessage,
			shouldCallGRPC: true,
		},
	}

	for _, tt := range tests {
		runForceFetch(
			t,
			tt,
			mockPlayerRedisClient,
			mockPlayerGRPCClient,
			"ForceFetchPlayerMatchHistory",
			service.ForceFetchPlayerMatchHistory,
			FORCE_FETCH_MATCHES_OPERATION,
		)
	}
}
