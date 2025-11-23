package testutil

import (
	"context"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/pkg/database/models"
	pb "goleague/pkg/grpc"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock Implementations used on the Player service tests.
// ============================================================================

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

// gRPC Client mock implementations.
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

func (m *MockTierlistRepository) GetTierlist(filters *filters.TierlistFilter) ([]*repositories.TierlistResult, error) {
	args := m.Called(filters)
	return args.Get(0).([]*repositories.TierlistResult), args.Error(1)
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
