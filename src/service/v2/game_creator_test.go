package v2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGameCreatorService_GetGameCreators(t *testing.T) {
	t.Parallel()

	gameID := values.NewGameID()
	job1 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Programmer"), time.Now())
	customJob1 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Custom Job 1"), gameID, time.Now())
	creator1 := domain.NewGameCreatorWithJobs(
		domain.NewGameCreator(values.NewGameCreatorID(), values.NewTrapMemberID(uuid.New()), gameID, values.NewTrapMemberName("name"), time.Now()),
		[]*domain.GameCreatorJob{job1},
		[]*domain.GameCreatorCustomJob{customJob1},
	)
	creator2 := domain.NewGameCreatorWithJobs(
		domain.NewGameCreator(values.NewGameCreatorID(), values.NewTrapMemberID(uuid.New()), gameID, values.NewTrapMemberName("name2"), time.Now()),
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
			gameID:                 gameID,
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{creator1},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"複数のcreatorがいてもok": {
			gameID:                 gameID,
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{creator1, creator2},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"creatorが空でもok": {
			gameID:                 gameID,
			GetGameErr:             nil,
			executeGetGameCreators: true,
			creators:               []*domain.GameCreatorWithJobs{},
			GetGameCreatorsErr:     nil,
			err:                    nil,
		},
		"gameが見つからない場合ErrInvalidGameID": {
			gameID:                 gameID,
			GetGameErr:             repository.ErrRecordNotFound,
			executeGetGameCreators: false,
			creators:               nil,
			GetGameCreatorsErr:     nil,
			err:                    service.ErrInvalidGameID,
		},
		"GetGameがエラーなのでエラー": {
			gameID:                 gameID,
			GetGameErr:             assert.AnError,
			executeGetGameCreators: false,
			creators:               nil,
			GetGameCreatorsErr:     nil,
			err:                    assert.AnError,
		},
		"GetGameCreatorsがエラーなのでエラー": {
			gameID:                 gameID,
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
			gc := NewGameCreator(gameCreatorRepo, gameRepository, nil, nil)

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

			gc := NewGameCreator(gameCreatorRepo, gameRepository, nil, nil)

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

func TestCreateGameCustomJob(t *testing.T) {
	t.Parallel()

	gameID := values.NewGameID()
	displayName := values.NewGameCreatorJobDisplayName("Sound Designer")
	existingCustomJob := domain.NewGameCreatorCustomJob(
		values.NewGameCreatorJobID(),
		values.NewGameCreatorJobDisplayName("Programmer"),
		gameID,
		time.Now(),
	)
	duplicatedCustomJob := domain.NewGameCreatorCustomJob(
		values.NewGameCreatorJobID(),
		displayName,
		gameID,
		time.Now(),
	)

	testCases := map[string]struct {
		getGameErr                  error
		executeGetCustomJobs        bool
		existingCustomJobs          []*domain.GameCreatorCustomJob
		getCustomJobsErr            error
		executeCreateGameCustomJobs bool
		createGameCustomJobsErr     error
		wantErr                     error
	}{
		"正常に作成できる": {
			executeGetCustomJobs:        true,
			existingCustomJobs:          []*domain.GameCreatorCustomJob{existingCustomJob},
			executeCreateGameCustomJobs: true,
		},
		"gameが存在しない場合ErrInvalidGameID": {
			getGameErr: repository.ErrRecordNotFound,
			wantErr:    service.ErrInvalidGameID,
		},
		"GetGameがエラーの場合エラー": {
			getGameErr: assert.AnError,
			wantErr:    assert.AnError,
		},
		"GetGameCreatorCustomJobsByGameIDがエラーの場合エラー": {
			executeGetCustomJobs: true,
			getCustomJobsErr:     assert.AnError,
			wantErr:              assert.AnError,
		},
		"同じ表示名が存在する場合ErrDuplicateCustomJobDisplayName": {
			executeGetCustomJobs: true,
			existingCustomJobs:   []*domain.GameCreatorCustomJob{duplicatedCustomJob},
			wantErr:              service.ErrDuplicateCustomJobDisplayName,
		},
		"CreateGameCreatorCustomJobsがエラーの場合エラー": {
			executeGetCustomJobs:        true,
			existingCustomJobs:          []*domain.GameCreatorCustomJob{},
			executeCreateGameCustomJobs: true,
			createGameCustomJobsErr:     assert.AnError,
			wantErr:                     assert.AnError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			gameRepository := mock.NewMockGameV2(ctrl)
			db := mock.NewMockDB(ctrl)
			gc := NewGameCreator(gameCreatorRepo, gameRepository, db, nil)

			gameRepository.EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeRecord).
				Return(nil, testCase.getGameErr)
			if testCase.executeGetCustomJobs {
				gameCreatorRepo.EXPECT().
					GetGameCreatorCustomJobsByGameID(gomock.Any(), gameID).
					Return(testCase.existingCustomJobs, testCase.getCustomJobsErr)
			}

			var createdCustomJob *domain.GameCreatorCustomJob
			if testCase.executeCreateGameCustomJobs {
				gameCreatorRepo.EXPECT().
					CreateGameCreatorCustomJobs(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, customJobs []*domain.GameCreatorCustomJob) error {
						if assert.Len(t, customJobs, 1) {
							createdCustomJob = customJobs[0]
						}
						return testCase.createGameCustomJobsErr
					})
			}

			customJob, err := gc.CreateGameCustomJob(t.Context(), gameID, displayName)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				assert.Nil(t, customJob)
				return
			}

			assert.NoError(t, err)
			assert.Same(t, createdCustomJob, customJob)
			assert.NotEqual(t, values.GameCreatorJobID{}, customJob.GetID())
			assert.Equal(t, gameID, customJob.GetGameID())
			assert.Equal(t, displayName, customJob.GetDisplayName())
			assert.False(t, customJob.GetCreatedAt().IsZero())
		})
	}
}

