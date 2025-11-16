package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/pkg/database/models"
	pb "goleague/pkg/grpc"
	"goleague/pkg/messages"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Player mock implementations.
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) SearchPlayer(filters *filters.PlayerSearchFilter) ([]*models.PlayerInfo, error) {
	args := m.Called(filters)
	return args.Get(0).([]*models.PlayerInfo), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerIdByNameTagRegion(name, tag, region string) (uint, error) {
	args := m.Called(name, tag, region)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerMatchHistoryIds(filters *filters.PlayerMatchHistoryFilter) ([]uint, error) {
	args := m.Called(filters)
	return args.Get(0).([]uint), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerById(id uint) (*models.PlayerInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*models.PlayerInfo), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerRatingsById(id uint) ([]models.RatingEntry, error) {
	args := m.Called(id)
	return args.Get(0).([]models.RatingEntry), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerStats(filters *filters.PlayerStatsFilter) ([]repositories.RawPlayerStatsStruct, error) {
	args := m.Called(filters)
	return args.Get(0).([]repositories.RawPlayerStatsStruct), args.Error(1)
}

// Match mock implementations.
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) GetAllEvents(matchID uint) ([]models.AllEvents, error) {
	args := m.Called(matchID)
	return args.Get(0).([]models.AllEvents), args.Error(1)
}

func (m *MockMatchRepository) GetMatchPreviewsByInternalId(matchID uint) ([]repositories.RawMatchPreview, error) {
	args := m.Called(matchID)
	return args.Get(0).([]repositories.RawMatchPreview), args.Error(1)
}

func (m *MockMatchRepository) GetMatchPreviewsByInternalIds(matchIDs []uint) ([]repositories.RawMatchPreview, error) {
	args := m.Called(matchIDs)
	return args.Get(0).([]repositories.RawMatchPreview), args.Error(1)
}

func (m *MockMatchRepository) GetMatchByMatchId(matchID string) (*models.MatchInfo, error) {
	args := m.Called(matchID)
	return args.Get(0).(*models.MatchInfo), args.Error(1)
}

func (m *MockMatchRepository) GetParticipantFramesByInternalId(matchID uint) ([]repositories.RawMatchParticipantFrame, error) {
	args := m.Called(matchID)
	return args.Get(0).([]repositories.RawMatchParticipantFrame), args.Error(1)
}

// Cache mock implementations.
type MockMatchCache struct {
	mock.Mock
}

func (m *MockMatchCache) GetMatchesPreviewByMatchIds(ctx context.Context, matchIds []uint) ([]dto.MatchPreview, []uint, error) {
	args := m.Called(ctx, matchIds)
	return args.Get(0).([]dto.MatchPreview), args.Get(1).([]uint), args.Error(2)
}

func (m *MockMatchCache) SetMatchPreview(ctx context.Context, preview dto.MatchPreview) error {
	args := m.Called(ctx, preview)
	return args.Error(0)
}

type MockPlayerGRPCClient struct {
	mock.Mock
}

func (m *MockPlayerGRPCClient) ForceFetchPlayer(filters *filters.PlayerForceFetchFilter, operation string) (*pb.Summoner, error) {
	args := m.Called(filters, operation)
	return args.Get(0).(*pb.Summoner), args.Error(1)
}

func (m *MockPlayerGRPCClient) ForceFetchPlayerMatchHistory(filters *filters.PlayerForceFetchMatchListFilter, operation string) (*pb.MatchHistoryFetchNotification, error) {
	args := m.Called(filters, operation)
	return args.Get(0).(*pb.MatchHistoryFetchNotification), args.Error(1)
}

type MockPlayerRedisClient struct {
	mock.Mock
}

func (m *MockPlayerRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockPlayerRedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

// Helper to initialize the mocks.
func setupPlayerService() (*PlayerService, *MockPlayerRepository, *MockMatchRepository, *MockMatchCache, *MockPlayerGRPCClient, *MockPlayerRedisClient) {
	mockPlayerRepo := &MockPlayerRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockMatchCache := &MockMatchCache{}
	mockPlayerGRPCClient := &MockPlayerGRPCClient{}
	mockPlayerRedisClient := &MockPlayerRedisClient{}

	service := &PlayerService{
		db:               &gorm.DB{},
		grpcClient:       mockPlayerGRPCClient,
		matchCache:       mockMatchCache,
		MatchRepository:  mockMatchRepo,
		PlayerRepository: mockPlayerRepo,
		redis:            mockPlayerRedisClient,
	}

	return service, mockPlayerRepo, mockMatchRepo, mockMatchCache, mockPlayerGRPCClient, mockPlayerRedisClient
}

func TestCreatePlayerRateLimitKey(t *testing.T) {
	service, _, _, _, _, _ := setupPlayerService()

	tests := []struct {
		name     string
		gameName string
		gameTag  string
		region   string
		prefix   string
	}{
		{
			name:     "creation",
			gameName: "TestPlayer",
			gameTag:  "TAG1",
			region:   "NA1",
			prefix:   "test",
		},
		{
			name:     "case insensitive",
			gameName: "testplayer",
			gameTag:  "tag1",
			region:   "na1",
			prefix:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.createPlayerRateLimitKey(tt.gameName, tt.gameTag, tt.region, tt.prefix)
			assert.True(t, len(result) > 0)
			assert.Contains(t, result, tt.prefix+":")

			// Hash shouldn't change.
			result2 := service.createPlayerRateLimitKey(tt.gameName, tt.gameTag, tt.region, tt.prefix)
			assert.Equal(t, result, result2)
		})
	}
}

func TestForceFetchPlayer(t *testing.T) {
	service, _, _, _, mockPlayerGRPCClient, mockPlayerRedisClient := setupPlayerService()

	tests := []struct {
		name           string
		filters        *filters.PlayerForceFetchFilter
		rateLimitError error
		grpcResponse   *pb.Summoner
		grpcError      error
		expectedError  string
		shouldCallGRPC bool
	}{
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
			grpcError:      errors.New("grpc connection failed"),
			expectedError:  "grpc connection failed",
			shouldCallGRPC: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoolCmd := &redis.BoolCmd{}
			if tt.rateLimitError != nil {
				mockBoolCmd.SetVal(false)
				mockDurationCmd := &redis.DurationCmd{}
				mockDurationCmd.SetVal(time.Minute)
				mockPlayerRedisClient.On("SetNX", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), "processing", time.Minute*5).
					Return(mockBoolCmd).Once()
				mockPlayerRedisClient.On("TTL", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string")).
					Return(mockDurationCmd).Once()
			} else {
				mockBoolCmd.SetVal(true)
				mockPlayerRedisClient.On("SetNX", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), "processing", time.Minute*5).
					Return(mockBoolCmd).Once()
			}

			if tt.shouldCallGRPC {
				mockPlayerGRPCClient.On("ForceFetchPlayer", tt.filters, FORCE_FETCH_OPERATION).
					Return(tt.grpcResponse, tt.grpcError).Once()
			}

			result, err := service.ForceFetchPlayer(tt.filters)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.grpcResponse, result)
			}

			mockPlayerRedisClient.AssertExpectations(t)
			mockPlayerGRPCClient.AssertExpectations(t)
		})
	}
}

