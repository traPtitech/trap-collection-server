package ristretto

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
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

	mockConf := mock.NewMockCacheRistretto(ctrl)
	mockConf.
		EXPECT().
		ActiveUsersTTL().
		Return(time.Hour, nil)
	userCache, err := NewUser(mockConf)
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		userInfo    *service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			keyExist:    true,
			userInfo: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				values.NewTrapMemberName("mazrean"),
				values.TrapMemberStatusActive,
				false,
			),
		},
		{
			description: "キーが存在しないのでErrCacheMiss",
			keyExist:    false,
			isErr:       true,
			err:         cache.ErrCacheMiss,
		},
	}

	for i, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			accessToken := values.NewOIDCAccessToken(fmt.Sprintf("access token%d", i))
			if testCase.keyExist {
				ok := userCache.meCache.Set(string(accessToken), testCase.userInfo, 8)
				assert.True(t, ok)

				userCache.meCache.Wait()
			}

			user, err := userCache.GetMe(ctx, accessToken)

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

			assert.Equal(t, testCase.userInfo, user)
		})
	}
}

func TestSetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type test struct {
		description string
		keyExist    bool
		beforeValue *service.UserInfo
		accessToken values.OIDCAccessToken
		userInfo    *service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			accessToken: values.NewOIDCAccessToken("access token1"),
			userInfo: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				values.NewTrapMemberName("mazrean"),
				values.TrapMemberStatusActive,
				false,
			),
		},
		{
			description: "元からキーがあっても上書きする",
			keyExist:    true,
			beforeValue: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				values.NewTrapMemberName("mazrean"),
				values.TrapMemberStatusActive,
				false,
			),
			accessToken: values.NewOIDCAccessToken("access token2"),
			userInfo: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				values.NewTrapMemberName("mazrean"),
				values.TrapMemberStatusActive,
				false,
			),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockConf := mock.NewMockCacheRistretto(ctrl)
			mockConf.
				EXPECT().
				ActiveUsersTTL().
				Return(time.Hour, nil)
			userCache, err := NewUser(mockConf)
			if err != nil {
				t.Fatalf("failed to create user cache: %v", err)
			}

			if testCase.keyExist {
				ok := userCache.meCache.Set(string(testCase.accessToken), testCase.beforeValue, 8)
				assert.True(t, ok)

				userCache.meCache.Wait()
			}

			session := domain.NewOIDCSession(testCase.accessToken, time.Now().Add(2*time.Second))

			err = userCache.SetMe(ctx, session, testCase.userInfo)

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

			// キャッシュが設定されるまで待機
			userCache.meCache.Wait()

			// OIDCSessionの期限前なのでキャッシュされている
			value, ok := userCache.meCache.Get(string(testCase.accessToken))
			assert.True(t, ok)
			assert.Equal(t, testCase.userInfo, value)

			<-time.NewTimer(2 * time.Second).C

			// OIDCSessionの期限が切れたらキャッシュは削除される
			_, ok = userCache.meCache.Get(string(testCase.accessToken))
			assert.False(t, ok)
		})
	}
}

func TestGetAllActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mock.NewMockCacheRistretto(ctrl)
	mockConf.
		EXPECT().
		ActiveUsersTTL().
		Return(time.Hour, nil)
	userCache, err := NewUser(mockConf)
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		users       []*service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			keyExist:    true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
		{
			description: "ユーザー数が500人でも問題なし",
			keyExist:    true,
			users:       make([]*service.UserInfo, 500),
		},
		{
			description: "キーが存在しないのでErrCacheMiss",
			keyExist:    false,
			isErr:       true,
			err:         cache.ErrCacheMiss,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.keyExist {
				ok := userCache.activeUsers.Set(activeUsersKey, testCase.users, 8)
				assert.True(t, ok)

				userCache.activeUsers.Wait()
				defer userCache.activeUsers.Clear()
			}

			users, err := userCache.GetAllActiveUsers(ctx)

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

func TestSetAllActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mock.NewMockCacheRistretto(ctrl)
	mockConf.
		EXPECT().
		ActiveUsersTTL().
		Return(time.Hour, nil)
	userCache, err := NewUser(mockConf)
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		beforeValue []*service.UserInfo
		users       []*service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
		{
			description: "ユーザー数が500人でもエラーなし",
			users:       make([]*service.UserInfo, 500),
		},
		{
			description: "元からキーがあっても上書きする",
			keyExist:    true,
			beforeValue: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.keyExist {
				ok := userCache.activeUsers.Set(activeUsersKey, testCase.beforeValue, 1)
				assert.True(t, ok)

				userCache.activeUsers.Wait()
			}

			err := userCache.SetAllActiveUsers(ctx, testCase.users)

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

			// キャッシュが設定されるまで待機
			userCache.activeUsers.Wait()

			// OIDCSessionの期限前なのでキャッシュされている
			actualUsers, ok := userCache.activeUsers.Get(activeUsersKey)
			assert.True(t, ok)

			for i, user := range testCase.users {
				if user == nil {
					assert.Nil(t, actualUsers[i])
				} else {
					assert.Equal(t, *actualUsers[i], *user)
				}
			}
		})
	}
}

func TestGetActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mock.NewMockCacheRistretto(ctrl)
	mockConf.
		EXPECT().
		ActiveUsersTTL().
		Return(time.Hour, nil)
	userCache, err := NewUser(mockConf)
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		users       []*service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			keyExist:    true,
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
		{
			description: "ユーザー数が500人でも問題なし",
			keyExist:    true,
			users:       make([]*service.UserInfo, 500),
		},
		{
			description: "キーが存在しないのでErrCacheMiss",
			keyExist:    false,
			isErr:       true,
			err:         cache.ErrCacheMiss,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.keyExist {
				ok := userCache.activeUsers.Set(activeUsersKey, testCase.users, 8)
				assert.True(t, ok)

				userCache.activeUsers.Wait()
				defer userCache.activeUsers.Clear()
			}

			users, err := userCache.GetActiveUsers(ctx)

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

func TestSetActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mock.NewMockCacheRistretto(ctrl)
	mockConf.
		EXPECT().
		ActiveUsersTTL().
		Return(time.Hour, nil)
	userCache, err := NewUser(mockConf)
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		beforeValue []*service.UserInfo
		users       []*service.UserInfo
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
		{
			description: "ユーザー数が500人でもエラーなし",
			users:       make([]*service.UserInfo, 500),
		},
		{
			description: "元からキーがあっても上書きする",
			keyExist:    true,
			beforeValue: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(uuid.New()),
					values.NewTrapMemberName("mazrean"),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.keyExist {
				ok := userCache.activeUsers.Set(activeUsersKey, testCase.beforeValue, 1)
				assert.True(t, ok)

				userCache.activeUsers.Wait()
			}

			err := userCache.SetActiveUsers(ctx, testCase.users)

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

			// キャッシュが設定されるまで待機
			userCache.activeUsers.Wait()

			// OIDCSessionの期限前なのでキャッシュされている
			actualUsers, ok := userCache.activeUsers.Get(activeUsersKey)
			assert.True(t, ok)

			for i, user := range testCase.users {
				if user == nil {
					assert.Nil(t, actualUsers[i])
				} else {
					assert.Equal(t, *actualUsers[i], *user)
				}
			}
		})
	}
}
