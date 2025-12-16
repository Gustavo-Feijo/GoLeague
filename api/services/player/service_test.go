package playerservice

import (
	"context"
	"fmt"
	"testing"

	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	playerrepo "goleague/api/repositories/player"
	"goleague/internal/testutil"
	"goleague/pkg/database/models"
	"goleague/pkg/messages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Simple test for asserting that everything is fine with the player service creation.
func TestNewPlayerService(t *testing.T) {
	_, _, _, mockMatchCache, mockPlayerGRPCClient, mockPlayerRedisClient := setupTestService()
	deps := &PlayerServiceDeps{
		DB:         new(gorm.DB),
		GrpcClient: mockPlayerGRPCClient,
		MatchCache: mockMatchCache,
		Redis:      mockPlayerRedisClient,
	}

	service := NewPlayerService(deps)
	assert.NotNil(t, service)
	assert.Equal(t, new(gorm.DB), service.db)
	assert.Equal(t, mockPlayerGRPCClient, service.grpcClient)
	assert.NotNil(t, service.MatchRepository)
	assert.NotNil(t, service.PlayerRepository)
}

// Test rate limit key generation.
func TestCreatePlayerRateLimitKey(t *testing.T) {
	service, _, _, _, _, _ := setupTestService()

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

// Test the database player search.
func TestGetPlayerSearch(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupTestService()

	tests := []struct {
		name           string
		filter         *filters.PlayerSearchFilter
		repoResponse   *testutil.OperationRestult[[]*models.PlayerInfo]
		expectedResult []*dto.PlayerSearch
		expectedError  string
	}{
		{
			name:   "successful search",
			filter: &filters.PlayerSearchFilter{},
			repoResponse: testutil.NewSuccessResult([]*models.PlayerInfo{
				{
					ID:             1,
					RiotIdGameName: "TestPlayer",
					ProfileIcon:    123,
					Puuid:          "test-puuid-search",
					Region:         "NA1",
					SummonerLevel:  100,
					RiotIdTagline:  "TAG1",
				},
			}),
			expectedResult: []*dto.PlayerSearch{
				{
					Id:            1,
					Name:          "TestPlayer",
					ProfileIcon:   123,
					Puuid:         "test-puuid-search",
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
			repoResponse:   testutil.NewErrorResult[[]*models.PlayerInfo]("database error"),
			expectedResult: nil,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("SearchPlayer", mock.Anything, tt.filter).Return(tt.repoResponse.Data, tt.repoResponse.Err).Once()

			result, err := service.GetPlayerSearch(context.Background(), tt.filter)

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

// Test a given player match history fetching.
func TestGetPlayerMatchHistory(t *testing.T) {
	service, mockPlayerRepo, mockMatchRepo, mockMatchCache, _, _ := setupTestService()

	tests := []struct {
		name           string
		filter         *filters.PlayerMatchHistoryFilter
		playerInfo     *testutil.OperationRestult[*models.PlayerInfo]
		matchIds       *testutil.OperationRestult[[]uint]
		cachedMatches  *testutil.OperationRestult[[]dto.MatchPreview]
		missingMatches []uint
		rawPreviews    *testutil.OperationRestult[[]matchrepo.RawMatchPreview]
		expectedError  string
	}{
		{
			name: "successful with all matches cached",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			matchIds:   testutil.NewSuccessResult([]uint{1, 2}),
			cachedMatches: testutil.NewSuccessResult([]dto.MatchPreview{
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match1"},
				},
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match2"},
				},
			}),
			rawPreviews:    testutil.NewSuccessResult([]matchrepo.RawMatchPreview{}),
			missingMatches: []uint{},
			expectedError:  "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewErrorResult[*models.PlayerInfo](gorm.ErrRecordNotFound.Error()),
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "no matches found",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			matchIds:      testutil.NewSuccessResult([]uint{}),
			expectedError: "",
		},
		{
			name: "couldn't get matches",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			matchIds:      testutil.NewErrorResult[[]uint](gorm.ErrInvalidData.Error()),
			expectedError: "unsupported data",
		},
		{
			name: "match previews error",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			matchIds:   testutil.NewSuccessResult([]uint{1, 2}),
			cachedMatches: testutil.NewSuccessResult([]dto.MatchPreview{
				{
					Metadata: &dto.MatchPreviewMetadata{MatchId: "match1"},
				},
			}),
			missingMatches: []uint{2},
			expectedError:  "couldn't get the match history",
			rawPreviews:    testutil.NewErrorResult[[]matchrepo.RawMatchPreview]("error"),
		},
		{
			name: "cached match previews error",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:     testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			matchIds:       testutil.NewSuccessResult([]uint{1, 2}),
			cachedMatches:  testutil.NewErrorResult[[]dto.MatchPreview]("cache error"),
			missingMatches: []uint{1, 2},
			expectedError:  "",
			rawPreviews:    testutil.NewSuccessResult([]matchrepo.RawMatchPreview{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo.Data, tt.playerInfo.Err).Once()

			if tt.playerInfo.Err == nil {
				expectedFilter := &filters.PlayerMatchHistoryFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerInfo.Data.ID,
				}
				mockPlayerRepo.On("GetPlayerMatchHistoryIds", mock.Anything, expectedFilter).
					Return(tt.matchIds.Data, tt.matchIds.Err).Once()

				if len(tt.matchIds.Data) > 0 {
					mockMatchCache.On("GetMatchesPreviewByMatchIds", mock.Anything, tt.matchIds.Data).
						Return(tt.cachedMatches.Data, tt.missingMatches, tt.cachedMatches.Err).Once()

					// Cache failed, need to presume all matches are missing on cache.
					if tt.cachedMatches.Err != nil {
						assert.Equal(t, len(tt.missingMatches), len(tt.matchIds.Data))
					}

					if len(tt.missingMatches) > 0 {
						mockMatchRepo.On("GetMatchPreviewsByInternalIds", mock.Anything, tt.missingMatches).
							Return(tt.rawPreviews.Data, tt.rawPreviews.Err).Once()
					}
				}
			}

			result, err := service.GetPlayerMatchHistory(context.Background(), tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if len(tt.matchIds.Data) == 0 {
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

// Test a call to get a given player information.
func TestGetPlayerInfo(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupTestService()

	tests := []struct {
		name          string
		filter        *filters.PlayerInfoFilter
		playerInfo    *testutil.OperationRestult[*models.PlayerInfo]
		playerRatings *testutil.OperationRestult[[]models.RatingEntry]
		expectedError string
	}{
		{
			name: "successful player info retrieval",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(
				&models.PlayerInfo{
					ID:             1,
					RiotIdGameName: "TestPlayer",
					ProfileIcon:    123,
					Puuid:          "test-puuid-get-info1",
					Region:         "NA1",
					SummonerLevel:  100,
					RiotIdTagline:  "TAG1",
				}),
			playerRatings: testutil.NewSuccessResult([]models.RatingEntry{
				{
					LeaguePoints: 1000,
					Losses:       10,
					Queue:        "RANKED_SOLO_5x5",
					Rank:         "I",
					Region:       "NA1",
					Tier:         "GOLD",
					Wins:         20,
				},
			}),
			expectedError: "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewErrorResult[*models.PlayerInfo](gorm.ErrInvalidDB.Error()),
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "player info error",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewErrorResult[*models.PlayerInfo](gorm.ErrInvalidDB.Error()),
			expectedError: "invalid db",
		},
		{
			name: "player rating error",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{
				ID:             1,
				RiotIdGameName: "TestPlayer",
				ProfileIcon:    123,
				Puuid:          "test-puuid-get-info2",
				Region:         "NA1",
				SummonerLevel:  100,
				RiotIdTagline:  "TAG1",
			}),
			playerRatings: testutil.NewErrorResult[[]models.RatingEntry](gorm.ErrInvalidDB.Error()),
			expectedError: "invalid db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo.Data, tt.playerInfo.Err).Once()

			if tt.playerInfo.Err == nil {
				mockPlayerRepo.On("GetPlayerRatingsById", mock.Anything, tt.playerInfo.Data.ID).
					Return(tt.playerRatings.Data, tt.playerRatings.Err).Once()
			}

			result, err := service.GetPlayerInfo(context.Background(), tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.playerInfo.Data.ID, result.Id)
				assert.Equal(t, tt.playerInfo.Data.RiotIdGameName, result.Name)
				assert.Equal(t, len(tt.playerRatings.Data), len(result.Rating))
			}

			mockPlayerRepo.AssertExpectations(t)
		})
	}
}

// Test a fetch for a given player stats history.
func TestGetPlayerStats(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupTestService()

	tests := []struct {
		name          string
		filter        *filters.PlayerStatsFilter
		playerInfo    *testutil.OperationRestult[*models.PlayerInfo]
		playerStats   *testutil.OperationRestult[[]playerrepo.RawPlayerStatsStruct]
		expectedError string
	}{
		{
			name: "successful player stats retrieval",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			playerStats: testutil.NewSuccessResult([]playerrepo.RawPlayerStatsStruct{
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
			}),
			expectedError: "",
		},
		{
			name: "successful player stats retrieval only position filter",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			playerStats: testutil.NewSuccessResult([]playerrepo.RawPlayerStatsStruct{
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
			}),
			expectedError: "",
		},
		{
			name: "successful player stats retrieval only queue filter",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo: testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			playerStats: testutil.NewSuccessResult([]playerrepo.RawPlayerStatsStruct{
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
			}),
			expectedError: "",
		},
		{
			name: "player not found",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewErrorResult[*models.PlayerInfo](gorm.ErrInvalidDB.Error()),
			expectedError: fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "stats error",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			playerStats:   testutil.NewErrorResult[[]playerrepo.RawPlayerStatsStruct](gorm.ErrInvalidDB.Error()),
			expectedError: "couldn't get the player stats",
		},
		{
			name: "no stats found",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    testutil.NewSuccessResult(&models.PlayerInfo{ID: 1}),
			playerStats:   testutil.NewSuccessResult([]playerrepo.RawPlayerStatsStruct{}),
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo.Data, tt.playerInfo.Err).Once()

			if tt.playerInfo.Err == nil {
				expectedFilter := &filters.PlayerStatsFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerInfo.Data.ID,
				}
				mockPlayerRepo.On("GetPlayerStats", mock.Anything, expectedFilter).
					Return(tt.playerStats.Data, tt.playerStats.Err).Once()
			}

			result, err := service.GetPlayerStats(context.Background(), tt.filter)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if len(tt.playerStats.Data) == 0 {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}

			mockPlayerRepo.AssertExpectations(t)
		})
	}
}
