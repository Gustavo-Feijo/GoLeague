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
	mockPlayerRepo := new(testutil.MockPlayerRepository)
	mockMatchRepo := new(testutil.MockMatchRepository)
	mockMatchCache := new(testutil.MockMatchCache)
	mockPlayerGRPCClient := new(testutil.MockPlayerGRPCClient)
	mockPlayerRedisClient := new(testutil.MockPlayerRedisClient)

	service := &PlayerService{
		db:               new(gorm.DB),
		grpcClient:       mockPlayerGRPCClient,
		matchCache:       mockMatchCache,
		MatchRepository:  mockMatchRepo,
		PlayerRepository: mockPlayerRepo,
		redis:            mockPlayerRedisClient,
	}

	return service, mockPlayerRepo, mockMatchRepo, mockMatchCache, mockPlayerGRPCClient, mockPlayerRedisClient
}
