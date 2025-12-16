package repositories

import (
	"context"
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

	seeded := seedPlayerTestData(t, db)

	tests := []struct {
		name       string
		filters    *filters.PlayerSearchFilter
		returnData *testutil.OperationRestult[[]*models.PlayerInfo]
		setupFunc  func(db *gorm.DB)
	}{
		{
			name:       "nilfilter",
			filters:    nil,
			returnData: testutil.NewErrorResult[[]*models.PlayerInfo](messages.FiltersNotNil),
		},
		{
			name:       "allfilters",
			filters:    filters.NewPlayerSearchFilter(filters.PlayerSearchParams{Name: "Fa", Tag: "T1", Region: "kr1"}),
			returnData: testutil.NewSuccessResult([]*models.PlayerInfo{seeded["faker"], seeded["fafafa"]}),
		},
		{
			name:       "dbconnectionerr",
			filters:    filters.NewPlayerSearchFilter(filters.PlayerSearchParams{Name: "test", Tag: "br1", Region: "br1"}),
			returnData: testutil.NewErrorResult[[]*models.PlayerInfo]("sql: database is closed"),
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

		if tt.returnData.Err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.returnData.Err.Error())
			assert.Nil(t, result)
			continue
		}

		var expected []*models.PlayerInfo
		for _, player := range tt.returnData.Data {
			expected = append(expected, buildPartialPlayer(player))
		}

		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	}
}

func TestGetPlayerByNameTagRegion(t *testing.T) {
	db, cleanup := testutil.NewTestConnection(t)
	defer cleanup()

	repository := NewPlayerRepository(db)

	seeded := seedPlayerTestData(t, db)

	type playerData struct {
		gameName string
		tagLine  string
		region   string
	}

	tests := []struct {
		name       string
		player     *playerData
		returnData *testutil.OperationRestult[*models.PlayerInfo]
		setupFunc  func(db *gorm.DB)
	}{
		{
			name: "existentplayer",
			player: &playerData{
				gameName: "Faker",
				tagLine:  "T1",
				region:   "KR1",
			},
			returnData: testutil.NewSuccessResult(seeded["faker"]),
		},
		{
			name: "nonexistentplayer",
			player: &playerData{
				gameName: "Goku",
				tagLine:  "GEN",
				region:   "KR1",
			},
			returnData: testutil.NewErrorResult[*models.PlayerInfo]("player not found"),
		},
		{
			name:       "dbconnectionerr",
			returnData: testutil.NewErrorResult[*models.PlayerInfo]("sql: database is closed"),
			player: &playerData{
				gameName: "Faker",
				tagLine:  "T1",
				region:   "KR1",
			},
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

		result, err := repository.GetPlayerByNameTagRegion(context.Background(), tt.player.gameName, tt.player.tagLine, tt.player.region)

		if tt.returnData.Err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.returnData.Err.Error())
			assert.Nil(t, result)
			continue
		}

		// Normalize timestamp to match seeding.
		normalizePlayerTimes([]*models.PlayerInfo{tt.returnData.Data, result})

		assert.Equal(t, tt.returnData.Data, result)
	}
}

func TestGetPlayerById(t *testing.T) {
	db, cleanup := testutil.NewTestConnection(t)
	defer cleanup()

	repository := NewPlayerRepository(db)

	seeded := seedPlayerTestData(t, db)

	tests := []struct {
		name       string
		playerId   uint
		returnData *testutil.OperationRestult[*models.PlayerInfo]
		setupFunc  func(db *gorm.DB)
	}{
		{
			name:       "existentplayer",
			playerId:   12,
			returnData: testutil.NewSuccessResult(seeded["fafafa"]),
		},
		{
			name:       "nonexistentplayer",
			playerId:   20,
			returnData: testutil.NewErrorResult[*models.PlayerInfo]("couldn't get the player by the ID: record not found"),
		},
		{
			name:       "dbconnectionerr",
			playerId:   1,
			returnData: testutil.NewErrorResult[*models.PlayerInfo]("sql: database is closed"),
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

		if tt.returnData.Err != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.returnData.Err.Error())
			assert.Nil(t, result)
			continue
		}

		// Normalize timestamp to match seeding.
		normalizePlayerTimes([]*models.PlayerInfo{tt.returnData.Data, result})

		assert.Equal(t, tt.returnData.Data, result)
	}
}
