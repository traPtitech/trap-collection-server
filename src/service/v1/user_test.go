package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userService := NewUser(mockUserAuth, mockUserCache)

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
