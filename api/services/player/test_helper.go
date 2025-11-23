package playerservice

import (
	"goleague/api/services/testutil"

	"gorm.io/gorm"
)

// Helper to initialize the mocks.
func setupTestService() (
	*PlayerService,
	*testutil.MockPlayerRepository,
	*testutil.MockMatchRepository,
	*testutil.MockMatchCache,
	*testutil.MockPlayerGRPCClient,
	*testutil.MockPlayerRedisClient,
) {
	mockPlayerRepo := &testutil.MockPlayerRepository{}
	mockMatchRepo := &testutil.MockMatchRepository{}
	mockMatchCache := &testutil.MockMatchCache{}
	mockPlayerGRPCClient := &testutil.MockPlayerGRPCClient{}
	mockPlayerRedisClient := &testutil.MockPlayerRedisClient{}

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
