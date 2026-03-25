package v2

import (
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
		cacheUsers                          []*service.UserInfo
		cacheGetActiveUsersErr              error
		authUsers                           []*service.UserInfo
		authGetActiveUsersErr               error
		executeGetGameCreatorCustomJobsByID bool
		existingCustomJobs                  []*domain.GameCreatorCustomJob
		getGameCreatorCustomJobsByIDErr     error
		executeGetGameCreatorPresetJobs     bool
		presetJobs                          []*domain.GameCreatorJob
		getGameCreatorPresetJobsErr         error
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
		"active user取得でエラーなのでエラー": {
			gameID:                 gameID,
			inputs:                 []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			cacheGetActiveUsersErr: cache.ErrCacheMiss,
			authGetActiveUsersErr:  assert.AnError,
			wantErr:                assert.AnError,
		},
		"active usersに存在しないユーザーが入力された場合ErrInvalidUserID": {
			gameID:                 gameID,
			inputs:                 []*service.EditGameCreatorJobInput{{UserID: invalidUserID}},
			cacheGetActiveUsersErr: cache.ErrCacheMiss,
			authUsers:              []*service.UserInfo{user1, user2},
			wantErr:                service.ErrInvalidUserID,
		},
		"既存custom job名と重複したnew custom job名の場合ErrDuplicateCustomJobDisplayName": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID(), NewCustomJobNames: []values.GameCreatorJobDisplayName{existingCustomJob.GetDisplayName()}}},
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			wantErr:                             service.ErrDuplicateCustomJobDisplayName,
		},
		"同一ユーザー入力内に重複したjob idがある場合ErrDuplicateGameCreatorJobID": {
			gameID: gameID,
			inputs: []*service.EditGameCreatorJobInput{{
				UserID: user1.GetID(),
				Jobs:   []values.GameCreatorJobID{presetJob1.GetID(), presetJob1.GetID()},
			}},
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			wantErr:                             service.ErrDuplicateGameCreatorJobID,
		},
		"既存のカスタムジョブの取得でエラーなのでエラー": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			getGameCreatorCustomJobsByIDErr:     assert.AnError,
			wantErr:                             assert.AnError,
		},
		"preset job取得でエラーなのでエラー": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetGameCreatorPresetJobs:     true,
			getGameCreatorPresetJobsErr:         assert.AnError,
			wantErr:                             assert.AnError,
		},
		"creator取得でエラーなのでエラー": {
			gameID:                              gameID,
			inputs:                              []*service.EditGameCreatorJobInput{{UserID: user1.GetID()}},
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheUsers:                          []*service.UserInfo{user1, user2},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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
			cacheGetActiveUsersErr:              cache.ErrCacheMiss,
			authUsers:                           []*service.UserInfo{user1},
			executeGetGameCreatorCustomJobsByID: true,
			existingCustomJobs:                  []*domain.GameCreatorCustomJob{existingCustomJob},
			executeGetGameCreatorPresetJobs:     true,
			presetJobs:                          []*domain.GameCreatorJob{presetJob1, presetJob2},
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

			if testCase.getGameErr == nil {
				cacheUsers := testCase.cacheUsers
				if cacheUsers == nil {
					cacheUsers = []*service.UserInfo{}
				}
				mockUserCache.EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(cacheUsers, testCase.cacheGetActiveUsersErr)
				if testCase.cacheGetActiveUsersErr != nil {
					authUsers := testCase.authUsers
					if authUsers == nil {
						authUsers = []*service.UserInfo{}
					}
					mockUserAuth.EXPECT().
						GetActiveUsers(gomock.Any(), sess).
						Return(authUsers, testCase.authGetActiveUsersErr)
					if testCase.authGetActiveUsersErr == nil {
						mockUserCache.EXPECT().
							SetActiveUsers(gomock.Any(), authUsers).
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
				mockGameCreatorRepo.EXPECT().
					GetCreatorsByUserIDs(gomock.Any(), gomock.Any()).
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
