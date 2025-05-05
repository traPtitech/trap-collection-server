package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/auth/mock"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGenerateAuthState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	client, session, err := oidcService.GenerateAuthState(ctx)
	assert.NoError(t, err)

	assert.Equal(t, oidcService.client, client)
	assert.Equal(t, values.OIDCCodeChallengeMethodSha256, session.GetCodeChallengeMethod())
}

func TestCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	type test struct {
		description       string
		GetOIDCSessionErr error
		isErr             bool
		err               error
	}

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
		},
		{
			description:       "GetOIDCSessionでErrInvalidCredentials",
			GetOIDCSessionErr: auth.ErrInvalidCredentials,
			isErr:             true,
			err:               service.ErrInvalidAuthStateOrCode,
		},
		{
			description:       "GetOIDCSessionでエラー",
			GetOIDCSessionErr: errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			codeVerifier, err := values.NewOIDCCodeVerifier()
			assert.NoError(t, err)

			code := values.NewOIDCAuthorizationCode("")
			authState := domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			)
			var session *domain.OIDCSession
			if testCase.GetOIDCSessionErr == nil {
				session = domain.NewOIDCSession(
					values.NewOIDCAccessToken("access token"),
					time.Now(),
				)
			}
			mockOIDCAuth.
				EXPECT().
				GetOIDCSession(ctx, oidcService.client, code, authState).
				Return(session, testCase.GetOIDCSessionErr)

			actualSession, err := oidcService.Callback(ctx, authState, code)

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

			assert.Equal(t, session, actualSession)
		})
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	type test struct {
		description          string
		RevokeOIDCSessionErr error
		isErr                bool
		err                  error
	}

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
		},
		{
			description:          "GetOIDCSessionでエラー",
			RevokeOIDCSessionErr: errors.New("error"),
			isErr:                true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("access token"),
				time.Now(),
			)
			mockOIDCAuth.
				EXPECT().
				RevokeOIDCSession(ctx, session).
				Return(testCase.RevokeOIDCSessionErr)

			err := oidcService.Logout(ctx, session)

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

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	type test struct {
		description string
		isExpired   bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "期限前なので問題なし",
		},
		{
			description: "期限切れなのでエラー",
			isExpired:   true,
			isErr:       true,
			err:         service.ErrOIDCSessionExpired,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expiresAt time.Time
			if testCase.isExpired {
				expiresAt = time.Now().Add(-1 * time.Hour)
			} else {
				expiresAt = time.Now().Add(1 * time.Hour)
			}

			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("access token"),
				expiresAt,
			)

			err := oidcService.Authenticate(ctx, session)

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

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	type test struct {
		description      string
		cacheUser        *service.UserInfo
		cacheGetMeErr    error
		executeAuthGetMe bool
		authUser         *service.UserInfo
		authGetMeErr     error
		cacheSetMeErr    error
		user             *service.UserInfo
		isErr            bool
		err              error
	}

	userInfo := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("mazrean"),
		values.TrapMemberStatusActive,
		false,
	)

	testCases := []test{
		{
			description: "cacheがhitするのでエラーなし",
			cacheUser:   userInfo,
			user:        userInfo,
		},
		{
			description:      "cacheがhitしないがauthからの取り出しに成功するのでエラーなし",
			cacheGetMeErr:    cache.ErrCacheMiss,
			executeAuthGetMe: true,
			authUser:         userInfo,
			user:             userInfo,
		},
		{
			description:      "cacheがエラー(ErrCacheMiss以外)でもauthからの取り出しに成功するのでエラーなし",
			cacheGetMeErr:    errors.New("cache error"),
			executeAuthGetMe: true,
			authUser:         userInfo,
			user:             userInfo,
		},
		{
			description:      "cacheがhitせずauthからの取り出しがエラーなのでエラー",
			cacheGetMeErr:    cache.ErrCacheMiss,
			executeAuthGetMe: true,
			authGetMeErr:     errors.New("auth error"),
			isErr:            true,
		},
		{
			description:      "cacheがhitしないがauthからの取り出しに成功するのでcache設定に失敗してもエラーなし",
			cacheGetMeErr:    cache.ErrCacheMiss,
			executeAuthGetMe: true,
			authUser:         userInfo,
			cacheSetMeErr:    errors.New("cache error"),
			user:             userInfo,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("access token"),
				time.Now(),
			)

			mockUserCache.
				EXPECT().
				GetMe(ctx, session.GetAccessToken()).
				Return(testCase.cacheUser, testCase.cacheGetMeErr)
			if testCase.executeAuthGetMe {
				mockUserAuth.
					EXPECT().
					GetMe(ctx, session).
					Return(testCase.authUser, testCase.authGetMeErr)
				if testCase.authGetMeErr == nil {
					mockUserCache.
						EXPECT().
						SetMe(ctx, session, testCase.authUser).
						Return(testCase.cacheSetMeErr)
				}
			}

			user, err := oidcService.GetMe(ctx, session)

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

			assert.Equal(t, testCase.user, user)
		})
	}
}

func TestGetActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCAuth := mock.NewMockOIDC(ctrl)

	mockConf := mockConfig.NewMockServiceV1(ctrl)
	mockConf.
		EXPECT().
		ClientID().
		Return("clientID", nil)
	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)
	user := NewUser(mockUserAuth, mockUserCache)
	oidcService, err := NewOIDC(mockConf, user, mockOIDCAuth)
	if err != nil {
		t.Fatalf("failed to create oidc service: %v", err)
		return
	}

	type test struct {
		description                  string
		cacheUsers                   []*service.UserInfo
		cacheGetAllActiveUsersErr    error
		executeAuthGetAllActiveUsers bool
		authUsers                    []*service.UserInfo
		authGetAllActiveUsersErr     error
		cacheSetAllActiveUsersErr    error
		users                        []*service.UserInfo
		isErr                        bool
		err                          error
	}

	users := []*service.UserInfo{
		service.NewUserInfo(
			values.NewTrapMemberID(uuid.New()),
			values.NewTrapMemberName("mazrean"),
			values.TrapMemberStatusActive,
			false,
		),
	}

	testCases := []test{
		{
			description: "cacheがhitするのでエラーなし",
			cacheUsers:  users,
			users:       users,
		},
		{
			description:                  "cacheがhitしないがauthからの取り出しに成功するのでエラーなし",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			users:                        users,
		},
		{
			description:                  "cacheがエラー(ErrCacheMiss以外)でもauthからの取り出しに成功するのでエラーなし",
			cacheGetAllActiveUsersErr:    errors.New("cache error"),
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			users:                        users,
		},
		{
			description:                  "cacheがhitせずauthからの取り出しがエラーなのでエラー",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authGetAllActiveUsersErr:     errors.New("auth error"),
			isErr:                        true,
		},
		{
			description:                  "cacheがhitしないがauthからの取り出しに成功するのでcache設定に失敗してもエラーなし",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			cacheSetAllActiveUsersErr:    errors.New("cache error"),
			users:                        users,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("access token"),
				time.Now(),
			)

			mockUserCache.
				EXPECT().
				GetActiveUsers(ctx).
				Return(testCase.cacheUsers, testCase.cacheGetAllActiveUsersErr)
			if testCase.executeAuthGetAllActiveUsers {
				mockUserAuth.
					EXPECT().
					GetActiveUsers(ctx, session).
					Return(testCase.authUsers, testCase.authGetAllActiveUsersErr)
				if testCase.authGetAllActiveUsersErr == nil {
					mockUserCache.
						EXPECT().
						SetActiveUsers(ctx, testCase.authUsers).
						Return(testCase.cacheSetAllActiveUsersErr)
				}
			}

			users, err := oidcService.GetActiveUsers(ctx, session)

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

			assert.Equal(t, testCase.users, users)
		})
	}
}
