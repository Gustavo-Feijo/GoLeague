package tierlistservice

import (
	"encoding/json"
	"goleague/api/dto"
	"goleague/api/filters"
	tierlistrepo "goleague/api/repositories/tierlist"
	servicetestutil "goleague/api/services/testutil"
	"goleague/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock setup struct
type mockSetup struct {
	filters  *filters.TierlistFilter
	key      string
	strategy string

	memCache *servicetestutil.MockMemCache[[]*dto.TierlistResult]
	redis    *servicetestutil.MockTierlistRedisClient
	repo     *servicetestutil.MockTierlistRepository

	repoData *testutil.RepoGetData[[]*tierlistrepo.TierlistResult]

	expectedResult []*dto.TierlistResult
	err            error
}

// Helper to initialize the mocks.
func setupTestService() (*TierlistService, *servicetestutil.MockTierlistRepository, *servicetestutil.MockMemCache[[]*dto.TierlistResult], *servicetestutil.MockTierlistRedisClient) {
	mockTierlistRepository := new(servicetestutil.MockTierlistRepository)
	mockMemCache := new(servicetestutil.MockMemCache[[]*dto.TierlistResult])
	mockRedisTierlistClient := new(servicetestutil.MockTierlistRedisClient)

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
func createSuccessRepoTierlist() []*tierlistrepo.TierlistResult {
	return []*tierlistrepo.TierlistResult{
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
	setup.memCache.On("Get", setup.key).Return(setup.expectedResult)
}

// Not available on memory, but available on Redis.
func setupRedisCacheHit(setup mockSetup) {
	setup.memCache.On("Get", setup.key).Return(nil)

	data, _ := json.Marshal(setup.expectedResult)
	setup.redis.On("Get", mock.AnythingOfType(servicetestutil.DefaultTimerCtx), setup.key).Return(string(data), nil)
	setup.memCache.On("Set", setup.key, setup.expectedResult, TierlistMemoryCacheDuration).Return(nil)
}

// Data available only on database.
func setupNoCacheHit(setup mockSetup) {
	setup.memCache.On("Get", setup.key).Return(nil)
	setup.redis.On("Get", mock.AnythingOfType(servicetestutil.DefaultTimerCtx), setup.key).Return("", nil)

	setup.repo.On("GetTierlist", mock.Anything, setup.filters).Return(setup.repoData.Data, setup.repoData.Err)

	setup.memCache.On("Set", setup.key, setup.expectedResult, TierlistMemoryCacheDuration).Return(nil)

	data, _ := json.Marshal(setup.expectedResult)
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
