package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestAddAdmin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockAdminAuthRepository := mockRepository.NewMockAdminAuthV2(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	user := NewUser(mockUserAuth, mockUserCache)

	adminAuthService := NewAdminAuth(mockDB, mockAdminAuthRepository, user)

	type test struct {
		description       string
		authSession       *domain.OIDCSession
		getActiveUsersErr error
		userID            values.TraPMemberID
		executeGetAdmins  bool
		GetAdminsErr      error
		beforeAdmins      []values.TraPMemberID
		executeAddAdmin   bool
		AddAdminErr       error
		expectedAdmins    []*service.UserInfo
		isErr             bool
		err               error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())

	userInfo1 := service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive, false)
	userInfo2 := service.NewUserInfo(userID2, "mazrean", values.TrapMemberStatusActive, false)
	userInfo3 := service.NewUserInfo(userID3, "pikachu", values.TrapMemberStatusActive, false)

	activeUsers := []*service.UserInfo{userInfo1, userInfo2, userInfo3}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			userID:           userID1,
			executeGetAdmins: true,
			beforeAdmins:     []values.TraPMemberID{userID2},
			executeAddAdmin:  true,
			expectedAdmins:   []*service.UserInfo{userInfo2, userInfo1},
		},
		{
			description: "全ユーザーの取得に失敗したのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			getActiveUsersErr: errors.New("test"),
			isErr:             true,
		},
		{
			description: "存在しないユーザーなのでErrInvalidUserID",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			userID: values.NewTrapMemberID(uuid.New()),
			isErr:  true,
			err:    service.ErrInvalidUserID,
		},
		{
			description: "GetAdminsがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			userID:           userID1,
			executeGetAdmins: true,
			GetAdminsErr:     errors.New("test"),
			isErr:            true,
		},
		{
			description: "既にユーザーがadminなのでErrNoAdminsUpdated",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			userID:           userID1,
			executeGetAdmins: true,
			beforeAdmins:     []values.TraPMemberID{userID1},
			isErr:            true,
			err:              service.ErrNoAdminsUpdated,
		},
		{
			description: "AddAdminがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			userID:           userID1,
			executeGetAdmins: true,
			beforeAdmins:     []values.TraPMemberID{userID2},
			executeAddAdmin:  true,
			AddAdminErr:      errors.New("test"),
			isErr:            true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.getActiveUsersErr != nil {
				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(nil, errors.New("test"))
				mockUserAuth.
					EXPECT().
					GetActiveUsers(gomock.Any(), testCase.authSession).
					Return(nil, errors.New("test"))
			} else {
				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(activeUsers, nil)
			}

			if testCase.executeGetAdmins {
				mockAdminAuthRepository.
					EXPECT().
					GetAdmins(gomock.Any()).
					Return(testCase.beforeAdmins, testCase.GetAdminsErr)
			}

			if testCase.executeAddAdmin {
				mockAdminAuthRepository.
					EXPECT().
					AddAdmin(gomock.Any(), testCase.userID).
					Return(testCase.AddAdminErr)
			}

			adminInfos, err := adminAuthService.AddAdmin(ctx, testCase.authSession, testCase.userID)

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

			assert.Len(t, adminInfos, len(testCase.expectedAdmins))
			for i, adminInfo := range adminInfos {
				assert.Equal(t, testCase.expectedAdmins[i], adminInfo)
			}
		})

	}
}

