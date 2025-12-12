package testutil

import (
	"context"
	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	playerrepo "goleague/api/repositories/player"
	tierlistrepo "goleague/api/repositories/tierlist"
	"goleague/pkg/database/models"
	pb "goleague/pkg/grpc"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// Assert the expectations of all mocks.
func VerifyAllMocks(t *testing.T, mocks ...any) {
	t.Helper()

	for _, m := range mocks {
		if mockObj, ok := m.(interface{ AssertExpectations(*testing.T) bool }); ok {
			mockObj.AssertExpectations(t)
		}
	}
}

// ============================================================================
// Mock Implementations used on the Player service tests.
// ============================================================================

// Player mock implementations.
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) SearchPlayer(ctx context.Context, filters *filters.PlayerSearchFilter) ([]*models.PlayerInfo, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*models.PlayerInfo), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerIdByNameTagRegion(ctx context.Context, name, tag, region string) (uint, error) {
	args := m.Called(ctx, name, tag, region)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerMatchHistoryIds(ctx context.Context, filters *filters.PlayerMatchHistoryFilter) ([]uint, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]uint), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerById(ctx context.Context, id uint) (*models.PlayerInfo, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.PlayerInfo), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerRatingsById(ctx context.Context, id uint) ([]models.RatingEntry, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]models.RatingEntry), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerStats(ctx context.Context, filters *filters.PlayerStatsFilter) ([]playerrepo.RawPlayerStatsStruct, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]playerrepo.RawPlayerStatsStruct), args.Error(1)
}

// Match mock implementations.
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) GetAllEvents(ctx context.Context, matchID uint) ([]models.AllEvents, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0).([]models.AllEvents), args.Error(1)
}

func (m *MockMatchRepository) GetMatchPreviewsByInternalId(ctx context.Context, matchID uint) ([]matchrepo.RawMatchPreview, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0).([]matchrepo.RawMatchPreview), args.Error(1)
}

func (m *MockMatchRepository) GetMatchPreviewsByInternalIds(ctx context.Context, matchIDs []uint) ([]matchrepo.RawMatchPreview, error) {
	args := m.Called(ctx, matchIDs)
	return args.Get(0).([]matchrepo.RawMatchPreview), args.Error(1)
}

func (m *MockMatchRepository) GetMatchByMatchId(ctx context.Context, matchID string) (*models.MatchInfo, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0).(*models.MatchInfo), args.Error(1)
}

func (m *MockMatchRepository) GetParticipantFramesByInternalId(ctx context.Context, matchID uint) ([]matchrepo.RawMatchParticipantFrame, error) {
	args := m.Called(ctx, matchID)
	return args.Get(0).([]matchrepo.RawMatchParticipantFrame), args.Error(1)
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

// gRPC Client mock implementations.
type MockPlayerGRPCClient struct {
	mock.Mock
}

func (m *MockPlayerGRPCClient) ForceFetchPlayer(ctx context.Context, filters *filters.PlayerForceFetchFilter, operation string) (*pb.Summoner, error) {
	args := m.Called(ctx, filters, operation)
	return args.Get(0).(*pb.Summoner), args.Error(1)
}

func (m *MockPlayerGRPCClient) ForceFetchPlayerMatchHistory(ctx context.Context, filters *filters.PlayerForceFetchMatchListFilter, operation string) (*pb.MatchHistoryFetchNotification, error) {
	args := m.Called(ctx, filters, operation)
	return args.Get(0).(*pb.MatchHistoryFetchNotification), args.Error(1)
}

// Player redis client mock implementation.
type MockPlayerRedisClient struct {
	mock.Mock
}

func (m *MockPlayerRedisClient) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockPlayerRedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

// ============================================================================
// Mock Implementations used in the Tierlist service tests.
// ============================================================================

// Tierlist Repo mock implementation.
type MockTierlistRepository struct {
	mock.Mock
}

func (m *MockTierlistRepository) GetTierlist(ctx context.Context, filters *filters.TierlistFilter) ([]*tierlistrepo.TierlistResult, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*tierlistrepo.TierlistResult), args.Error(1)
}

// ChampionCache mock implementation.
type MockChampionCache struct {
	mock.Mock
}

func (m *MockChampionCache) GetChampionCopy(ctx context.Context, championId string) (map[string]any, error) {
	args := m.Called(ctx, championId)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockChampionCache) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MemCache mock implementation.
type MockMemCache[T any] struct {
	mock.Mock
}

func (m *MockMemCache[T]) StartCleanupWorker() {
	m.Called()
}
func (m *MockMemCache[T]) Cleanup() {
	m.Called()
}
func (m *MockMemCache[T]) Close() {
	m.Called()
}
func (m *MockMemCache[T]) Set(key string, value T, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockMemCache[T]) Get(key string) T {
	args := m.Called(key)
	if args.Get(0) == nil {
		var zero T
		return zero
	}
	return args.Get(0).(T)
}

// Redis client mock implementation.
type MockTierlistRedisClient struct {
	mock.Mock
}

func (m *MockTierlistRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockTierlistRedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
