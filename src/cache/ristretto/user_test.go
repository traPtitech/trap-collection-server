package ristretto

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userCache, err := NewUser()
	if err != nil {
		t.Fatalf("failed to create user cache: %v", err)
	}

	type test struct {
		description string
		keyExist    bool
		valueBroken bool
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
			),
		},
		{
			description: "キーが存在しないのでErrCacheMiss",
			keyExist:    false,
			isErr:       true,
			err:         cache.ErrCacheMiss,
		},
		{
			// 実際には発生しないが念の為確認
			description: "値が壊れているのでエラー",
			keyExist:    true,
			valueBroken: true,
			isErr:       true,
		},
	}

	for i, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			accessToken := values.NewOIDCAccessToken(fmt.Sprintf("access token%d", i))
			if testCase.keyExist {
				if testCase.valueBroken {
					ok := userCache.meCache.Set(string(accessToken), "broken", 8)
					assert.True(t, ok)

					userCache.meCache.Wait()
				} else {
					ok := userCache.meCache.Set(string(accessToken), testCase.userInfo, 8)
					assert.True(t, ok)

					userCache.meCache.Wait()
				}
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