func TestGetAdmins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockAdminAuthRepository := mockRepository.NewMockAdminAuthV2(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	user := NewUser(mockUserAuth, mockUserCache)

	adminAuthService := NewAdminAuth(mockDB, mockAdminAuthRepository, user)

	type test struct {
		description        string
		authSession        *domain.OIDCSession
		getActiveUsersErr  error
		executeGetAdmins   bool
		adminIDs           []values.TraPMemberID
		GetAdminsErr       error
		expectedAdminInfos []*service.UserInfo
		isErr              bool
		err                error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())

	userInfo1 := service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive, false)
	userInfo2 := service.NewUserInfo(userID2, "mazrean", values.TrapMemberStatusActive, false)
	userInfo3 := service.NewUserInfo(userID3, "pikachu", values.TrapMemberStatusActive, false)

	activeUsers := []*service.UserInfo{userInfo1, userInfo2, userInfo3}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetAdmins:   true,
			adminIDs:           []values.TraPMemberID{userID1, userID2},
			expectedAdminInfos: []*service.UserInfo{userInfo1, userInfo2},
		},
		{
			description: "全ユーザーの取得に失敗したのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			getActiveUsersErr: errors.New("test"),
			isErr:             true,
		},
		{
			description: "GetAdminsがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetAdmins: true,
			GetAdminsErr:     errors.New("test"),
			isErr:            true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.getActiveUsersErr != nil {
				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(nil, errors.New("test"))
				mockUserAuth.
					EXPECT().
					GetActiveUsers(gomock.Any(), testCase.authSession).
					Return(nil, errors.New("test"))
			} else {
				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(activeUsers, nil)
			}

			if testCase.executeGetAdmins {
				mockAdminAuthRepository.
					EXPECT().
					GetAdmins(gomock.Any()).
					Return(testCase.adminIDs, testCase.GetAdminsErr)
			}

			adminInfos, err := adminAuthService.GetAdmins(ctx, testCase.authSession)

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

			assert.Len(t, adminInfos, len(testCase.expectedAdminInfos))
			for i, adminInfo := range adminInfos {
				assert.Equal(t, testCase.expectedAdminInfos[i], adminInfo)
			}
		})

	}
}

func TestDeleteAdmin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockAdminAuthRepository := mockRepository.NewMockAdminAuthV2(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	user := NewUser(mockUserAuth, mockUserCache)

	adminAuthService := NewAdminAuth(mockDB, mockAdminAuthRepository, user)

	type test struct {
		description           string
		authSession           *domain.OIDCSession
		getMeErr              error
		myInfo                *service.UserInfo
		executeGetActiveUsers bool
		getActiveUsersErr     error
		userID                values.TraPMemberID
		executeDeleteAdmin    bool
		DeleteAdminsErr       error
		executeGetAdmins      bool
		afterAdminIDs         []values.TraPMemberID
		GetAdminsErr          error
		expectedAdminInfos    []*service.UserInfo
		isErr                 bool
		err                   error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())

	userInfo1 := service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive, false)
	userInfo2 := service.NewUserInfo(userID2, "mazrean", values.TrapMemberStatusActive, false)
	userInfo3 := service.NewUserInfo(userID3, "pikachu", values.TrapMemberStatusActive, false)

	activeUsers := []*service.UserInfo{userInfo1, userInfo2, userInfo3}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                userID2,
			executeDeleteAdmin:    true,
			executeGetAdmins:      true,
			afterAdminIDs:         []values.TraPMemberID{userID1},
			expectedAdminInfos:    []*service.UserInfo{userInfo1},
		},
		{
			description: "getMeがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			getMeErr: errors.New("test"),
			isErr:    true,
		},
		{
			description: "自分を削除しようとしているのでErrCannotDeleteMeFromAdmins",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo: userInfo1,
			userID: userID1,
			isErr:  true,
			err:    service.ErrCannotDeleteMeFromAdmins,
		},
		{
			description: "getActiveUsersがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                userID2,
			getActiveUsersErr:     errors.New("test"),
			isErr:                 true,
		},
		{
			description: "ユーザーが存在しないのでErrInvalidUserID",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                values.NewTrapMemberID(uuid.New()),
			isErr:                 true,
			err:                   service.ErrInvalidUserID,
		},
		{
			description: "DeleteAdminがErrNoRecordDeletedなのでErrNotAdmin",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                userID2,
			executeDeleteAdmin:    true,
			DeleteAdminsErr:       repository.ErrNoRecordDeleted,
			isErr:                 true,
			err:                   service.ErrNotAdmin,
		},
		{
			description: "DeleteAdminがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                userID2,
			executeDeleteAdmin:    true,
			DeleteAdminsErr:       errors.New("test"),
			isErr:                 true,
		},
		{
			description: "GetAdminsがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			myInfo:                userInfo1,
			executeGetActiveUsers: true,
			userID:                userID2,
			executeDeleteAdmin:    true,
			executeGetAdmins:      true,
			GetAdminsErr:          errors.New("test"),
			isErr:                 true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.getMeErr != nil {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("test"))
				mockUserAuth.
					EXPECT().
					GetMe(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("test"))
			} else {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), gomock.Any()).
					Return(testCase.myInfo, nil)
			}

			if testCase.executeGetActiveUsers {
				if testCase.getActiveUsersErr != nil {
					mockUserCache.
						EXPECT().
						GetActiveUsers(gomock.Any()).
						Return(nil, errors.New("test"))
					mockUserAuth.
						EXPECT().
						GetActiveUsers(gomock.Any(), gomock.Any()).
						Return(nil, errors.New("test"))
				} else {
					mockUserCache.
						EXPECT().
						GetActiveUsers(gomock.Any()).
						Return(activeUsers, nil)
				}
			}

			if testCase.executeDeleteAdmin {
				mockAdminAuthRepository.
					EXPECT().
					DeleteAdmin(ctx, testCase.userID).
					Return(testCase.DeleteAdminsErr)
			}

			if testCase.executeGetAdmins {
				mockAdminAuthRepository.
					EXPECT().
					GetAdmins(ctx).
					Return(testCase.afterAdminIDs, testCase.GetAdminsErr)
			}

			adminInfos, err := adminAuthService.DeleteAdmin(ctx, testCase.authSession, testCase.userID)

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

			assert.Len(t, adminInfos, len(testCase.expectedAdminInfos))
			for i, adminInfo := range adminInfos {
				assert.Equal(t, testCase.expectedAdminInfos[i], adminInfo)
			}
		})

	}
}