func TestGetPlayerSearch(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupPlayerService()

	tests := []struct {
		name           string
		filter         *filters.PlayerSearchFilter
		repoResponse   []*models.PlayerInfo
		repoError      error
		expectedResult []*dto.PlayerSearch
		expectedError  string
	}{
		{
			name:   "successful search",
			filter: &filters.PlayerSearchFilter{},
			repoResponse: []*models.PlayerInfo{
				{
					ID:             1,
					RiotIdGameName: "TestPlayer",
					ProfileIcon:    123,
					Puuid:          "test-puuid",
					Region:         "NA1",
					SummonerLevel:  100,
					RiotIdTagline:  "TAG1",
				},
			},
			repoError: nil,
			expectedResult: []*dto.PlayerSearch{
				{
					Id:            1,
					Name:          "TestPlayer",
					ProfileIcon:   123,
					Puuid:         "test-puuid",
					Region:        "NA1",
					SummonerLevel: 100,
					Tag:           "TAG1",
				},
			},
			expectedError: "",
		},
		{
			name:           "repository error",
			filter:         &filters.PlayerSearchFilter{},
			repoResponse:   nil,
			repoError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("SearchPlayer", tt.filter).Return(tt.repoResponse, tt.repoError).Once()

			result, err := service.GetPlayerSearch(tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockPlayerRepo.AssertExpectations(t)
		})
	}
}

