package matchservice

import (
	"context"
	"errors"
	"goleague/api/converters"
	"goleague/api/dto"
	"goleague/api/filters"
	matchrepo "goleague/api/repositories/match"
	"goleague/api/services/testutil"
	"goleague/pkg/database/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Simple test for asserting that everything is fine with the match service creation.
func TestNewMatchService(t *testing.T) {
	deps := &MatchServiceDeps{
		DB: new(gorm.DB),
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
		mockPreviews *RepoGetData[[]matchrepo.RawMatchPreview]
		mockFrames   *RepoGetData[[]matchrepo.RawMatchParticipantFrame]
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
			mockPreviews:  getMockRepoError[[]matchrepo.RawMatchPreview](),
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
			mockFrames:    getMockRepoError[[]matchrepo.RawMatchParticipantFrame](),
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
			service, mockMatchRepository, mockMatchCache := setupTestService()

			setupMocks(mockSetup{
				err:        tt.expectedError,
				filters:    tt.filters,
				matchCache: mockMatchCache,

				repo:         mockMatchRepository,
				mockMatch:    tt.mockMatch,
				mockPreviews: tt.mockPreviews,
				mockFrames:   tt.mockFrames,
				mockEvents:   tt.mockEvents,

				returnData: tt.returnData,
			})

			result, err := service.GetFullMatchData(context.Background(), tt.filters)

			assertGetMatchResult(t, result, err, tt.returnData, tt.expectedError)

			testutil.VerifyAllMocks(t, mockMatchCache, mockMatchRepository)
		})
	}
}

func TestGetMatchByMatchId(t *testing.T) {
	service, mockMatchRepo, _ := setupTestService()

	expectedMatch := &models.MatchInfo{MatchId: "NA1_12345"}

	mockMatchRepo.On("GetMatchByMatchId", mock.Anything, "NA1_12345").Return(expectedMatch, nil)

	result, err := service.GetMatchByMatchId(context.Background(), "NA1_12345")

	assert.NoError(t, err)
	assert.Equal(t, expectedMatch, result)
	mockMatchRepo.AssertExpectations(t)
}
