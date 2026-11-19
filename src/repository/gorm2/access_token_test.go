package gorm2

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

func TestGetAccessTokenInfo(t *testing.T) {
	type test struct {
		description        string
		statusName         string
		statusActive       bool
		invalidAccessToken bool
		lockType           repository.LockType
		wantStatus         values.LauncherUserStatus
		isErr              bool
		err                error
		errMessage         string
	}

	testCases := []test{
		{
			description:  "有効なプロダクトキーの状態を取得できる",
			statusName:   schema.ProductKeyStatusActive,
			statusActive: true,
			wantStatus:   values.LauncherUserStatusActive,
		},
		{
			description:  "無効なプロダクトキーの状態を取得できる",
			statusName:   schema.ProductKeyStatusInactive,
			statusActive: true,
			wantStatus:   values.LauncherUserStatusInactive,
		},
		{
			description:  "不明なプロダクトキーの状態なのでエラー",
			statusName:   "unknown",
			statusActive: true,
			isErr:        true,
			errMessage:   "invalid product key status:",
		},
		{
			description:  "プロダクトキーの状態が空なのでエラー",
			statusName:   "",
			statusActive: true,
			isErr:        true,
			errMessage:   "invalid product key status:",
		},
		{
			description: "有効状態のマスタが無効化されているのでエラー",
			statusName:  schema.ProductKeyStatusActive,
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "無効状態のマスタが無効化されているのでエラー",
			statusName:  schema.ProductKeyStatusInactive,
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description:        "アクセストークンが存在しないのでエラー",
			statusName:         schema.ProductKeyStatusActive,
			statusActive:       true,
			invalidAccessToken: true,
			isErr:              true,
			err:                repository.ErrRecordNotFound,
		},
		{
			description:  "ロック付きでプロダクトキーの状態を取得できる",
			statusName:   schema.ProductKeyStatusActive,
			statusActive: true,
			lockType:     repository.LockTypeRecord,
			wantStatus:   values.LauncherUserStatusActive,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ctx := t.Context()
			// 共有マスタの product_key_statuses を変更するため、他のテストに影響しないよう、
			// ケース終了時に作成したデータとともにロールバックする。
			tx := testDB.db.Begin()
			require.NoError(t, tx.Error)
			t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
			repositoryDB := &DB{db: tx}
			db, err := repositoryDB.getDB(ctx)
			require.NoError(t, err)
			accessTokenRepository := NewAccessToken(repositoryDB)

			var status schema.ProductKeyStatusTable
			require.NoError(t, db.Where("name = ?", testCase.statusName).
				FirstOrCreate(&status, schema.ProductKeyStatusTable{Name: testCase.statusName}).Error)
			require.NoError(t, db.Model(&status).Update("active", testCase.statusActive).Error)

			now := time.Now().Truncate(time.Second)
			editionID := values.NewEditionID()
			productKeyID := values.NewLauncherUserID()
			accessTokenID := values.NewLauncherSessionID()
			productKey, err := values.NewLauncherUserProductKey()
			require.NoError(t, err)
			accessToken, err := values.NewLauncherSessionAccessToken()
			require.NoError(t, err)

			dbEdition := schema.EditionTable{
				ID: uuid.UUID(editionID), Name: "TestGetAccessTokenInfo", CreatedAt: now,
			}
			require.NoError(t, db.Create(&dbEdition).Error)
			dbProductKey := schema.ProductKeyTable{
				ID: uuid.UUID(productKeyID), EditionID: uuid.UUID(editionID),
				ProductKey: string(productKey), StatusID: status.ID, CreatedAt: now,
			}
			require.NoError(t, db.Create(&dbProductKey).Error)
			dbAccessToken := schema.AccessTokenTable{
				ID: uuid.UUID(accessTokenID), ProductKeyID: uuid.UUID(productKeyID),
				AccessToken: string(accessToken), ExpiresAt: now.Add(time.Hour), CreatedAt: now,
			}
			require.NoError(t, db.Create(&dbAccessToken).Error)

			if testCase.invalidAccessToken {
				accessToken, err = values.NewLauncherSessionAccessToken()
				require.NoError(t, err)
			}
			info, err := accessTokenRepository.GetAccessTokenInfo(ctx, accessToken, testCase.lockType)
			if testCase.isErr {
				require.Error(t, err)
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				}
				if testCase.errMessage != "" {
					assert.ErrorContains(t, err, testCase.errMessage)
				}
				assert.Nil(t, info)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, productKeyID, info.ProductKey.GetID())
			assert.Equal(t, productKey, info.ProductKey.GetProductKey())
			assert.Equal(t, testCase.wantStatus, info.ProductKey.GetStatus())
			assert.Equal(t, accessTokenID, info.AccessToken.GetID())
			assert.Equal(t, accessToken, info.AccessToken.GetAccessToken())
			assert.WithinDuration(t, dbAccessToken.ExpiresAt, info.AccessToken.GetExpiresAt(), time.Second)
			assert.Equal(t, editionID, info.Edition.GetID())
			assert.Equal(t, values.NewEditionName(dbEdition.Name), info.Edition.GetName())
		})
	}
}
