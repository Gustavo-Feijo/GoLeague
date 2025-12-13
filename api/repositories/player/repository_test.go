package repositories

import (
	"context"
	"errors"
	"fmt"
	"goleague/api/filters"
	"goleague/internal/testutil"
	"goleague/pkg/database/models"
	"goleague/pkg/messages"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewPlayerRepository(t *testing.T) {
	repository := NewPlayerRepository(&gorm.DB{})
	assert.NotNil(t, repository)
}

func TestSearchPlayer(t *testing.T) {
	db, cleanup := testutil.NewTestConnection(t)
	defer cleanup()

	repository := NewPlayerRepository(db)

	seedPlayerTestData(t, db)

	tests := []struct {
		name          string
		filters       *filters.PlayerSearchFilter
		returnData    []*models.PlayerInfo
		expectedError error
		setupFunc     func(db *gorm.DB)
	}{
		{
			name:          "nilfilter",
			filters:       nil,
			expectedError: fmt.Errorf(messages.FiltersNotNil),
		},
		{
			name:          "allfilters",
			filters:       filters.NewPlayerSearchFilter(filters.PlayerSearchParams{Name: "Fa", Tag: "T1", Region: "kr1"}),
			returnData:    getPlayerSearchExpectedResult(t, "allfilters"),
			expectedError: nil,
		},
		{
			name:          "dbconnectionerr",
			filters:       filters.NewPlayerSearchFilter(filters.PlayerSearchParams{Name: "test", Tag: "br1", Region: "br1"}),
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

		result, err := repository.SearchPlayer(context.Background(), tt.filters)

		if tt.expectedError != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError.Error())
			assert.Nil(t, result)
			continue
		}

		assert.Equal(t, tt.returnData, result)
	}
}

func TestGetPlayerById(t *testing.T) {
	db, cleanup := testutil.NewTestConnection(t)
	defer cleanup()

	repository := NewPlayerRepository(db)

	seedPlayerTestData(t, db)

	tests := []struct {
		name          string
		playerId      uint
		returnData    *models.PlayerInfo
		expectedError error
		setupFunc     func(db *gorm.DB)
	}{
		{
			name:          "existentplayer",
			playerId:      12,
			returnData:    getPlayerByIdExpectedResult(t, "existentplayer"),
			expectedError: nil,
		},
		{
			name:          "nonexistentplayer",
			playerId:      20,
			returnData:    nil,
			expectedError: fmt.Errorf("couldn't get the player by the ID: record not found"),
		},
		{
			name:          "dbconnectionerr",
			playerId:      1,
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

		result, err := repository.GetPlayerById(context.Background(), tt.playerId)

		if tt.expectedError != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError.Error())
			assert.Nil(t, result)
			continue
		}

		// Normalize timestamp to match seeding.
		normalizePlayerTimes([]*models.PlayerInfo{result})

		assert.Equal(t, tt.returnData, result)
	}
}