func TestAdminAuthorize(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockAdminAuthRepository := mockRepository.NewMockAdminAuthV2(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	user := NewUser(mockUserAuth, mockUserCache)

	adminAuthService := NewAdminAuth(mockDB, mockAdminAuthRepository, user)

	type test struct {
		description      string
		authSession      *domain.OIDCSession
		executeGetMe     bool
		myInfo           *service.UserInfo
		getMeErr         error
		executeGetAdmins bool
		adminIDs         []values.TraPMemberID
		GetAdminsErr     error
		isErr            bool
		err              error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())

	myInfo1 := service.NewUserInfo(userID1, "ikura-hamu", values.TrapMemberStatusActive, false)
	myInfo3 := service.NewUserInfo(userID3, "mazrean", values.TrapMemberStatusActive, false)

	adminIDs := []values.TraPMemberID{userID1, userID2}

	testCases := []test{
		{
			description: "特に問題ないのでエラー無し",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetMe:     true,
			myInfo:           myInfo1,
			executeGetAdmins: true,
			adminIDs:         adminIDs,
		},
		{
			description: "sessionが期限切れなのでErrOIDCSessionExpired",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(-time.Hour),
			),
			isErr: true,
			err:   service.ErrOIDCSessionExpired,
		},
		{
			description: "getMeがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetMe: true,
			getMeErr:     errors.New("test"),
			isErr:        true,
		},
		{
			description: "GetAdminsがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetMe:     true,
			myInfo:           myInfo1,
			executeGetAdmins: true,
			GetAdminsErr:     errors.New("test"),
			isErr:            true,
		},
		{
			description: "ユーザーがadminではないのでErrForbidden",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			executeGetMe:     true,
			myInfo:           myInfo3,
			executeGetAdmins: true,
			isErr:            true,
			err:              service.ErrForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetMe {
				if testCase.getMeErr != nil {
					mockUserCache.
						EXPECT().
						GetMe(gomock.Any(), gomock.Any()).
						Return(nil, errors.New("test"))
					mockUserAuth.
						EXPECT().
						GetMe(gomock.Any(), gomock.Any()).
						Return(nil, errors.New("test"))
				} else {
					mockUserCache.
						EXPECT().
						GetMe(gomock.Any(), gomock.Any()).
						Return(testCase.myInfo, nil)
				}
			}

			if testCase.executeGetAdmins {
				mockAdminAuthRepository.
					EXPECT().
					GetAdmins(gomock.Any()).
					Return(testCase.adminIDs, testCase.GetAdminsErr)
			}

			err := adminAuthService.AdminAuthorize(ctx, testCase.authSession)

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