func TestGetPlayerMatchHistory(t *testing.T) {
	service, mockPlayerRepo, mockMatchRepo, mockMatchCache, _, _ := setupPlayerService()

	tests := []struct {
		name             string
		filter           *filters.PlayerMatchHistoryFilter
		playerId         uint
		playerIdError    error
		matchIds         []uint
		matchIdsError    error
		cachedMatches    []dto.MatchPreview
		missingMatches   []uint
		cacheError       error
		rawPreviews      []repositories.RawMatchPreview
		rawPreviewsError error
		expectedError    string
	}{
		{
			name: "successful with all matches cached",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			matchIds:      []uint{1, 2},
			matchIdsError: nil,
			cachedMatches: []dto.MatchPreview{
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match1"},
				},
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match2"},
				},
			},
			missingMatches: []uint{},
			cacheError:     nil,
			expectedError:  "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      0,
			playerIdError: gorm.ErrRecordNotFound,
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "no matches found",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			matchIds:      []uint{},
			matchIdsError: nil,
			expectedError: "",
		},
		{
			name: "couldn't get matches",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			matchIds:      []uint{},
			matchIdsError: gorm.ErrInvalidData,
			expectedError: "unsupported data",
		},
		{
			name: "match previews error",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			matchIds:      []uint{1, 2},
			matchIdsError: nil,
			cachedMatches: []dto.MatchPreview{
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match1"},
				},
			},
			missingMatches:   []uint{2},
			cacheError:       nil,
			expectedError:    "couldn't get the match history",
			rawPreviews:      nil,
			rawPreviewsError: errors.New("error"),
		},
		{
			name: "cached match previews error",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:         1,
			playerIdError:    nil,
			matchIds:         []uint{1, 2},
			matchIdsError:    nil,
			cachedMatches:    nil,
			missingMatches:   []uint{1, 2},
			cacheError:       errors.New("cache error"),
			expectedError:    "",
			rawPreviews:      nil,
			rawPreviewsError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerIdByNameTagRegion", tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerId, tt.playerIdError).Once()

			if tt.playerIdError == nil {
				expectedFilter := &filters.PlayerMatchHistoryFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerId,
				}
				mockPlayerRepo.On("GetPlayerMatchHistoryIds", expectedFilter).
					Return(tt.matchIds, tt.matchIdsError).Once()

				if len(tt.matchIds) > 0 {
					mockMatchCache.On("GetMatchesPreviewByMatchIds", mock.AnythingOfType("*context.timerCtx"), tt.matchIds).
						Return(tt.cachedMatches, tt.missingMatches, tt.cacheError).Once()

					// Cache failed, need to presume all matches are missing on cache.
					if tt.cacheError != nil {
						assert.Equal(t, len(tt.missingMatches), len(tt.matchIds))
					}

					if len(tt.missingMatches) > 0 {
						mockMatchRepo.On("GetMatchPreviewsByInternalIds", tt.missingMatches).
							Return(tt.rawPreviews, tt.rawPreviewsError).Once()
					}
				}
			}

			result, err := service.GetPlayerMatchHistory(tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if len(tt.matchIds) == 0 {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}

			mockPlayerRepo.AssertExpectations(t)
			mockMatchRepo.AssertExpectations(t)
			mockMatchCache.AssertExpectations(t)
		})
	}
}