func TestDeleteGameCreator(t *testing.T) {
	t.Parallel()

	gameID := values.NewGameID()
	creatorID := values.NewGameCreatorID()
	creator := domain.NewGameCreator(
		creatorID,
		values.NewTrapMemberID(uuid.New()),
		gameID,
		values.NewTrapMemberName("creator"),
		time.Now(),
	)
	creatorOfAnotherGame := domain.NewGameCreator(
		creatorID,
		values.NewTrapMemberID(uuid.New()),
		values.NewGameID(),
		values.NewTrapMemberName("creator2"),
		time.Now(),
	)

	testCases := map[string]struct {
		creator                        *domain.GameCreator
		getGameCreatorByIDErr          error
		executeDeleteCustomJobs        bool
		deleteGameCreatorCustomJobsErr error
		executeDeletePresetJobs        bool
		deleteGameCreatorPresetJobsErr error
		executeDeleteGameCreator       bool
		deleteGameCreatorErr           error
		wantErr                        error
	}{
		"存在しないcreator IDの場合ErrInvalidGameCreatorID": {
			getGameCreatorByIDErr: repository.ErrRecordNotFound,
			wantErr:               service.ErrInvalidGameCreatorID,
		},
		"GetGameCreatorByIDがエラーの場合エラー": {
			getGameCreatorByIDErr: assert.AnError,
			wantErr:               assert.AnError,
		},
		"creatorが別のgameに属する場合ErrInvalidGameCreatorGamePair": {
			creator: creatorOfAnotherGame,
			wantErr: service.ErrInvalidGameCreatorGamePair,
		},
		"DeleteGameCreatorCustomJobsがエラーの場合エラー": {
			creator:                        creator,
			executeDeleteCustomJobs:        true,
			deleteGameCreatorCustomJobsErr: assert.AnError,
			wantErr:                        assert.AnError,
		},
		"DeleteGameCreatorPresetJobsがエラーの場合エラー": {
			creator:                        creator,
			executeDeleteCustomJobs:        true,
			executeDeletePresetJobs:        true,
			deleteGameCreatorPresetJobsErr: assert.AnError,
			wantErr:                        assert.AnError,
		},
		"DeleteGameCreatorで削除対象がない場合ErrInvalidGameCreatorGamePair": {
			creator:                  creator,
			executeDeleteCustomJobs:  true,
			executeDeletePresetJobs:  true,
			executeDeleteGameCreator: true,
			deleteGameCreatorErr:     repository.ErrNoRecordDeleted,
			wantErr:                  service.ErrInvalidGameCreatorGamePair,
		},
		"DeleteGameCreatorでエラーが発生した場合エラー": {
			creator:                  creator,
			executeDeleteCustomJobs:  true,
			executeDeletePresetJobs:  true,
			executeDeleteGameCreator: true,
			deleteGameCreatorErr:     assert.AnError,
			wantErr:                  assert.AnError,
		},
		"正常にcreatorを削除できる": {
			creator:                  creator,
			executeDeleteCustomJobs:  true,
			executeDeletePresetJobs:  true,
			executeDeleteGameCreator: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			db := mock.NewMockDB(ctrl)
			gc := NewGameCreator(gameCreatorRepo, nil, db, nil)

			gameCreatorRepo.EXPECT().
				GetGameCreatorByID(gomock.Any(), creatorID).
				Return(testCase.creator, testCase.getGameCreatorByIDErr)
			if testCase.executeDeleteCustomJobs {
				gameCreatorRepo.EXPECT().
					DeleteGameCreatorCustomJobs(gomock.Any(), creatorID).
					Return(testCase.deleteGameCreatorCustomJobsErr)
			}
			if testCase.executeDeletePresetJobs {
				gameCreatorRepo.EXPECT().
					DeleteGameCreatorPresetJobs(gomock.Any(), creatorID).
					Return(testCase.deleteGameCreatorPresetJobsErr)
			}
			if testCase.executeDeleteGameCreator {
				gameCreatorRepo.EXPECT().
					DeleteGameCreator(gomock.Any(), gameID, creatorID).
					Return(testCase.deleteGameCreatorErr)
			}

			err := gc.DeleteGameCreator(t.Context(), gameID, creatorID)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCreateGameCreator(t *testing.T) {
	t.Parallel()

	user1 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("user1"),
		values.TrapMemberStatusActive,
		false,
	)
	user2 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("user2"),
		values.TrapMemberStatusDeactivated,
		false,
	)
	users := []*service.UserInfo{
		user1,
		user2,
	}

	testCases := map[string]struct {
		gameID                    values.GameID
		userID                    values.TraPMemberID
		getAllUsersErr            error
		userName                  values.TraPMemberName
		executeGetGame            bool
		game                      *domain.Game
		GetGameErr                error
		executeCreateGameCreators bool
		CreateGameCreatorsErr     error
		wantErr                   error
	}{
		"getAllUsersがエラーなのでエラー": {
			gameID:         values.NewGameID(),
			userID:         user1.GetID(),
			getAllUsersErr: assert.AnError,
			wantErr:        assert.AnError,
		},
		"userが見つからないのでErrInvalidUserID": {
			gameID:  values.NewGameID(),
			userID:  values.NewTrapMemberID(uuid.New()),
			wantErr: service.ErrInvalidUserID,
		},
		"gameが見つからないのでErrInvalidGameID": {
			gameID:         values.NewGameID(),
			userID:         user1.GetID(),
			userName:       user1.GetName(),
			executeGetGame: true,
			GetGameErr:     repository.ErrRecordNotFound,
			wantErr:        service.ErrInvalidGameID,
		},
		"GetGameがエラーなのでエラー": {
			gameID:         values.NewGameID(),
			userID:         user1.GetID(),
			userName:       user1.GetName(),
			executeGetGame: true,
			GetGameErr:     assert.AnError,
			wantErr:        assert.AnError,
		},
		"CreateGameCreatorsがエラーなのでエラー": {
			gameID:                    values.NewGameID(),
			userID:                    user1.GetID(),
			userName:                  user1.GetName(),
			executeGetGame:            true,
			GetGameErr:                nil,
			executeCreateGameCreators: true,
			CreateGameCreatorsErr:     assert.AnError,
			wantErr:                   assert.AnError,
		},
		"正しく作成できる": {
			gameID:                    values.NewGameID(),
			userID:                    user1.GetID(),
			userName:                  user1.GetName(),
			executeGetGame:            true,
			executeCreateGameCreators: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameCreatorMock := mock.NewMockGameCreator(ctrl)
			gameMock := mock.NewMockGameV2(ctrl)
			dbMock := mock.NewMockDB(ctrl)
			authMock := mockAuth.NewMockUser(ctrl)
			cacheMock := mockCache.NewMockUser(ctrl)
			userMock := NewUser(authMock, cacheMock)
			gc := NewGameCreator(gameCreatorMock, gameMock, dbMock, userMock)

			if testCase.getAllUsersErr != nil {
				cacheMock.EXPECT().GetAllUsers(gomock.Any()).Return(nil, testCase.getAllUsersErr)
				authMock.EXPECT().GetAllUsers(gomock.Any(), gomock.Any()).Return(nil, testCase.getAllUsersErr)
			} else {
				cacheMock.EXPECT().GetAllUsers(gomock.Any()).Return(users, nil)
			}

			if testCase.executeGetGame {
				gameMock.EXPECT().
					GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
					Return(testCase.game, testCase.GetGameErr)
			}
			if testCase.executeCreateGameCreators {
				gameCreatorMock.EXPECT().
					CreateGameCreators(gomock.Any(), gomock.Cond(func(creators []*domain.GameCreator) bool {
						return len(creators) == 1 && creators[0].GetGameID() == testCase.gameID && creators[0].GetUserID() == testCase.userID && creators[0].GetUserName() == testCase.userName
					})).
					Return(testCase.CreateGameCreatorsErr)
			}

			session := domain.NewOIDCSession(values.NewOIDCAccessToken("access_token"), time.Now().Add(time.Hour))
			creator, err := gc.CreateGameCreator(t.Context(), session, testCase.gameID, testCase.userID)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.gameID, creator.GetGameID())
			assert.Equal(t, testCase.userID, creator.GetUserID())
			assert.Equal(t, testCase.userName, creator.GetUserName())
		})
	}
}

