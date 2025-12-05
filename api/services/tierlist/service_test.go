package tierlistservice

import (
	"context"
	"errors"
	"goleague/api/dto"
	"goleague/api/filters"
	tierlistrepo "goleague/api/repositories/tierlist"
	"goleague/api/services/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Simple test for asserting that everything is fine with the tierlist service creation.
func TestNewTierlistService(t *testing.T) {
	_, _, mockMemCache, mockTierlistRedisClient := setupTestService()
	deps := &TierlistServiceDeps{
		DB:       new(gorm.DB),
		MemCache: mockMemCache,
		Redis:    mockTierlistRedisClient,
	}

	service := NewTierlistService(deps)
	assert.NotNil(t, service)
	assert.Equal(t, new(gorm.DB), service.db)
	assert.NotNil(t, service.TierlistRepository)
}

// Run tests on the possible outcomes of the GetTierlist.
func TestGetTierlist(t *testing.T) {

	tests := []struct {
		name                 string
		returnData           []*dto.TierlistResult
		testStrategy         string
		filters              *filters.TierlistFilter
		repositoryReturnData *RepoGetTierlist
		expectedError        error
	}{
		{
			name:         "fromMemCache",
			returnData:   createExpectedSuccessFullTierlist(),
			testStrategy: "memcache",
			filters:      &filters.TierlistFilter{Queue: 420, NumericTier: 1},
		},
		{
			name:         "fromRedis",
			returnData:   createExpectedSuccessFullTierlist(),
			testStrategy: "redis",
			filters:      &filters.TierlistFilter{Queue: 420, NumericTier: 1},
		},
		{
			name:         "fromRepo",
			returnData:   createExpectedSuccessFullTierlist(),
			testStrategy: "nocache",
			filters:      &filters.TierlistFilter{Queue: 420, NumericTier: 1},
			repositoryReturnData: &RepoGetTierlist{
				data: createSuccessRepoTierlist(),
				err:  nil,
			},
		},
		{
			name:         "fromRepoEmpty",
			returnData:   []*dto.TierlistResult{},
			testStrategy: "nocache",
			filters:      &filters.TierlistFilter{Queue: 420, NumericTier: 1},
			repositoryReturnData: &RepoGetTierlist{
				data: []*tierlistrepo.TierlistResult{},
				err:  nil,
			},
		},
		{
			name:         "fromRepoErr",
			returnData:   nil,
			testStrategy: "nocache",
			filters:      &filters.TierlistFilter{Queue: 420, NumericTier: 1},
			repositoryReturnData: &RepoGetTierlist{
				data: nil,
				err:  errors.New(testutil.DatabaseError),
			},
			expectedError: errors.New(testutil.DatabaseError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockTierlistRepository, mockMemCache, mockRedis := setupTestService()

			key := service.getTierlistKey(tt.filters)

			setupMocks(mockSetup{
				err:        tt.expectedError,
				filters:    tt.filters,
				key:        key,
				memCache:   mockMemCache,
				redis:      mockRedis,
				repo:       mockTierlistRepository,
				repoData:   tt.repositoryReturnData,
				returnData: tt.returnData,
				strategy:   tt.testStrategy,
			})

			result, err := service.GetTierlist(context.Background(), tt.filters)

			assertTierlistResult(t, result, err, tt.returnData, tt.expectedError)

			testutil.VerifyAllMocks(t, mockMemCache, mockRedis, mockTierlistRepository)
		})
	}
}

// Simple test to verify behavior when invalid json is returned from redis.
func TestInvalidRedisKey(t *testing.T) {
	key := "testKey"
	service, _, _, mockRedis := setupTestService()

	mockRedis.On("Get", mock.AnythingOfType(testutil.DefaultTimerCtx), key).Return("invalid json", nil)

	result := service.getFromRedis(key)
	assert.Nil(t, result)

	mockRedis.AssertExpectations(t)
}

// Generate multiple keys to verify if key creation is right.
func TestGetTierlistKey(t *testing.T) {
	service, _, _, _ := setupTestService()

	tests := []struct {
		name     string
		filters  *filters.TierlistFilter
		expected string
	}{
		{
			name:     "Basic key",
			filters:  &filters.TierlistFilter{},
			expected: "tierlist",
		},
		{
			name:     "With queue",
			filters:  &filters.TierlistFilter{Queue: 420},
			expected: "tierlist:queue_420",
		},
		{
			name:     "With tier",
			filters:  &filters.TierlistFilter{NumericTier: 1},
			expected: "tierlist:tier_1",
		},
		{
			name:     "With higher tiers flag",
			filters:  &filters.TierlistFilter{GetTiersAbove: true},
			expected: "tierlist:with_higher_tiers",
		},
		{
			name: "With all parameters",
			filters: &filters.TierlistFilter{
				Queue:         420,
				NumericTier:   1,
				GetTiersAbove: true,
			},
			expected: "tierlist:queue_420:tier_1:with_higher_tiers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := service.getTierlistKey(tt.filters)
			assert.Equal(t, tt.expected, key)
		})
	}
}