func TestGetPlayerInfo(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupPlayerService()

	tests := []struct {
		name            string
		filter          *filters.PlayerInfoFilter
		playerId        uint
		playerIdError   error
		playerInfo      *models.PlayerInfo
		playerInfoError error
		playerRatings   []models.RatingEntry
		ratingsError    error
		expectedError   string
	}{
		{
			name: "successful player info retrieval",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerInfo: &models.PlayerInfo{
				ID:             1,
				RiotIdGameName: "TestPlayer",
				ProfileIcon:    123,
				Puuid:          "test-puuid",
				Region:         "NA1",
				SummonerLevel:  100,
				RiotIdTagline:  "TAG1",
			},
			playerInfoError: nil,
			playerRatings: []models.RatingEntry{
				{
					LeaguePoints: 1000,
					Losses:       10,
					Queue:        "RANKED_SOLO_5x5",
					Rank:         "I",
					Region:       "NA1",
					Tier:         "GOLD",
					Wins:         20,
				},
			},
			ratingsError:  nil,
			expectedError: "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      0,
			playerIdError: gorm.ErrInvalidDB,
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "player info error",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:        1,
			playerIdError:   nil,
			playerInfo:      nil,
			playerInfoError: gorm.ErrInvalidDB,
			expectedError:   "invalid db",
		},
		{
			name: "player rating error",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerInfo: &models.PlayerInfo{
				ID:             1,
				RiotIdGameName: "TestPlayer",
				ProfileIcon:    123,
				Puuid:          "test-puuid",
				Region:         "NA1",
				SummonerLevel:  100,
				RiotIdTagline:  "TAG1",
			},
			playerInfoError: nil,
			playerRatings:   nil,
			ratingsError:    gorm.ErrInvalidDB,
			expectedError:   "invalid db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerIdByNameTagRegion", tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerId, tt.playerIdError).Once()

			if tt.playerIdError == nil {
				mockPlayerRepo.On("GetPlayerById", tt.playerId).
					Return(tt.playerInfo, tt.playerInfoError).Once()

				if tt.playerInfoError == nil {
					mockPlayerRepo.On("GetPlayerRatingsById", tt.playerId).
						Return(tt.playerRatings, tt.ratingsError).Once()
				}
			}

			result, err := service.GetPlayerInfo(tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.playerInfo.ID, result.Id)
				assert.Equal(t, tt.playerInfo.RiotIdGameName, result.Name)
				assert.Equal(t, len(tt.playerRatings), len(result.Rating))
			}

			mockPlayerRepo.AssertExpectations(t)
		})
	}
}

func TestGetPlayerStats(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupPlayerService()

	tests := []struct {
		name          string
		filter        *filters.PlayerStatsFilter
		playerId      uint
		playerIdError error
		playerStats   []repositories.RawPlayerStatsStruct
		statsError    error
		expectedError string
	}{
		{
			name: "successful player stats retrieval",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerStats: []repositories.RawPlayerStatsStruct{
				{
					ChampionId:       1,
					TeamPosition:     "ADC",
					QueueId:          420,
					AggregationLevel: "by_queue",
					AverageAssists:   5.5,
					AverageDeaths:    2.1,
					AverageKills:     8.3,
					CsPerMin:         7.2,
					KDA:              3.9,
					Matches:          10,
					WinRate:          65.0,
				},
			},
			statsError:    nil,
			expectedError: "",
		},
		{
			name: "successful player stats retrieval only position filter",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerStats: []repositories.RawPlayerStatsStruct{
				{
					ChampionId:       400,
					TeamPosition:     "MIDDLE",
					QueueId:          -1,
					AggregationLevel: "by_position",
					AverageAssists:   5.5,
					AverageDeaths:    2.1,
					AverageKills:     8.3,
					CsPerMin:         7.2,
					KDA:              3.9,
					Matches:          10,
					WinRate:          65.0,
				},
			},
			statsError:    nil,
			expectedError: "",
		},
		{
			name: "successful player stats retrieval only queue filter",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerStats: []repositories.RawPlayerStatsStruct{
				{
					ChampionId:       -1,
					TeamPosition:     "ALL",
					QueueId:          -1,
					AggregationLevel: "by_queue",
					AverageAssists:   5.5,
					AverageDeaths:    2.1,
					AverageKills:     8.3,
					CsPerMin:         7.2,
					KDA:              3.9,
					Matches:          10,
					WinRate:          65.0,
				},
			},
			statsError:    nil,
			expectedError: "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      0,
			playerIdError: gorm.ErrInvalidDB,
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "stats error",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			statsError:    gorm.ErrInvalidDB,
			expectedError: "couldn't get the player stats",
		},
		{
			name: "no stats found",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerId:      1,
			playerIdError: nil,
			playerStats:   []repositories.RawPlayerStatsStruct{},
			statsError:    nil,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerIdByNameTagRegion", tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerId, tt.playerIdError).Once()

			if tt.playerIdError == nil {
				expectedFilter := &filters.PlayerStatsFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerId,
				}
				mockPlayerRepo.On("GetPlayerStats", expectedFilter).
					Return(tt.playerStats, tt.statsError).Once()
			}

			result, err := service.GetPlayerStats(tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if len(tt.playerStats) == 0 {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}

			mockPlayerRepo.AssertExpectations(t)
		})
	}
}
