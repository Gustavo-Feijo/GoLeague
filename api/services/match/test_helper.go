package matchservice

import (
	"encoding/json"
	"errors"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/api/services/testutil"
	"goleague/pkg/database/models"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RepoGetData[T any] struct {
	data T
	err  error
}

// Mock setup struct
type mockSetup struct {
	filters *filters.GetFullMatchDataFilter

	memCache   *testutil.MockMemCache
	matchCache *testutil.MockMatchCache
	repo       *testutil.MockMatchRepository

	mockMatch    *RepoGetData[*models.MatchInfo]
	mockPreviews *RepoGetData[[]repositories.RawMatchPreview]
	mockFrames   *RepoGetData[[]repositories.RawMatchParticipantFrame]
	mockEvents   *RepoGetData[[]models.AllEvents]

	returnData *dto.FullMatchData

	err error
}

// Helper to initialize the mocks.
func setupTestService() (
	*MatchService,
	*testutil.MockMatchRepository,
	*testutil.MockMatchCache,
	*testutil.MockMemCache,
) {
	mockMatchRepo := new(testutil.MockMatchRepository)
	mockMemCache := new(testutil.MockMemCache)
	mockMatchCache := new(testutil.MockMatchCache)

	service := &MatchService{
		db:              new(gorm.DB),
		memCache:        mockMemCache,
		MatchRepository: mockMatchRepo,
	}

	return service, mockMatchRepo, mockMatchCache, mockMemCache
}

func setupMocks(setup mockSetup) {

	if setup.mockMatch != nil {
		setup.repo.On("GetMatchByMatchId", setup.filters.MatchId).Return(setup.mockMatch.data, setup.mockMatch.err)
	}

	if setup.mockPreviews != nil {
		setup.repo.On("GetMatchPreviewsByInternalId", setup.mockMatch.data.ID).Return(setup.mockPreviews.data, setup.mockPreviews.err)
	}

	if setup.mockFrames != nil {
		setup.repo.On("GetParticipantFramesByInternalId", setup.mockMatch.data.ID).Return(setup.mockFrames.data, setup.mockFrames.err)
	}

	if setup.mockEvents != nil {
		setup.repo.On("GetAllEvents", setup.mockMatch.data.ID).Return(setup.mockEvents.data, setup.mockEvents.err)
	}
}

// Return a generic typed error return for a database call.
func getMockRepoError[T any]() *RepoGetData[T] {
	return &RepoGetData[T]{
		data: *new(T),
		err:  errors.New(testutil.DatabaseError),
	}
}

// Wrap a generic data into a RepoGetData struct.
func toRepoGetData[T any](data T) *RepoGetData[T] {
	return &RepoGetData[T]{
		data: data,
		err:  nil,
	}
}

// Return a simple mock for a match return.
func getMockMatch() *models.MatchInfo {
	return &models.MatchInfo{
		ID:             1,
		GameVersion:    "15.23.726.9074",
		MatchId:        "BR1_3169094685",
		MatchStart:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		MatchDuration:  993,
		MatchWinner:    100,
		MatchSurrender: true,
		MatchRemake:    false,
		AverageRating:  15000,
		FrameInterval:  60000,
		FullyFetched:   true,
		QueueId:        900,
		CreatedAt:      time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
	}
}

// Return a mock for empty previews return.
func getEmptyMockPreviews() []repositories.RawMatchPreview {
	return []repositories.RawMatchPreview{}
}

// Return mocked previews.
func getMockPreviews() []repositories.RawMatchPreview {
	return []repositories.RawMatchPreview{
		{

			Assists:              6,
			AverageRating:        0,
			ChampionID:           67,
			ChampionLevel:        19,
			Deaths:               8,
			Duration:             993,
			InternalId:           1,
			Kills:                9,
			MatchID:              "BR1_TEST",
			NeutralMinionsKilled: 22,
			ParticipantId:        1,
			PlayerId:             9,
			QueueID:              900,
			Region:               "BR1",
			RiotIDGameName:       "TestPlayer1",
			RiotIDTagline:        "br1",
			Team:                 200,
			TotalMinionsKilled:   95,
			Win:                  false,
			WinnerTeamId:         100,
		},
		{
			Assists:              10,
			AverageRating:        0,
			ChampionID:           42,
			ChampionLevel:        20,
			Deaths:               7,
			Duration:             993,
			InternalId:           2,
			Kills:                9,
			NeutralMinionsKilled: 27,
			MatchID:              "BR1_TEST",
			ParticipantId:        2,
			PlayerId:             1,
			QueueID:              900,
			Region:               "BR1",
			RiotIDGameName:       "TestPlayer2",
			RiotIDTagline:        "br1",
			Team:                 200,
			TotalMinionsKilled:   142,
			Win:                  false,
			WinnerTeamId:         100,
		},
	}
}

// Return mock events for the matches.
func getMockEvents() []models.AllEvents {
	p1 := 1
	p2 := 2

	return []models.AllEvents{
		{
			MatchId:       1,
			Timestamp:     4999,
			EventType:     "item",
			ParticipantId: &p1,
			Data: datatypes.JSON([]byte(`{
                "item_id": 1055,
                "after_id": null,
                "action": "ITEM_PURCHASED"
            }`)),
		},
		{
			MatchId:       1,
			Timestamp:     5266,
			EventType:     "item",
			ParticipantId: &p2,
			Data: datatypes.JSON([]byte(`{
                "item_id": 2003,
                "after_id": null,
                "action": "ITEM_PURCHASED"
            }`)),
		},
		{
			MatchId:       1,
			Timestamp:     192903,
			EventType:     "kill_struct",
			ParticipantId: nil,
			Data: datatypes.JSON([]byte(`{
                "building_type": null,
                "event_type": "TURRET_PLATE_DESTROYED",
                "lane_type": "MID_LANE",
                "team_id": 100,
                "tower_type": null,
                "x": 5846,
                "y": 6396
            }`)),
		},
	}
}

func getMockFrames() []repositories.RawMatchParticipantFrame {
	return []repositories.RawMatchParticipantFrame{
		{
			CurrentGold:                   500,
			FrameIndex:                    0,
			JungleMinionsKilled:           0,
			Level:                         1,
			MagicDamageDone:               0,
			MagicDamageDoneToChampions:    0,
			MagicDamageTaken:              0,
			MatchStatID:                   1,
			MinionsKilled:                 0,
			ParticipantID:                 1,
			PhysicalDamageDone:            0,
			PhysicalDamageDoneToChampions: 0,
			PhysicalDamageTaken:           0,
			TotalDamageDone:               0,
			TotalDamageDoneToChampions:    0,
			TotalDamageTaken:              0,
			TotalGold:                     500,
			TrueDamageDone:                0,
			TrueDamageDoneToChampions:     0,
			TrueDamageTaken:               0,
			XP:                            0,
		},
		{
			CurrentGold:                   150,
			FrameIndex:                    1,
			JungleMinionsKilled:           0,
			Level:                         1,
			MagicDamageDone:               0,
			MagicDamageDoneToChampions:    0,
			MagicDamageTaken:              0,
			MatchStatID:                   2,
			MinionsKilled:                 0,
			ParticipantID:                 2,
			PhysicalDamageDone:            265,
			PhysicalDamageDoneToChampions: 265,
			PhysicalDamageTaken:           0,
			TotalDamageDone:               302,
			TotalDamageDoneToChampions:    302,
			TotalDamageTaken:              0,
			TotalGold:                     500,
			TrueDamageDone:                37,
			TrueDamageDoneToChampions:     37,
			TrueDamageTaken:               0,
			XP:                            0,
		},
	}
}

// Load a given file from the disk and return it as a JSON.
func loadExpectedData[T any](path string) T {
	var data T

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return *new(T)
	}

	_ = json.Unmarshal(jsonData, &data)
	return data
}

// Assert the expected returned results.
func assertGetMatchResult(
	t *testing.T,
	result *dto.FullMatchData,
	err error,
	expectedData *dto.FullMatchData,
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

	// dto.FullMatchData has some datatypes.JSON fields, without converting it can be different due to formatting.
	// Simply parse it to be reliable.
	expectedJSON, _ := json.Marshal(expectedData)
	resultJSON, _ := json.Marshal(result)

	assert.JSONEq(t, string(expectedJSON), string(resultJSON))
}
