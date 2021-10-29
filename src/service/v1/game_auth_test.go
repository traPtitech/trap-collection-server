package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestAddGameCollaborators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameAuthService := NewGameAuth(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		userUtils,
	)

	type test struct {
		description                  string
		gameID                       values.GameID
		userIDs                      []values.TraPMemberID
		GetGameErr                   error
		users                        []*service.UserInfo
		executeGetAllActiveUser      bool
		isGetAllActiveUserErr        bool
		executeAddGameManagementRole bool
		AddGameManagementRolesErr    error
		isErr                        bool
		err                          error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(userID1, "mazrean", values.TrapMemberStatusActive),
			},
			executeAddGameManagementRole: true,
		},
		{
			description: "GetGameがRecordNotFoundなのでエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			GetGameErr: repository.ErrRecordNotFound,
			isErr:      true,
			err:        service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラー(RecordNotFound以外)なのでエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			GetGameErr: errors.New("error"),
			isErr:      true,
		},
		{
			description: "getAllActiveUserがエラーなのでエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			executeGetAllActiveUser: true,
			isGetAllActiveUserErr:   true,
			isErr:                   true,
		},
		{
			description: "userが存在しないのでエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			executeGetAllActiveUser: true,
			users:                   []*service.UserInfo{},
			isErr:                   true,
			err:                     service.ErrInvalidUserID,
		},
		{
			description: "userIDが複数でも問題なし",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
				userID2,
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(userID1, "mazrean", values.TrapMemberStatusActive),
				service.NewUserInfo(userID2, "mazrean", values.TrapMemberStatusActive),
			},
			executeAddGameManagementRole: true,
		},
		{
			description: "userIDが複数で一部が存在しなくてもエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
				userID2,
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(userID1, "mazrean", values.TrapMemberStatusActive),
			},
			isErr: true,
			err:   service.ErrInvalidUserID,
		},
		{
			description: "AddGameManagementRoleがエラーなのでエラー",
			gameID:      values.NewGameID(),
			userIDs: []values.TraPMemberID{
				userID1,
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(userID1, "mazrean", values.TrapMemberStatusActive),
			},
			executeAddGameManagementRole: true,
			AddGameManagementRolesErr:    errors.New("error"),
			isErr:                        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := &domain.OIDCSession{}

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetAllActiveUser {
				if testCase.isGetAllActiveUserErr {
					mockUserCache.
						EXPECT().
						GetAllActiveUsers(gomock.Any()).
						Return(nil, cache.ErrCacheMiss)
					mockUserAuth.
						EXPECT().
						GetAllActiveUsers(gomock.Any(), session).
						Return(nil, errors.New("error"))
				} else {
					mockUserCache.
						EXPECT().
						GetAllActiveUsers(gomock.Any()).
						Return(testCase.users, nil)
				}
			}

			if testCase.executeAddGameManagementRole {
				mockGameManagementRoleRepository.
					EXPECT().
					AddGameManagementRoles(
						gomock.Any(),
						testCase.gameID,
						testCase.userIDs,
						values.GameManagementRoleCollaborator,
					).Return(testCase.AddGameManagementRolesErr)
			}

			err := gameAuthService.AddGameCollaborators(ctx, session, testCase.gameID, testCase.userIDs)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}
		})
	}
}

func TestUpdateGameManagementRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameAuthService := NewGameAuth(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		userUtils,
	)

	type test struct {
		description                     string
		gameID                          values.GameID
		userID                          values.TraPMemberID
		role                            values.GameManagementRole
		nowRole                         values.GameManagementRole
		GetGameManagementRoleErr        error
		executeUpdateGameManagementRole bool
		UpdateGameManagementRoleErr     error
		isErr                           bool
		err                             error
	}

	testCases := []test{
		{
			description:                     "特に問題ないのでエラーなし",
			gameID:                          values.NewGameID(),
			userID:                          values.NewTrapMemberID(uuid.New()),
			role:                            values.GameManagementRoleCollaborator,
			nowRole:                         values.GameManagementRoleAdministrator,
			executeUpdateGameManagementRole: true,
		},
		{
			description:              "GetGameManagementRoleがErrRecordNotFoundなのでエラー",
			gameID:                   values.NewGameID(),
			userID:                   values.NewTrapMemberID(uuid.New()),
			role:                     values.GameManagementRoleCollaborator,
			nowRole:                  values.GameManagementRoleAdministrator,
			GetGameManagementRoleErr: repository.ErrRecordNotFound,
			isErr:                    true,
			err:                      service.ErrInvalidRole,
		},
		{
			description:              "GetGameManagementRoleがエラー(ErrRecordNotFound)なのでエラー",
			gameID:                   values.NewGameID(),
			userID:                   values.NewTrapMemberID(uuid.New()),
			role:                     values.GameManagementRoleCollaborator,
			nowRole:                  values.GameManagementRoleAdministrator,
			GetGameManagementRoleErr: errors.New("error"),
			isErr:                    true,
		},
		{
			description: "roleが既に指定通りなのでエラー",
			gameID:      values.NewGameID(),
			userID:      values.NewTrapMemberID(uuid.New()),
			role:        values.GameManagementRoleCollaborator,
			nowRole:     values.GameManagementRoleCollaborator,
			isErr:       true,
			err:         service.ErrNoGameManagementRoleUpdated,
		},
		{
			description:                     "UpdateGameManagementRoleがエラーなのでエラー",
			gameID:                          values.NewGameID(),
			userID:                          values.NewTrapMemberID(uuid.New()),
			role:                            values.GameManagementRoleCollaborator,
			nowRole:                         values.GameManagementRoleAdministrator,
			executeUpdateGameManagementRole: true,
			UpdateGameManagementRoleErr:     errors.New("error"),
			isErr:                           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameManagementRoleRepository.
				EXPECT().
				GetGameManagementRole(gomock.Any(), testCase.gameID, testCase.userID, repository.LockTypeRecord).
				Return(testCase.nowRole, testCase.GetGameManagementRoleErr)

			if testCase.executeUpdateGameManagementRole {
				mockGameManagementRoleRepository.
					EXPECT().
					UpdateGameManagementRole(
						gomock.Any(),
						testCase.gameID,
						testCase.userID,
						testCase.role,
					).Return(testCase.UpdateGameManagementRoleErr)
			}

			err := gameAuthService.UpdateGameManagementRole(
				ctx,
				testCase.gameID,
				testCase.userID,
				testCase.role,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveGameCollaborator(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameAuthService := NewGameAuth(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		userUtils,
	)

	type test struct {
		description                     string
		gameID                          values.GameID
		userID                          values.TraPMemberID
		nowRole                         values.GameManagementRole
		GetGameManagementRoleErr        error
		executeRemoveGameManagementRole bool
		RemoveGameManagementRoleErr     error
		isErr                           bool
		err                             error
	}

	testCases := []test{
		{
			description:                     "特に問題ないのでエラーなし",
			gameID:                          values.NewGameID(),
			userID:                          values.NewTrapMemberID(uuid.New()),
			nowRole:                         values.GameManagementRoleCollaborator,
			executeRemoveGameManagementRole: true,
		},
		{
			description:              "GetGameManagementRoleがErrRecordNotFoundなのでエラー",
			gameID:                   values.NewGameID(),
			userID:                   values.NewTrapMemberID(uuid.New()),
			nowRole:                  values.GameManagementRoleCollaborator,
			GetGameManagementRoleErr: repository.ErrRecordNotFound,
			isErr:                    true,
			err:                      service.ErrInvalidRole,
		},
		{
			description:              "GetGameManagementRoleがエラー(ErrRecordNotFound以外)なのでエラー",
			gameID:                   values.NewGameID(),
			userID:                   values.NewTrapMemberID(uuid.New()),
			nowRole:                  values.GameManagementRoleCollaborator,
			GetGameManagementRoleErr: errors.New("error"),
			isErr:                    true,
		},
		{
			description: "roleがCollaboratorでないのでエラー",
			gameID:      values.NewGameID(),
			userID:      values.NewTrapMemberID(uuid.New()),
			nowRole:     values.GameManagementRoleAdministrator,
			isErr:       true,
			err:         service.ErrInvalidRole,
		},
		{
			description:                     "RemoveGameCollaboratorがエラーなのでエラー",
			gameID:                          values.NewGameID(),
			userID:                          values.NewTrapMemberID(uuid.New()),
			nowRole:                         values.GameManagementRoleCollaborator,
			executeRemoveGameManagementRole: true,
			RemoveGameManagementRoleErr:     errors.New("error"),
			isErr:                           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameManagementRoleRepository.
				EXPECT().
				GetGameManagementRole(gomock.Any(), testCase.gameID, testCase.userID, repository.LockTypeRecord).
				Return(testCase.nowRole, testCase.GetGameManagementRoleErr)

			if testCase.executeRemoveGameManagementRole {
				mockGameManagementRoleRepository.
					EXPECT().
					RemoveGameManagementRole(
						gomock.Any(),
						testCase.gameID,
						testCase.userID,
					).Return(testCase.RemoveGameManagementRoleErr)
			}

			err := gameAuthService.RemoveGameCollaborator(
				ctx,
				testCase.gameID,
				testCase.userID,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGameManagers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameAuthService := NewGameAuth(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		userUtils,
	)

	type test struct {
		description                    string
		gameID                         values.GameID
		GetGameErr                     error
		executeGetGameManagersByGameID bool
		userIDAndRoles                 []*repository.UserIDAndManagementRole
		GetGameManagersByGameIDErr     error
		executeGetAllActiveUser        bool
		users                          []*service.UserInfo
		isGetAllActiveUserErr          bool
		gameManagers                   []*service.GameManager
		isErr                          bool
		err                            error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description:                    "特に問題ないのでエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					userID1,
					"mazrean",
					values.TrapMemberStatusActive,
				),
			},
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでエラー",
			gameID:      values.NewGameID(),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                    "GetGameManagersByGameIDがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			GetGameManagersByGameIDErr:     errors.New("error"),
			isErr:                          true,
		},
		{
			description:                    "GetAllActiveUserがエラーなのでエラー",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetAllActiveUser: true,
			isGetAllActiveUserErr:   true,
			isErr:                   true,
		},
		{
			description:                    "roleのないユーザーが存在いてもエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					userID1,
					"mazrean",
					values.TrapMemberStatusActive,
				),
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					"mazrean1",
					values.TrapMemberStatusActive,
				),
			},
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
			},
		},
		{
			// 凍結されたユーザーがroleに含まれている可能性があるため
			description:                    "存在しないユーザーが存在してもエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
				{
					UserID: values.NewTrapMemberID(uuid.New()),
					Role:   values.GameManagementRoleCollaborator,
				},
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					userID1,
					"mazrean",
					values.TrapMemberStatusActive,
				),
			},
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
			},
		},
		{
			// 実際にはまずあり得ないが、念の為確認
			description:                    "roleを持つユーザーが存在しなくてもエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles:                 []*repository.UserIDAndManagementRole{},
			executeGetAllActiveUser:        true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					userID1,
					"mazrean",
					values.TrapMemberStatusActive,
				),
			},
			gameManagers: []*service.GameManager{},
		},
		{
			// 実際にはまずあり得ないが、念の為確認
			description:                    "ユーザーが存在しなくてもエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetAllActiveUser: true,
			users:                   []*service.UserInfo{},
			gameManagers:            []*service.GameManager{},
		},
		{
			description:                    "Managerが複数いてもエラーなし",
			gameID:                         values.NewGameID(),
			executeGetGameManagersByGameID: true,
			userIDAndRoles: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
				{
					UserID: userID2,
					Role:   values.GameManagementRoleCollaborator,
				},
			},
			executeGetAllActiveUser: true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					userID1,
					"mazrean",
					values.TrapMemberStatusActive,
				),
				service.NewUserInfo(
					userID2,
					"mazrean1",
					values.TrapMemberStatusActive,
				),
			},
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
				{
					UserID:     userID2,
					UserName:   "mazrean1",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleCollaborator,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := &domain.OIDCSession{}

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetGameManagersByGameID {
				mockGameManagementRoleRepository.
					EXPECT().
					GetGameManagersByGameID(gomock.Any(), testCase.gameID).
					Return(testCase.userIDAndRoles, testCase.GetGameManagersByGameIDErr)
			}

			if testCase.executeGetAllActiveUser {
				if testCase.isGetAllActiveUserErr {
					mockUserCache.
						EXPECT().
						GetAllActiveUsers(gomock.Any()).
						Return(nil, cache.ErrCacheMiss)
					mockUserAuth.
						EXPECT().
						GetAllActiveUsers(gomock.Any(), session).
						Return(nil, errors.New("error"))
				} else {
					mockUserCache.
						EXPECT().
						GetAllActiveUsers(gomock.Any()).
						Return(testCase.users, nil)
				}
			}

			gameManagers, err := gameAuthService.GetGameManagers(ctx, session, testCase.gameID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Len(t, gameManagers, len(testCase.gameManagers))
			for i, gameManager := range gameManagers {
				assert.Equal(t, testCase.gameManagers[i].UserID, gameManager.UserID)
				assert.Equal(t, testCase.gameManagers[i].UserName, gameManager.UserName)
				assert.Equal(t, testCase.gameManagers[i].UserStatus, gameManager.UserStatus)
				assert.Equal(t, testCase.gameManagers[i].Role, gameManager.Role)
			}
		})
	}
}
