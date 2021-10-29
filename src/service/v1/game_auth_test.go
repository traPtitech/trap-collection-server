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
