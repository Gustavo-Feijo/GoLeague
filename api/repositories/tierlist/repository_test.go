package repositories

import (
	"context"
	"errors"
	"goleague/api/filters"
	"goleague/api/repositories/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewTierlistRepository(t *testing.T) {
	repository := NewTierlistRepository(&gorm.DB{})
	assert.NotNil(t, repository)
}

func TestGetTierlist(t *testing.T) {
	db, cleanup := testutil.NewTestConnection(t)
	defer cleanup()

	repository := NewTierlistRepository(db)

	seedTestData(t, db)
	tests := []struct {
		name          string
		filters       *filters.TierlistFilter
		returnData    []*TierlistResult
		expectedError error
		setupFunc     func(db *gorm.DB)
	}{
		{
			name:          "nofilters",
			filters:       filters.NewTierlistFilter(filters.TierlistQueryParams{}),
			expectedError: nil,
			returnData:    getTierlistExpectedResult(t, "nofilters"),
		},
		{
			name:          "allgold",
			filters:       filters.NewTierlistFilter(filters.TierlistQueryParams{Queue: 420, Tier: "GOLD", Rank: "I", AboveTier: false}),
			expectedError: nil,
			returnData:    getTierlistExpectedResult(t, "allgold"),
		},
		{
			name:          "allabovegold",
			filters:       filters.NewTierlistFilter(filters.TierlistQueryParams{Queue: 420, Tier: "GOLD", Rank: "I", AboveTier: true}),
			expectedError: nil,
			returnData:    getTierlistExpectedResult(t, "allabovegold"),
		},
		{
			name:          "dbconnectionerr",
			filters:       filters.NewTierlistFilter(filters.TierlistQueryParams{}),
			returnData:    nil,
			expectedError: errors.New("sql: database is closed"),
			setupFunc: func(db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}
	for _, tt := range tests {
		if tt.setupFunc != nil {
			tt.setupFunc(db)
			sqlDb, _ := db.DB()
			defer sqlDb.Conn(context.Background())
		}

		result, err := repository.GetTierlist(context.Background(), tt.filters)

		if tt.expectedError != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError.Error())
			assert.Nil(t, result)
			return
		}

		assert.Equal(t, tt.returnData, result)
	}
}
