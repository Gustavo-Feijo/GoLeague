package matchservice

import (
	"encoding/json"
	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	servicetestutil "goleague/api/services/testutil"
	"goleague/internal/testutil"
	"goleague/pkg/database/models"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Mock setup struct
type mockSetup struct {
	filters *filters.GetFullMatchDataFilter

	matchCache *servicetestutil.MockMatchCache
	repo       *servicetestutil.MockMatchRepository

	mockMatch    *testutil.RepoGetData[*models.MatchInfo]
	mockPreviews *testutil.RepoGetData[[]matchrepo.RawMatchPreview]
	mockFrames   *testutil.RepoGetData[[]matchrepo.RawMatchParticipantFrame]
	mockEvents   *testutil.RepoGetData[[]models.AllEvents]

	returnData *dto.FullMatchData

	err error
}

// Helper to initialize the mocks.
func setupTestService() (
	*MatchService,
	*servicetestutil.MockMatchRepository,
	*servicetestutil.MockMatchCache,
) {
	mockMatchRepo := new(servicetestutil.MockMatchRepository)
	mockMatchCache := new(servicetestutil.MockMatchCache)

	service := &MatchService{
		db:              new(gorm.DB),
		MatchRepository: mockMatchRepo,
	}

	return service, mockMatchRepo, mockMatchCache
}

func setupMocks(setup mockSetup) {

	if setup.mockMatch != nil {
		setup.repo.On("GetMatchByMatchId", mock.Anything, setup.filters.MatchId).Return(setup.mockMatch.Data, setup.mockMatch.Err)
	}

	if setup.mockPreviews != nil {
		setup.repo.On("GetMatchPreviewsByInternalId", mock.Anything, setup.mockMatch.Data.ID).Return(setup.mockPreviews.Data, setup.mockPreviews.Err)
	}

	if setup.mockFrames != nil {
		setup.repo.On("GetParticipantFramesByInternalId", mock.Anything, setup.mockMatch.Data.ID).Return(setup.mockFrames.Data, setup.mockFrames.Err)
	}

	if setup.mockEvents != nil {
		setup.repo.On("GetAllEvents", mock.Anything, setup.mockMatch.Data.ID).Return(setup.mockEvents.Data, setup.mockEvents.Err)
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
func getEmptyMockPreviews() []matchrepo.RawMatchPreview {
	return []matchrepo.RawMatchPreview{}
}

// Return mocked previews.
func getMockPreviews() []matchrepo.RawMatchPreview {
	return []matchrepo.RawMatchPreview{
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

func getMockFrames() []matchrepo.RawMatchParticipantFrame {
	return []matchrepo.RawMatchParticipantFrame{
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
