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
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestAdministratorAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	type test struct {
		description    string
		authSession    *domain.OIDCSession
		administrators []string
		user           *service.UserInfo
		isGetMeErr     bool
		isErr          bool
		err            error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{
				"mazrean",
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
		},
		{
			description: "ユーザー情報の取得に失敗したのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{
				"mazrean",
			},
			isGetMeErr: true,
			isErr:      true,
		},
		{
			description: "ユーザーが管理者でないのでForbidden",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{
				"mazrean",
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean1",
				values.TrapMemberStatusActive,
			),
			isErr: true,
			err:   service.ErrForbidden,
		},
		{
			description: "管理者が複数でも管理者ならばエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{
				"mazrean",
				"mazrean1",
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
		},
		{
			description: "管理者が複数で管理者でないのでForbidden",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{
				"mazrean",
				"mazrean1",
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean2",
				values.TrapMemberStatusActive,
			),
			isErr: true,
			err:   service.ErrForbidden,
		},
		{
			description: "管理者がいないのでForbidden",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			administrators: []string{},
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			isErr: true,
			err:   service.ErrForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockConf := mock.NewMockServiceV1(ctrl)
			mockConf.
				EXPECT().
				Administrators().
				Return(testCase.administrators, nil)
			administratorAuthService, err := NewAdministratorAuth(mockConf, userUtils)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
				return
			}

			if testCase.isGetMeErr {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(nil, cache.ErrCacheMiss)
				mockUserAuth.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession).
					Return(nil, errors.New("error"))
			} else {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(testCase.user, nil)
			}

			err = administratorAuthService.AdministratorAuth(ctx, testCase.authSession)

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
