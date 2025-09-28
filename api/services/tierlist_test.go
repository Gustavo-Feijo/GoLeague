package services

import (
	"context"
	"errors"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockTierlistRepository struct {
	mock.Mock
}

func (m *MockTierlistRepository) GetTierlist(filters *filters.TierlistFilter) ([]*repositories.TierlistResult, error) {
	args := m.Called(filters)
	return args.Get(0).([]*repositories.TierlistResult), args.Get(1).(error)
}

type MockChampionCache struct {
	mock.Mock
}

func (m *MockChampionCache) GetChampionCopy(ctx context.Context, championId string) (map[string]any, error) {
	args := m.Called(ctx, championId)
	return args.Get(0).(map[string]any), args.Get(1).(error)
}

func (m *MockChampionCache) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Get(0).(error)
}

type MockMemCache struct {
	mock.Mock
}

func (m *MockMemCache) StartCleanupWorker() {
	m.Called()
}
func (m *MockMemCache) Cleanup() {
	m.Called()
}
func (m *MockMemCache) Close() {
	m.Called()
}
func (m *MockMemCache) Set(key string, value any, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockMemCache) Get(key string) any {
	args := m.Called(key)
	return args.Get(0)
}

type MockTierlistRedisClient struct {
	mock.Mock
}

func (m *MockTierlistRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Get(1).(error)
}

func (m *MockTierlistRedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Get(0).(error)
}

// Helper to initialize the mocks.
func setupTierlistService() (*TierlistService, *MockTierlistRepository, *MockChampionCache, *MockMemCache, *MockTierlistRedisClient) {
	mockTierliStRepository := &MockTierlistRepository{}
	mockChampionCache := &MockChampionCache{}
	mockMemCache := &MockMemCache{}
	mockRedisTierlistClient := &MockTierlistRedisClient{}

	service := &TierlistService{
		db:                 &gorm.DB{},
		championCache:      mockChampionCache,
		memCache:           mockMemCache,
		redis:              mockRedisTierlistClient,
		TierlistRepository: mockTierliStRepository,
	}

	return service, mockTierliStRepository, mockChampionCache, mockMemCache, mockRedisTierlistClient
}

func createExpectedFullTierlist() []*dto.FullTierlist {
	return []*dto.FullTierlist{
		{
			BanCount:     10,
			Banrate:      0.15,
			PickCount:    50,
			PickRate:     0.25,
			TeamPosition: "ADC",
			WinRate:      0.52,
			Champion: map[string]any{
				"id":    "1",
				"name":  "TestChampion",
				"title": "The Test",
				"image": "test.jpg",
			},
		},
		{
			BanCount:     5,
			Banrate:      0.08,
			PickCount:    30,
			PickRate:     0.18,
			TeamPosition: "MID",
			WinRate:      0.48,
			Champion: map[string]any{
				"id":    "2",
				"name":  "TestChampion",
				"title": "The Test",
				"image": "test.jpg",
			},
		},
	}
}

func TestGetTierlistFromMemCache(t *testing.T) {
	service, _, _, mockMemCache, _ := setupTierlistService()

	filters := &filters.TierlistFilter{Queue: 420, NumericTier: 1}
	expectedData := createExpectedFullTierlist()

	mockMemCache.On("Get", "tierlist:queue_420:tier_1").Return(expectedData)

	result, err := service.GetTierlist(filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	mockMemCache.AssertExpectations(t)
}

func TestGetTierlistRepositoryErrorReturnsError(t *testing.T) {
	service, mockRepo, _, mockMemCache, mockRedis := setupTierlistService()

	filters := &filters.TierlistFilter{Queue: 420, NumericTier: 1}
	key := "tierlist:queue_420:tier_1"

	// Mock cache misses
	mockMemCache.On("Get", key).Return(nil)
	mockRedis.On("Get", mock.AnythingOfType("*context.timerCtx"), key).Return("", errors.New("not found"))

	// Mock repository error
	mockRepo.On("GetTierlist", filters).Return(([]*repositories.TierlistResult)(nil), errors.New("database error"))

	result, err := service.GetTierlist(filters)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockMemCache.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestGetTierlistKey(t *testing.T) {
	service, _, _, _, _ := setupTierlistService()

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
