package ristretto

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				ok := userCache.users.Set(activeUsersKey, testCase.users, 8)
				assert.True(t, ok)

				userCache.users.Wait()
				defer userCache.users.Clear()
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
				ok := userCache.users.Set(activeUsersKey, testCase.beforeValue, 1)
				assert.True(t, ok)

				userCache.users.Wait()
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
			userCache.users.Wait()

			// OIDCSessionの期限前なのでキャッシュされている
			actualUsers, ok := userCache.users.Get(activeUsersKey)
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
				ok := userCache.users.Set(activeUsersKey, testCase.users, 8)
				assert.True(t, ok)

				userCache.users.Wait()
				defer userCache.users.Clear()
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
				ok := userCache.users.Set(activeUsersKey, testCase.beforeValue, 1)
				assert.True(t, ok)

				userCache.users.Wait()
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
			userCache.users.Wait()

			// OIDCSessionの期限前なのでキャッシュされている
			actualUsers, ok := userCache.users.Get(activeUsersKey)
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

func TestGetAllUsers(t *testing.T) {
	t.Parallel()

	user1 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("mazrean"),
		values.TrapMemberStatusDeactivated,
		false,
	)
	user2 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("ikura-hamu"),
		values.TrapMemberStatusActive,
		false,
	)

	ttl := time.Hour
	testCases := map[string]struct {
		users []*service.UserInfo
		after time.Duration
		err   error
	}{
		"ttl前なのでキャッシュされている": {
			users: []*service.UserInfo{user1, user2},
			after: time.Second,
		},
		"ttl後なのでキャッシュされていない": {
			users: []*service.UserInfo{user1, user2},
			after: ttl + time.Second,
			err:   cache.ErrCacheMiss,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				ctrl := gomock.NewController(t)
				mockConf := mock.NewMockCacheRistretto(ctrl)

				mockConf.
					EXPECT().
					ActiveUsersTTL().
					Return(ttl, nil)

				usersCache, err := NewUser(mockConf)
				require.NoError(t, err)

				t.Cleanup(func() {
					usersCache.users.Close()
					usersCache.meCache.Close()
				})

				if len(testCase.users) > 0 {
					ok := usersCache.users.SetWithTTL(allUsersKey, testCase.users, 1, ttl)
					require.True(t, ok)
					usersCache.users.Wait()
				}

				time.Sleep(testCase.after)

				users, err := usersCache.GetAllUsers(ctx)

				if testCase.err != nil {
					require.ErrorIs(t, err, testCase.err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, testCase.users, users)
				}
			})
		})
	}
}

func TestSetAllUsers(t *testing.T) {
	t.Parallel()

	ttl := time.Hour

	user1 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("mazrean"),
		values.TrapMemberStatusDeactivated,
		false,
	)
	user2 := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberName("ikura-hamu"),
		values.TrapMemberStatusActive,
		false,
	)

	testCases := map[string]struct {
		beforeUsers []*service.UserInfo
		users       []*service.UserInfo
		err         error
	}{
		"ユーザー情報をセットできる": {
			users: []*service.UserInfo{user1, user2},
		},
		"元からキーがあっても上書きする": {
			beforeUsers: []*service.UserInfo{user1},
			users:       []*service.UserInfo{user2},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				ctrl := gomock.NewController(t)
				mockConf := mock.NewMockCacheRistretto(ctrl)

				mockConf.
					EXPECT().
					ActiveUsersTTL().
					Return(ttl, nil)

				usersCache, err := NewUser(mockConf)
				require.NoError(t, err)

				t.Cleanup(func() {
					usersCache.users.Close()
					usersCache.meCache.Close()
				})

				if len(testCase.beforeUsers) > 0 {
					ok := usersCache.users.SetWithTTL(allUsersKey, testCase.beforeUsers, 1, ttl)
					require.True(t, ok)
					usersCache.users.Wait()
				}

				err = usersCache.SetAllUsers(ctx, testCase.users)
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
					return
				}
				assert.NoError(t, err)

				usersCache.users.Wait()

				users, ok := usersCache.users.Get(allUsersKey)
				assert.True(t, ok)
				assert.Equal(t, testCase.users, users)

				time.Sleep(ttl + time.Second)

				_, ok = usersCache.users.Get(allUsersKey)
				assert.False(t, ok)
			})

		})
	}

}
