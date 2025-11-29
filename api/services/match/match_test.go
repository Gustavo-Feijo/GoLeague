package matchservice

import (
	"errors"
	"goleague/api/converters"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/api/services/testutil"
	"goleague/pkg/database/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Simple test for asserting that everything is fine with the match service creation.
func TestNewMatchService(t *testing.T) {
	_, _, _, mockMemCache := setupTestService()
	deps := &MatchServiceDeps{
		DB:       new(gorm.DB),
		MemCache: mockMemCache,
	}

	service := NewMatchService(deps)
	assert.NotNil(t, service)
	assert.Equal(t, new(gorm.DB), service.db)
	assert.NotNil(t, service.MatchRepository)
}

func TestGetFullMatchData(t *testing.T) {
	defaultFilter := &filters.GetFullMatchDataFilter{MatchId: "BR1_Test"}

	tests := []struct {
		name       string
		returnData *dto.FullMatchData
		filters    *filters.GetFullMatchDataFilter

		mockMatch    *RepoGetData[*models.MatchInfo]
		mockPreviews *RepoGetData[[]repositories.RawMatchPreview]
		mockFrames   *RepoGetData[[]repositories.RawMatchParticipantFrame]
		mockEvents   *RepoGetData[[]models.AllEvents]

		expectedError error
	}{
		{
			name:          "matchNotFoundDbError",
			returnData:    &dto.FullMatchData{},
			filters:       defaultFilter,
			mockMatch:     getMockRepoError[*models.MatchInfo](),
			expectedError: errors.New(testutil.DatabaseError),
		},
		{
			name:          "previewsNotFoundDbError",
			returnData:    &dto.FullMatchData{},
			filters:       defaultFilter,
			mockMatch:     toRepoGetData(getMockMatch()),
			mockPreviews:  getMockRepoError[[]repositories.RawMatchPreview](),
			expectedError: errors.New(testutil.DatabaseError),
		},
		{
			name:          "emptyPreviews",
			returnData:    &dto.FullMatchData{},
			filters:       defaultFilter,
			mockMatch:     toRepoGetData(getMockMatch()),
			mockPreviews:  toRepoGetData(getEmptyMockPreviews()),
			expectedError: errors.New(converters.ErrNoPreviews),
		},
		{
			name:          "framesNotFoundDbErr",
			returnData:    &dto.FullMatchData{},
			filters:       defaultFilter,
			mockMatch:     toRepoGetData(getMockMatch()),
			mockPreviews:  toRepoGetData(getMockPreviews()),
			mockFrames:    getMockRepoError[[]repositories.RawMatchParticipantFrame](),
			expectedError: errors.New(testutil.DatabaseError),
		},
		{
			name:          "eventsNotFoundDbErr",
			returnData:    &dto.FullMatchData{},
			filters:       defaultFilter,
			mockMatch:     toRepoGetData(getMockMatch()),
			mockPreviews:  toRepoGetData(getMockPreviews()),
			mockFrames:    toRepoGetData(getMockFrames()),
			mockEvents:    getMockRepoError[[]models.AllEvents](),
			expectedError: errors.New(testutil.DatabaseError),
		},
		{
			name:          "everythingFine",
			returnData:    loadExpectedData[*dto.FullMatchData]("testdata/fullmatch.json"),
			filters:       defaultFilter,
			mockMatch:     toRepoGetData(getMockMatch()),
			mockPreviews:  toRepoGetData(getMockPreviews()),
			mockFrames:    toRepoGetData(getMockFrames()),
			mockEvents:    toRepoGetData(getMockEvents()),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockMatchRepository, mockMatchCache, mockMemCache := setupTestService()

			setupMocks(mockSetup{
				err:        tt.expectedError,
				filters:    tt.filters,
				matchCache: mockMatchCache,
				memCache:   mockMemCache,

				repo:         mockMatchRepository,
				mockMatch:    tt.mockMatch,
				mockPreviews: tt.mockPreviews,
				mockFrames:   tt.mockFrames,
				mockEvents:   tt.mockEvents,

				returnData: tt.returnData,
			})

			result, err := service.GetFullMatchData(tt.filters)

			assertGetMatchResult(t, result, err, tt.returnData, tt.expectedError)

			testutil.VerifyAllMocks(t, mockMemCache, mockMatchCache, mockMatchRepository)
		})
	}
}

func TestGetMatchByMatchId(t *testing.T) {
	service, mockMatchRepo, _, _ := setupTestService()

	expectedMatch := &models.MatchInfo{MatchId: "NA1_12345"}

	mockMatchRepo.On("GetMatchByMatchId", "NA1_12345").Return(expectedMatch, nil)

	result, err := service.GetMatchByMatchId("NA1_12345")

	assert.NoError(t, err)
	assert.Equal(t, expectedMatch, result)
	mockMatchRepo.AssertExpectations(t)
}
