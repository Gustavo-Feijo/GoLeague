package playerservice

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	playerrepo "goleague/api/repositories/player"
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
					Puuid:          "test-puuid-search",
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
			repoResponse:   nil,
			repoError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("SearchPlayer", mock.Anything, tt.filter).Return(tt.repoResponse, tt.repoError).Once()

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
		name             string
		filter           *filters.PlayerMatchHistoryFilter
		playerInfo       *models.PlayerInfo
		playerInfoError  error
		matchIds         []uint
		matchIdsError    error
		cachedMatches    []dto.MatchPreview
		missingMatches   []uint
		cacheError       error
		rawPreviews      []matchrepo.RawMatchPreview
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
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			matchIds:        []uint{1, 2},
			matchIdsError:   nil,
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
			playerInfo:      nil,
			playerInfoError: gorm.ErrRecordNotFound,
			expectedError:   fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "no matches found",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			matchIds:        []uint{},
			matchIdsError:   nil,
			expectedError:   "",
		},
		{
			name: "couldn't get matches",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			matchIds:        []uint{},
			matchIdsError:   gorm.ErrInvalidData,
			expectedError:   "unsupported data",
		},
		{
			name: "match previews error",
			filter: &filters.PlayerMatchHistoryFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			matchIds:        []uint{1, 2},
			matchIdsError:   nil,
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
			playerInfo:       &models.PlayerInfo{ID: 1},
			playerInfoError:  nil,
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
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo, tt.playerInfoError).Once()

			if tt.playerInfoError == nil {
				expectedFilter := &filters.PlayerMatchHistoryFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerInfo.ID,
				}
				mockPlayerRepo.On("GetPlayerMatchHistoryIds", mock.Anything, expectedFilter).
					Return(tt.matchIds, tt.matchIdsError).Once()

				if len(tt.matchIds) > 0 {
					mockMatchCache.On("GetMatchesPreviewByMatchIds", mock.Anything, tt.matchIds).
						Return(tt.cachedMatches, tt.missingMatches, tt.cacheError).Once()

					// Cache failed, need to presume all matches are missing on cache.
					if tt.cacheError != nil {
						assert.Equal(t, len(tt.missingMatches), len(tt.matchIds))
					}

					if len(tt.missingMatches) > 0 {
						mockMatchRepo.On("GetMatchPreviewsByInternalIds", mock.Anything, tt.missingMatches).
							Return(tt.rawPreviews, tt.rawPreviewsError).Once()
					}
				}
			}

			result, err := service.GetPlayerMatchHistory(context.Background(), tt.filter)

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

// Test a call to get a given player information.
func TestGetPlayerInfo(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupTestService()

	tests := []struct {
		name            string
		filter          *filters.PlayerInfoFilter
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
			playerInfo: &models.PlayerInfo{
				ID:             1,
				RiotIdGameName: "TestPlayer",
				ProfileIcon:    123,
				Puuid:          "test-puuid-get-info1",
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
			playerInfo:      nil,
			playerInfoError: gorm.ErrInvalidDB,
			expectedError:   fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "player info error",
			filter: &filters.PlayerInfoFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
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
			playerInfo: &models.PlayerInfo{
				ID:             1,
				RiotIdGameName: "TestPlayer",
				ProfileIcon:    123,
				Puuid:          "test-puuid-get-info2",
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
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo, tt.playerInfoError).Once()

			if tt.playerInfoError == nil {
				mockPlayerRepo.On("GetPlayerRatingsById", mock.Anything, tt.playerInfo.ID).
					Return(tt.playerRatings, tt.ratingsError).Once()
			}

			result, err := service.GetPlayerInfo(context.Background(), tt.filter)

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

// Test a fetch for a given player stats history.
func TestGetPlayerStats(t *testing.T) {
	service, mockPlayerRepo, _, _, _, _ := setupTestService()

	tests := []struct {
		name            string
		filter          *filters.PlayerStatsFilter
		playerInfo      *models.PlayerInfo
		playerInfoError error
		playerStats     []playerrepo.RawPlayerStatsStruct
		statsError      error
		expectedError   string
	}{
		{
			name: "successful player stats retrieval",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			playerStats: []playerrepo.RawPlayerStatsStruct{
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
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			playerStats: []playerrepo.RawPlayerStatsStruct{
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
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			playerStats: []playerrepo.RawPlayerStatsStruct{
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
			playerInfo:      nil,
			playerInfoError: gorm.ErrInvalidDB,
			expectedError:   fmt.Sprintf(messages.CouldNotFindId, "player"),
		},
		{
			name: "stats error",
			filter: &filters.PlayerStatsFilter{
				GameName: "TestPlayer",
				GameTag:  "TAG1",
				Region:   "NA1",
			},
			playerInfo:    &models.PlayerInfo{ID: 1},
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
			playerInfo:      &models.PlayerInfo{ID: 1},
			playerInfoError: nil,
			playerStats:     []playerrepo.RawPlayerStatsStruct{},
			statsError:      nil,
			expectedError:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlayerRepo.On("GetPlayerByNameTagRegion", mock.Anything, tt.filter.GameName, tt.filter.GameTag, tt.filter.Region).
				Return(tt.playerInfo, tt.playerInfoError).Once()

			if tt.playerInfoError == nil {
				expectedFilter := &filters.PlayerStatsFilter{
					GameName: tt.filter.GameName,
					GameTag:  tt.filter.GameTag,
					Region:   tt.filter.Region,
					PlayerId: &tt.playerInfo.ID,
				}
				mockPlayerRepo.On("GetPlayerStats", mock.Anything, expectedFilter).
					Return(tt.playerStats, tt.statsError).Once()
			}

			result, err := service.GetPlayerStats(context.Background(), tt.filter)

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
