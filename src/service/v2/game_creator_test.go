package v2

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGameCreatorService_GetGameCreators(t *testing.T) {
	t.Parallel()

	job1 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Programmer"), time.Now())
	customJob1 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Custom Job 1"), values.NewGameID(), time.Now())
	creator1 := domain.NewGameCreatorWithJobs(
		domain.NewGameCreator(values.NewGameCreatorID(), values.NewTrapMemberID(uuid.New()), values.NewTrapMemberName("name"), time.Now()),
		[]*domain.GameCreatorJob{job1},
		[]*domain.GameCreatorCustomJob{customJob1},
	)
	creator2 := domain.NewGameCreatorWithJobs(
		domain.NewGameCreator(values.NewGameCreatorID(), values.NewTrapMemberID(uuid.New()), values.NewTrapMemberName("name2"), time.Now()),
		[]*domain.GameCreatorJob{job1},
		[]*domain.GameCreatorCustomJob{},
	)

	testCases := map[string]struct {
		gameID                 values.GameID
		GetGameErr             error
		executeGetGameCreators bool
		creators               []*domain.GameCreatorWithJobs
		GetGameCreatorsErr     error
		err                    error
	}{
		"ok": {
			gameID:                 values.NewGameID(),
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{creator1},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"複数のcreatorがいてもok": {
			gameID:                 values.NewGameID(),
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{creator1, creator2},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"creatorが空でもok": {
			gameID:                 values.NewGameID(),
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"gameが見つからない場合ErrInvalidGameID": {
			gameID:                 values.NewGameID(),
			GetGameErr:             repository.ErrRecordNotFound,
			executeGetGameCreators: false,
			creators:               nil,
			GetGameCreatorsErr:     nil,
			err:                    service.ErrInvalidGameID,
		},
		"GetGameがエラーなのでエラー": {
			gameID:                 values.NewGameID(),
			GetGameErr:             assert.AnError,
			executeGetGameCreators: false,
			creators:               nil,
			GetGameCreatorsErr:     nil,
			err:                    assert.AnError,
		},
		"GetGameCreatorsがエラーなのでエラー": {
			gameID:                 values.NewGameID(),
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               nil,
			GetGameCreatorsErr:     service.ErrNoAsset,
			err:                    service.ErrNoAsset,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			gameRepository := mock.NewMockGameV2(ctrl)
			gc := NewGameCreator(gameCreatorRepo, gameRepository)

			gameRepository.EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)
			if testCase.executeGetGameCreators {
				gameCreatorRepo.EXPECT().
					GetGameCreatorsByGameID(gomock.Any(), testCase.gameID).
					Return(testCase.creators, testCase.GetGameCreatorsErr)
			}

			creators, err := gc.GetGameCreators(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.creators, creators)
		})
	}
}

func TestGetGameCreatorJobs(t *testing.T) {
	t.Parallel()

	presetJob1 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Designer"), time.Now())
	presetJob2 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Producer"), time.Now())
	customJob1 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Custom Job 1"), values.NewGameID(), time.Now())
	customJob2 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Custom Job 2"), values.NewGameID(), time.Now())

	testCases := map[string]struct {
		gameID                          values.GameID
		GetGameErr                      error
		executeGetGameCreatorPresetJobs bool
		presetJobs                      []*domain.GameCreatorJob
		GetGameCreatorPresetJobsErr     error
		executeGetGameCreatorCustomJobs bool
		customJobs                      []*domain.GameCreatorCustomJob
		GetGameCreatorCustomJobsErr     error
		err                             error
	}{
		"GetGameがErrRecordNotFoundの場合ErrInvalidGameID": {
			gameID:     values.NewGameID(),
			GetGameErr: repository.ErrRecordNotFound,
			err:        service.ErrInvalidGameID,
		},
		"GetGameがエラーの場合そのままエラー": {
			gameID:     values.NewGameID(),
			GetGameErr: assert.AnError,
			err:        assert.AnError,
		},
		"GetGameCreatorPresetJobsがエラーの場合エラー": {
			gameID:                          values.NewGameID(),
			executeGetGameCreatorPresetJobs: true,
			GetGameCreatorPresetJobsErr:     assert.AnError,
			err:                             assert.AnError,
		},
		"GetGameCreatorCustomJobsがエラーの場合エラー": {
			gameID:                          values.NewGameID(),
			executeGetGameCreatorPresetJobs: true,
			presetJobs:                      []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobs: true,
			GetGameCreatorCustomJobsErr:     assert.AnError,
			err:                             assert.AnError,
		},
		"正常系": {
			gameID:                          values.NewGameID(),
			executeGetGameCreatorPresetJobs: true,
			presetJobs:                      []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobs: true,
			customJobs:                      []*domain.GameCreatorCustomJob{customJob1, customJob2},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			gameRepository := mock.NewMockGameV2(ctrl)

			gameRepository.EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)
			if testCase.executeGetGameCreatorPresetJobs {
				gameCreatorRepo.EXPECT().
					GetGameCreatorPresetJobs(gomock.Any()).
					Return(testCase.presetJobs, testCase.GetGameCreatorPresetJobsErr)
			}
			if testCase.executeGetGameCreatorCustomJobs {
				gameCreatorRepo.EXPECT().
					GetGameCreatorCustomJobsByGameID(gomock.Any(), testCase.gameID).
					Return(testCase.customJobs, testCase.GetGameCreatorCustomJobsErr)
			}

			gc := NewGameCreator(gameCreatorRepo, gameRepository)

			presetJobs, customJobs, err := gc.GetGameCreatorJobs(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.presetJobs, presetJobs)
			assert.Equal(t, testCase.customJobs, customJobs)
		})
	}
}
