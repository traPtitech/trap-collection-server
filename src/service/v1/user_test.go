package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	userService := NewUser(userUtils)

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

	user := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("mazrean"),
		values.TrapMemberStatusActive,
		false,
	)

	testCases := []test{
		{
			description: "cacheがhitするのでエラーなし",
			cacheUser:   user,
			user:        user,
		},
		{
			description:      "cacheがhitしないがauthからの取り出しに成功するのでエラーなし",
			cacheGetMeErr:    cache.ErrCacheMiss,
			executeAuthGetMe: true,
			authUser:         user,
			user:             user,
		},
		{
			description:      "cacheがエラー(ErrCacheMiss以外)でもauthからの取り出しに成功するのでエラーなし",
			cacheGetMeErr:    errors.New("cache error"),
			executeAuthGetMe: true,
			authUser:         user,
			user:             user,
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
			authUser:         user,
			cacheSetMeErr:    errors.New("cache error"),
			user:             user,
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

			user, err := userService.GetMe(ctx, session)

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

func TestGetAllActiveUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	userService := NewUser(userUtils)

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
		includeBot                   bool
	}

	users := []*service.UserInfo{
		service.NewUserInfo(
			values.NewTrapMemberID(uuid.New()),
			values.NewTrapMemberName("mazrean"),
			values.TrapMemberStatusActive,
			false,
		),
	}
	users2 := []*service.UserInfo{
		service.NewUserInfo(
			values.NewTrapMemberID(uuid.New()),
			values.NewTrapMemberName("w4ma"),
			values.TrapMemberStatusActive,
			false,
		),
		service.NewUserInfo(
			values.NewTrapMemberID(uuid.New()),
			values.NewTrapMemberName("w4mabot"),
			values.TrapMemberStatusActive,
			true,
		),
	}

	testCases := []test{
		{
			description: "cacheがhitするのでエラーなし",
			cacheUsers:  users,
			users:       users,
			includeBot:  true,
		},
		{
			description: "botを除外する設定でcacheUsersにbotが含まれる",
			cacheUsers:  users2,
			users:       []*service.UserInfo{users2[0]},
			includeBot:  false,
		},
		{
			description:                  "botを除外する設定でcacheがhitしないがauthからbotを含むユーザー情報を取り出す",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users2,
			users:                        []*service.UserInfo{users2[0]},
			includeBot:                   false,
		},
		{
			description:                  "cacheがhitしないがauthからの取り出しに成功するのでエラーなし",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			users:                        users,
			includeBot:                   true,
		},
		{
			description:                  "cacheがエラー(ErrCacheMiss以外)でもauthからの取り出しに成功するのでエラーなし",
			cacheGetAllActiveUsersErr:    errors.New("cache error"),
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			users:                        users,
			includeBot:                   true,
		},
		{
			description:                  "cacheがhitせずauthからの取り出しがエラーなのでエラー",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authGetAllActiveUsersErr:     errors.New("auth error"),
			isErr:                        true,
			includeBot:                   true,
		},
		{
			description:                  "cacheがhitしないがauthからの取り出しに成功するのでcache設定に失敗してもエラーなし",
			cacheGetAllActiveUsersErr:    cache.ErrCacheMiss,
			executeAuthGetAllActiveUsers: true,
			authUsers:                    users,
			cacheSetAllActiveUsersErr:    errors.New("cache error"),
			users:                        users,
			includeBot:                   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("access token"),
				time.Now(),
			)
			includeBot := testCase.includeBot

			mockUserCache.
				EXPECT().
				GetAllActiveUsers(ctx).
				Return(testCase.cacheUsers, testCase.cacheGetAllActiveUsersErr)
			if testCase.executeAuthGetAllActiveUsers {
				mockUserAuth.
					EXPECT().
					GetAllActiveUsers(ctx, session).
					Return(testCase.authUsers, testCase.authGetAllActiveUsersErr)
				if testCase.authGetAllActiveUsersErr == nil {
					mockUserCache.
						EXPECT().
						SetAllActiveUsers(ctx, testCase.authUsers).
						Return(testCase.cacheSetAllActiveUsersErr)
				}
			}

			users, err := userService.GetAllActiveUser(ctx, session, includeBot)

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
