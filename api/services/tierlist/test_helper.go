package tierlistservice

import (
	"encoding/json"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/api/services/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type RepoGetTierlist struct {
	data []*repositories.TierlistResult
	err  error
}

// Mock setup struct
type mockSetup struct {
	filters  *filters.TierlistFilter
	key      string
	strategy string

	memCache *testutil.MockMemCache
	redis    *testutil.MockTierlistRedisClient
	repo     *testutil.MockTierlistRepository

	repoData *RepoGetTierlist

	returnData []*dto.TierlistResult
	err        error
}

// Helper to initialize the mocks.
func setupTestService() (*TierlistService, *testutil.MockTierlistRepository, *testutil.MockMemCache, *testutil.MockTierlistRedisClient) {
	mockTierlistRepository := new(testutil.MockTierlistRepository)
	mockMemCache := new(testutil.MockMemCache)
	mockRedisTierlistClient := new(testutil.MockTierlistRedisClient)

	service := &TierlistService{
		db:                 new(gorm.DB),
		memCache:           mockMemCache,
		redis:              mockRedisTierlistClient,
		TierlistRepository: mockTierlistRepository,
	}

	return service, mockTierlistRepository, mockMemCache, mockRedisTierlistClient
}

// Create a tierlist correct example.
func createExpectedSuccessFullTierlist() []*dto.TierlistResult {
	return []*dto.TierlistResult{
		{
			BanCount:     10,
			BanRate:      0.15,
			PickCount:    50,
			PickRate:     0.25,
			TeamPosition: "ADC",
			WinRate:      0.52,
			ChampionId:   1,
		},
		{
			BanCount:     5,
			BanRate:      0.08,
			PickCount:    30,
			PickRate:     0.18,
			TeamPosition: "MID",
			WinRate:      0.48,
			ChampionId:   2,
		},
	}
}

// Create a tierlist correct repository return.
func createSuccessRepoTierlist() []*repositories.TierlistResult {
	return []*repositories.TierlistResult{
		{
			BanCount:     10,
			BanRate:      0.15,
			PickCount:    50,
			PickRate:     0.25,
			TeamPosition: "ADC",
			WinRate:      0.52,
			ChampionId:   1,
		},
		{
			BanCount:     5,
			BanRate:      0.08,
			PickCount:    30,
			PickRate:     0.18,
			TeamPosition: "MID",
			WinRate:      0.48,
			ChampionId:   2,
		},
	}
}

// Setup the mocks for the tierlist test based on cache strategy.
func setupMocks(setup mockSetup) {
	switch setup.strategy {
	case "memcache":
		setupMemCacheHit(setup)
	case "redis":
		setupRedisCacheHit(setup)
	case "nocache":
		setupNoCacheHit(setup)
	}
}

// Data already available on memory.
func setupMemCacheHit(setup mockSetup) {
	setup.memCache.On("Get", setup.key).Return(setup.returnData)
}

// Not available on memory, but available on Redis.
func setupRedisCacheHit(setup mockSetup) {
	setup.memCache.On("Get", setup.key).Return(nil)

	data, _ := json.Marshal(setup.returnData)
	setup.redis.On("Get", mock.AnythingOfType(testutil.DefaultTimerCtx), setup.key).Return(string(data), nil)
	setup.memCache.On("Set", setup.key, setup.returnData, TierlistMemoryCacheDuration).Return(nil)
}

// Data available only on database.
func setupNoCacheHit(setup mockSetup) {
	setup.memCache.On("Get", setup.key).Return(nil)
	setup.redis.On("Get", mock.AnythingOfType(testutil.DefaultTimerCtx), setup.key).Return("", nil)

	setup.repo.On("GetTierlist", setup.filters).Return(setup.repoData.data, setup.repoData.err)

	setup.memCache.On("Set", setup.key, setup.returnData, TierlistMemoryCacheDuration).Return(nil)

	data, _ := json.Marshal(setup.returnData)
	setup.redis.On("Set", mock.Anything, setup.key, string(data), TierlistRedisCacheDuration).Return(nil)
}

// Assert the expected returned results.
func assertTierlistResult(
	t *testing.T,
	result []*dto.TierlistResult,
	err error,
	expectedData []*dto.TierlistResult,
	expectedError error,
) {
	t.Helper()

	if expectedError != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Nil(t, result)
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, len(expectedData), len(result))
	assert.Equal(t, expectedData, result)
}

// Assert the expectations of all mocks.
func verifyAllMocks(t *testing.T, mocks ...any) {
	t.Helper()

	for _, m := range mocks {
		if mockObj, ok := m.(interface{ AssertExpectations(*testing.T) bool }); ok {
			mockObj.AssertExpectations(t)
		}
	}
}