func TestSetGameCreatorJobs(t *testing.T) {
	t.Parallel()

	gameID := values.NewGameID()
	creatorID := values.NewGameCreatorID()
	creator := domain.NewGameCreator(
		creatorID,
		values.NewTrapMemberID(uuid.New()),
		gameID,
		values.NewTrapMemberName("creator"),
		time.Now(),
	)
	creatorOfAnotherGame := domain.NewGameCreator(
		creatorID,
		values.NewTrapMemberID(uuid.New()),
		values.NewGameID(),
		values.NewTrapMemberName("creator of another game"),
		time.Now(),
	)
	presetJob1 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Programmer"), time.Now())
	presetJob2 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Designer"), time.Now())
	customJob1 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Director"), gameID, time.Now())
	customJob2 := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Composer"), gameID, time.Now())

	testCases := map[string]struct {
		jobIDs                                  []values.GameCreatorJobID
		creator                                 *domain.GameCreator
		GetGameCreatorByIDErr                   error
		executeGetPresetJobs                    bool
		allPresetJobs                           []*domain.GameCreatorJob
		GetPresetJobsErr                        error
		executeGetGameCreatorCustomJobsByGameID bool
		allCustomJobs                           []*domain.GameCreatorCustomJob
		GetGameCreatorCustomJobsByGameIDErr     error
		executeGetCustomJobsByCreatorID         bool
		customJobs                              []*domain.GameCreatorCustomJob
		GetCustomJobsByCreatorIDErr             error
		executeGetPresetJobsByCreatorID         bool
		presetJobs                              []*domain.GameCreatorJob
		GetPresetJobsByCreatorIDErr             error
		executeDeletePresetJobsByCreatorID      bool
		deletingPresetJobIDs                    []values.GameCreatorJobID
		DeletePresetJobsByCreatorIDErr          error
		executeDeleteCustomJobsByCreatorID      bool
		deletingCustomJobIDs                    []values.GameCreatorJobID
		DeleteCustomJobsByCreatorIDErr          error
		executeUpsertPresetJobsRelations        bool
		UpsertPresetJobsRelationsErr            error
		executeUpsertCustomJobsRelations        bool
		UpsertCustomJobsRelationsErr            error
		wantErr                                 error
	}{
		"存在しないcreator IDの場合ErrInvalidGameCreatorID": {
			GetGameCreatorByIDErr: repository.ErrRecordNotFound,
			wantErr:               service.ErrInvalidGameCreatorID,
		},
		"creator取得でエラーの場合エラー": {
			GetGameCreatorByIDErr: assert.AnError,
			wantErr:               assert.AnError,
		},
		"creatorが別のgameに属する場合ErrInvalidGameCreatorGamePair": {
			creator: creatorOfAnotherGame,
			wantErr: service.ErrInvalidGameCreatorGamePair,
		},
		"jobIDが重複する場合はErrDuplicateGameCreatorJobID": {
			jobIDs:  []values.GameCreatorJobID{presetJob1.GetID(), presetJob1.GetID()},
			creator: creator,
			wantErr: service.ErrDuplicateGameCreatorJobID,
		},
		"preset job一覧取得でエラーの場合エラー": {
			creator:              creator,
			executeGetPresetJobs: true,
			GetPresetJobsErr:     assert.AnError,
			wantErr:              assert.AnError,
		},
		"creatorのcustom job取得でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobsByGameID: true,
			executeGetCustomJobsByCreatorID:         true,
			GetCustomJobsByCreatorIDErr:             assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"creatorのpreset job取得でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobsByGameID: true,
			allCustomJobs:                           []*domain.GameCreatorCustomJob{customJob1},
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			GetPresetJobsByCreatorIDErr:             assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"preset job削除でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobsByGameID: true,
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			presetJobs:                              []*domain.GameCreatorJob{presetJob1},
			executeDeletePresetJobsByCreatorID:      true,
			deletingPresetJobIDs:                    []values.GameCreatorJobID{presetJob1.GetID()},
			DeletePresetJobsByCreatorIDErr:          assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"custom job削除でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobsByGameID: true,
			allCustomJobs:                           []*domain.GameCreatorCustomJob{customJob1},
			executeGetCustomJobsByCreatorID:         true,
			customJobs:                              []*domain.GameCreatorCustomJob{customJob1},
			executeGetPresetJobsByCreatorID:         true,
			executeDeletePresetJobsByCreatorID:      true,
			executeDeleteCustomJobsByCreatorID:      true,
			deletingCustomJobIDs:                    []values.GameCreatorJobID{customJob1.GetID()},
			DeleteCustomJobsByCreatorIDErr:          assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"存在しないpreset job IDの場合ErrInvalidPresetJobID": {
			jobIDs:                                  []values.GameCreatorJobID{presetJob1.GetID()},
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1},
			executeGetGameCreatorCustomJobsByGameID: true,
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			executeDeletePresetJobsByCreatorID:      true,
			executeDeleteCustomJobsByCreatorID:      true,
			executeUpsertPresetJobsRelations:        true,
			UpsertPresetJobsRelationsErr:            repository.ErrForeignKeyViolated,
			wantErr:                                 service.ErrInvalidPresetJobID,
		},
		"preset job relation更新でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			executeGetGameCreatorCustomJobsByGameID: true,
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			executeDeletePresetJobsByCreatorID:      true,
			executeDeleteCustomJobsByCreatorID:      true,
			executeUpsertPresetJobsRelations:        true,
			UpsertPresetJobsRelationsErr:            assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"gameに紐づかないcustom job IDの場合ErrInvalidGameAndCustomJobPair": {
			jobIDs:                                  []values.GameCreatorJobID{customJob1.GetID()},
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			executeGetGameCreatorCustomJobsByGameID: true,
			allCustomJobs:                           []*domain.GameCreatorCustomJob{customJob1},
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			executeDeletePresetJobsByCreatorID:      true,
			executeDeleteCustomJobsByCreatorID:      true,
			executeUpsertPresetJobsRelations:        true,
			executeUpsertCustomJobsRelations:        true,
			UpsertCustomJobsRelationsErr:            repository.ErrForeignKeyViolated,
			wantErr:                                 service.ErrInvalidGameAndCustomJobPair,
		},
		"custom job relation更新でエラーの場合エラー": {
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			executeGetGameCreatorCustomJobsByGameID: true,
			executeGetCustomJobsByCreatorID:         true,
			executeGetPresetJobsByCreatorID:         true,
			executeDeletePresetJobsByCreatorID:      true,
			executeDeleteCustomJobsByCreatorID:      true,
			executeUpsertPresetJobsRelations:        true,
			executeUpsertCustomJobsRelations:        true,
			UpsertCustomJobsRelationsErr:            assert.AnError,
			wantErr:                                 assert.AnError,
		},
		"preset jobとcustom jobを置き換えられる": {
			jobIDs:                                  []values.GameCreatorJobID{presetJob1.GetID(), customJob1.GetID()},
			creator:                                 creator,
			executeGetPresetJobs:                    true,
			allPresetJobs:                           []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetGameCreatorCustomJobsByGameID: true,
			allCustomJobs:                           []*domain.GameCreatorCustomJob{customJob1, customJob2},
			executeGetCustomJobsByCreatorID:         true,
			customJobs:                              []*domain.GameCreatorCustomJob{customJob2},
			executeGetPresetJobsByCreatorID:         true,
			presetJobs:                              []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeDeletePresetJobsByCreatorID:      true,
			deletingPresetJobIDs:                    []values.GameCreatorJobID{presetJob2.GetID()},
			executeDeleteCustomJobsByCreatorID:      true,
			deletingCustomJobIDs:                    []values.GameCreatorJobID{customJob2.GetID()},
			executeUpsertPresetJobsRelations:        true,
			executeUpsertCustomJobsRelations:        true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			db := mock.NewMockDB(ctrl)
			gc := NewGameCreator(gameCreatorRepo, nil, db, nil)

			gameCreatorRepo.EXPECT().
				GetGameCreatorByID(gomock.Any(), creatorID).
				Return(testCase.creator, testCase.GetGameCreatorByIDErr)
			if testCase.executeGetPresetJobs {
				gameCreatorRepo.EXPECT().
					GetGameCreatorPresetJobs(gomock.Any()).
					Return(testCase.allPresetJobs, testCase.GetPresetJobsErr)
			}
			if testCase.executeGetGameCreatorCustomJobsByGameID {
				gameCreatorRepo.EXPECT().
					GetGameCreatorCustomJobsByGameID(gomock.Any(), gameID).
					Return(testCase.allCustomJobs, testCase.GetGameCreatorCustomJobsByGameIDErr)
			}

			if testCase.executeGetCustomJobsByCreatorID {
				gameCreatorRepo.EXPECT().
					GetGameCreatorCustomJobsByCreatorID(gomock.Any(), creatorID).
					Return(testCase.customJobs, testCase.GetCustomJobsByCreatorIDErr)
			}
			if testCase.executeGetPresetJobsByCreatorID {
				gameCreatorRepo.EXPECT().
					GetGameCreatorPresetJobsByCreatorID(gomock.Any(), creatorID).
					Return(testCase.presetJobs, testCase.GetPresetJobsByCreatorIDErr)
			}
			if testCase.executeDeletePresetJobsByCreatorID {
				deletingPresetJobIDs := testCase.deletingPresetJobIDs
				if deletingPresetJobIDs == nil {
					deletingPresetJobIDs = []values.GameCreatorJobID{}
				}
				gameCreatorRepo.EXPECT().
					DeleteGameCreatorPresetJobsByCreatorID(gomock.Any(), creatorID, deletingPresetJobIDs).
					Return(testCase.DeletePresetJobsByCreatorIDErr)
			}
			if testCase.executeDeleteCustomJobsByCreatorID {
				deletingCustomJobIDs := testCase.deletingCustomJobIDs
				if deletingCustomJobIDs == nil {
					deletingCustomJobIDs = []values.GameCreatorJobID{}
				}
				gameCreatorRepo.EXPECT().
					DeleteGameCreatorCustomJobsByCreatorID(gomock.Any(), creatorID, deletingCustomJobIDs).
					Return(testCase.DeleteCustomJobsByCreatorIDErr)
			}

			presetJobIDs := make([]values.GameCreatorJobID, 0, len(testCase.jobIDs))
			customJobIDs := make([]values.GameCreatorJobID, 0, len(testCase.jobIDs))
			for _, jobID := range testCase.jobIDs {
				isPreset := false
				for _, presetJob := range testCase.allPresetJobs {
					if jobID == presetJob.GetID() {
						isPreset = true
						break
					}
				}
				if isPreset {
					presetJobIDs = append(presetJobIDs, jobID)
				} else {
					customJobIDs = append(customJobIDs, jobID)
				}
			}
			if testCase.executeUpsertPresetJobsRelations {
				gameCreatorRepo.EXPECT().
					UpsertGameCreatorPresetJobsRelations(gomock.Any(), map[values.GameCreatorID][]values.GameCreatorJobID{creatorID: presetJobIDs}).
					Return(testCase.UpsertPresetJobsRelationsErr)
			}
			if testCase.executeUpsertCustomJobsRelations {
				gameCreatorRepo.EXPECT().
					UpsertGameCreatorCustomJobsRelations(gomock.Any(), map[values.GameCreatorID][]values.GameCreatorJobID{creatorID: customJobIDs}).
					Return(testCase.UpsertCustomJobsRelationsErr)
			}

			err := gc.SetGameCreatorJobs(t.Context(), gameID, creatorID, testCase.jobIDs)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestEditGameCreators(t *testing.T) {
	t.Parallel()

	gameID := values.NewGameID()
	user1 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("user1"),
		values.TrapMemberStatusActive,
		false,
	)
	user2 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("user2"),
		values.TrapMemberStatusActive,
		false,
	)
	invalidUserID := values.NewTrapMemberID(uuid.New())

	presetJob1 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Programmer"), time.Now())
	presetJob2 := domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Designer"), time.Now())
	existingCustomJob := domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("Existing Custom Job"), gameID, time.Now())

	newCustomJobName1 := values.NewGameCreatorJobDisplayName("New Custom Job 1")
	newCustomJobName2 := values.NewGameCreatorJobDisplayName("New Custom Job 2")

	existingCreator := domain.NewGameCreator(values.NewGameCreatorID(), user1.GetID(), gameID, user1.GetName(), time.Now())

	testCases := map[string]struct {
		gameID                              values.GameID
		inputs                              []*service.EditGameCreatorJobInput
		getGameErr                          error
		executeGetGameCreatorPresetJobs     bool
		presetJobs                          []*domain.GameCreatorJob
		getGameCreatorPresetJobsErr         error
		executeGetAllUsers                  bool
		cacheUsers                          []*service.UserInfo
		cacheGetAllUsersErr                 error
		authUsers                           []*service.UserInfo
		authGetAllUsersErr                  error
		executeGetGameCreatorCustomJobsByID bool
		existingCustomJobs                  []*domain.GameCreatorCustomJob
		getGameCreatorCustomJobsByIDErr     error
		executeGetCreatorsByUserIDs         bool
		existingCreators                    []*domain.GameCreator
		getCreatorsByUserIDsErr             error
		executeCreateGameCreatorCustomJobs  bool
		createGameCreatorCustomJobsErr      error
		executeCreateGameCreators           bool
		createGameCreatorsErr               error
		executeUpsertPresetJobsRelations    bool
		upsertPresetJobsRelationsErr        error
		executeUpsertCustomJobsRelations    bool
		upsertCustomJobsRelationsErr        error
		wantErr                             error
	}{
		"ゲームが存在しない場合ErrInvalidGameID": {
			gameID:     gameID,
			inputs:     []*service.EditGameCreatorJobInput{},
			getGameErr: repository.ErrRecordNotFound,
			wantErr:    service.ErrInvalidGameID,
		},
		"ゲームの取得でエラーが発生した場合エラー": {
			gameID:     gameID,
			inputs:     []*service.EditGameCreatorJobInput{},
			getGameErr: assert.AnError,
			wantErr:    assert.AnError,
		},
		"preset job取得でエラーなのでエラー": {
			gameID:                          gameID,
			inputs:                          []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			executeGetGameCreatorPresetJobs: true,
			getGameCreatorPresetJobsErr:     assert.AnError,
			wantErr:                         assert.AnError,
		},
		"all user取得でエラーなのでエラー": {
			gameID:                          gameID,
			inputs:                          []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			executeGetGameCreatorPresetJobs: true,
			presetJobs:                      []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:              true,
			cacheGetAllUsersErr:             cache.ErrCacheMiss,
			authGetAllUsersErr:              assert.AnError,
			wantErr:                         assert.AnError,
		},
		"all usersに存在しないユーザーが入力された場合ErrInvalidUserID": {
			gameID:                          gameID,
			inputs:                          []*service.EditGameCreatorJobInput{{UserID: invalidUserID}},
			executeGetGameCreatorPresetJobs: true,
			presetJobs:                      []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:              true,
			cacheGetAllUsersErr:             cache.ErrCacheMiss,
			authUsers:                       []*service.UserInfo{user1, user2},
			wantErr:                         service.ErrInvalidUserID,
		},
		"入力内に重複したユーザーIDがある場合ErrDuplicateUserID": {
			gameID:                          gameID,
			inputs:                          []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}, {UserID: user1.GetID()}},
			executeGetGameCreatorPresetJobs: true,
			presetJobs:                      []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:              true,
			cacheGetAllUsersErr:             cache.ErrCacheMiss,
			authUsers:                       []*service.UserInfo{user1, user2},
			wantErr:                         service.ErrDuplicateUserID,
		},
		"既存custom job名と重複したnew custom job名の場合ErrDuplicateCustomJobDisplayName": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID(), NewCustomJobNames: []values.GameCreatorJobDisplayName{existingCustomJob.GetDisplayName()}}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			wantErr:                             service.ErrDuplicateCustomJobDisplayName,
		},
		"存在しない job id を指定した場合、ErrInvalidGameCreatorJobID": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID(), Jobs: []values.GameCreatorJobID{values.NewGameCreatorJobID()}}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			wantErr:                             service.ErrInvalidGameCreatorJobID,
		},
		"同一ユーザー入力内に重複したjob idがある場合ErrDuplicateGameCreatorJobID": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID: user1.GetID(),
				Jobs:   []values.GameCreatorJobID{presetJob1.GetID(), presetJob1.GetID()},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			wantErr:                             service.ErrDuplicateGameCreatorJobID,
		},
		"既存のカスタムジョブの取得でエラーなのでエラー": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			getGameCreatorCustomJobsByIDErr:     assert.AnError,
			wantErr:                             assert.AnError,
		},
		"creator取得でエラーなのでエラー": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetCreatorsByUserIDs:         true,
			getCreatorsByUserIDsErr:             assert.AnError,
			wantErr:                             assert.AnError,
		},
		"create custom jobsでエラーなのでエラー": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID:            user1.GetID(),
				NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName1},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{},
			executeCreateGameCreatorCustomJobs:  true,
			createGameCreatorCustomJobsErr:      assert.AnError,
			wantErr:                             assert.AnError,
		},
		"create creatorsでエラーなのでエラー": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID:            user1.GetID(),
				NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName1},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			executeGetAllUsers:                  true,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{},
			executeCreateGameCreatorCustomJobs:  true,
			executeCreateGameCreators:           true,
			createGameCreatorsErr:               assert.AnError,
			wantErr:                             assert.AnError,
		},
		"upsert preset jobs relationsでエラーなのでエラー": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID: user1.GetID(),
				Jobs:   []values.GameCreatorJobID{presetJob1.GetID(), existingCustomJob.GetID()},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{existingCreator},
			executeCreateGameCreatorCustomJobs:  true,
			executeCreateGameCreators:           true,
			executeUpsertPresetJobsRelations:    true,
			upsertPresetJobsRelationsErr:        assert.AnError,
			wantErr:                             assert.AnError,
		},
		"upsert custom jobs relationsでエラーなのでエラー": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID:            user1.GetID(),
				Jobs:              []values.GameCreatorJobID{presetJob1.GetID(), existingCustomJob.GetID()},
				NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName1},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{existingCreator},
			executeCreateGameCreatorCustomJobs:  true,
			executeCreateGameCreators:           true,
			executeUpsertPresetJobsRelations:    true,
			executeUpsertCustomJobsRelations:    true,
			upsertCustomJobsRelationsErr:        assert.AnError,
			wantErr:                             assert.AnError,
		},
		"user cache hitの正常系": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{
				{
					UserID:            user1.GetID(),
					Jobs:              []values.GameCreatorJobID{presetJob1.GetID(), existingCustomJob.GetID()},
					NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName1},
				},
				{
					UserID:            user2.GetID(),
					Jobs:              []values.GameCreatorJobID{presetJob2.GetID()},
					NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName2},
				},
			},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheUsers:                          []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{existingCreator},
			executeCreateGameCreatorCustomJobs:  true,
			executeCreateGameCreators:           true,
			executeUpsertPresetJobsRelations:    true,
			executeUpsertCustomJobsRelations:    true,
			wantErr:                             nil,
		},
		"user cache missの正常系": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID:            user1.GetID(),
				Jobs:              []values.GameCreatorJobID{presetJob1.GetID(), existingCustomJob.GetID()},
				NewCustomJobNames: []values.GameCreatorJobDisplayName{newCustomJobName1},
			}},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
			executeGetAllUsers:                  true,
			cacheGetAllUsersErr:                 cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetCreatorsByUserIDs:         true,
			existingCreators:                    []*domain.GameCreator{existingCreator},
			executeCreateGameCreatorCustomJobs:  true,
			executeCreateGameCreators:           true,
			executeUpsertPresetJobsRelations:    true,
			executeUpsertCustomJobsRelations:    true,
			wantErr:                             nil,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockGameCreatorRepo := mock.NewMockGameCreator(ctrl)
			mockGameRepository := mock.NewMockGameV2(ctrl)
			mockDB := mock.NewMockDB(ctrl)
			mockUserCache := mockCache.NewMockUser(ctrl)
			mockUserAuth := mockAuth.NewMockUser(ctrl)
			mockUser := NewUser(mockUserAuth, mockUserCache)

			gc := NewGameCreator(mockGameCreatorRepo, mockGameRepository, mockDB, mockUser)

			sess := domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))
			mockGameRepository.EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetAllUsers {
				cacheUsers := testCase.cacheUsers
				if cacheUsers == nil {
					cacheUsers = []*service.UserInfo{}
				}
				mockUserCache.EXPECT().
					GetAllUsers(gomock.Any()).
					Return(cacheUsers, testCase.cacheGetAllUsersErr)
				if testCase.cacheGetAllUsersErr != nil {
					authUsers := testCase.authUsers
					if authUsers == nil {
						authUsers = []*service.UserInfo{}
					}
					mockUserAuth.EXPECT().
						GetAllUsers(gomock.Any(), sess).
						Return(authUsers, testCase.authGetAllUsersErr)
					if testCase.authGetAllUsersErr == nil {
						mockUserCache.EXPECT().
							SetAllUsers(gomock.Any(), authUsers).
							Return(nil)
					}
				}
			}
			if testCase.executeGetGameCreatorCustomJobsByID {
				mockGameCreatorRepo.EXPECT().
					GetGameCreatorCustomJobsByGameID(gomock.Any(), testCase.gameID).
					Return(testCase.existingCustomJobs, testCase.getGameCreatorCustomJobsByIDErr)
			}
			if testCase.executeGetGameCreatorPresetJobs {
				mockGameCreatorRepo.EXPECT().
					GetGameCreatorPresetJobs(gomock.Any()).
					Return(testCase.presetJobs, testCase.getGameCreatorPresetJobsErr)
			}
			if testCase.executeGetCreatorsByUserIDs {
				userIDs := make([]values.TraPMemberID, len(testCase.inputs))
				for i, input := range testCase.inputs {
					userIDs[i] = input.UserID
				}
				mockGameCreatorRepo.EXPECT().
					GetGameCreatorsByUserIDs(gomock.Any(), testCase.gameID, userIDs).
					Return(testCase.existingCreators, testCase.getCreatorsByUserIDsErr)
			}
			if testCase.executeCreateGameCreatorCustomJobs {
				mockGameCreatorRepo.EXPECT().
					CreateGameCreatorCustomJobs(gomock.Any(), gomock.Any()).
					Return(testCase.createGameCreatorCustomJobsErr)
			}
			if testCase.executeCreateGameCreators {
				mockGameCreatorRepo.EXPECT().
					CreateGameCreators(gomock.Any(), gomock.Any()).
					Return(testCase.createGameCreatorsErr)
			}
			if testCase.executeUpsertPresetJobsRelations {
				mockGameCreatorRepo.EXPECT().
					UpsertGameCreatorPresetJobsRelations(gomock.Any(), gomock.Any()).
					Return(testCase.upsertPresetJobsRelationsErr)
			}
			if testCase.executeUpsertCustomJobsRelations {
				mockGameCreatorRepo.EXPECT().
					UpsertGameCreatorCustomJobsRelations(gomock.Any(), gomock.Any()).
					Return(testCase.upsertCustomJobsRelationsErr)
			}

			err := gc.EditGameCreators(t.Context(), sess, testCase.gameID, testCase.inputs)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}
