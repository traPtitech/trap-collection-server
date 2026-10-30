package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	serviceV2 "github.com/traPtitech/trap-collection-server/src/service/v2"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLite exercises the real query and scan; MariaDB migrations and locking need the integration suite.
func newAccessTokenSQLiteFixture(t *testing.T, statusName string, statusRowActive bool) (*DB, *domain.LauncherSession, values.EditionID) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	for _, statement := range []string{
		`CREATE TABLE editions (id TEXT PRIMARY KEY, name TEXT, questionnaire_url TEXT, created_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE product_key_statuses (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN)`,
		`CREATE TABLE product_keys (id TEXT PRIMARY KEY, edition_id TEXT, product_key TEXT, status_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE access_tokens (id TEXT PRIMARY KEY, product_key_id TEXT, access_token TEXT, expires_at DATETIME, created_at DATETIME, deleted_at DATETIME)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}

	now := time.Now()
	editionID := values.NewEditionID()
	keyID := values.NewLauncherUserID()
	require.NoError(t, db.Create(&schema.EditionTable{
		ID: uuid.UUID(editionID), Name: "synthetic edition", CreatedAt: now,
	}).Error)
	require.NoError(t, db.Exec("INSERT INTO product_key_statuses (id, name, active) VALUES (?, ?, ?)", 1, statusName, statusRowActive).Error)
	require.NoError(t, db.Create(&schema.ProductKeyTable{
		ID: uuid.UUID(keyID), EditionID: uuid.UUID(editionID),
		ProductKey: "aaaaa-bbbbb-ccccc-ddddd-eeeee", StatusID: 1, CreatedAt: now,
	}).Error)

	token := domain.NewLauncherSession(
		values.NewLauncherSessionID(),
		values.NewLauncherSessionAccessTokenFromString("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		now.Add(time.Hour),
	)
	repositoryDB := &DB{db: db}
	require.NoError(t, NewAccessToken(repositoryDB).SaveAccessToken(context.Background(), keyID, token))
	return repositoryDB, token, editionID
}

func TestGetAccessTokenInfoProductKeyStatus(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    string
		rowActive bool
		want      values.LauncherUserStatus
		invalid   bool
	}{
		{name: "active", status: schema.ProductKeyStatusActive, rowActive: true, want: values.LauncherUserStatusActive},
		{name: "inactive", status: schema.ProductKeyStatusInactive, rowActive: true, want: values.LauncherUserStatusInactive},
		{name: "unknown", status: "unknown", rowActive: true, invalid: true},
		{name: "empty", status: "", rowActive: true, invalid: true},
		{name: "disabled active status row", status: schema.ProductKeyStatusActive},
		{name: "disabled inactive status row", status: schema.ProductKeyStatusInactive},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, token, editionID := newAccessTokenSQLiteFixture(t, tc.status, tc.rowActive)
			info, err := NewAccessToken(db).GetAccessTokenInfo(context.Background(), token.GetAccessToken(), repository.LockTypeNone)
			if tc.invalid {
				require.ErrorContains(t, err, "invalid product key status:")
				require.Nil(t, info)
				return
			}
			if !tc.rowActive {
				require.ErrorIs(t, err, repository.ErrRecordNotFound)
				require.Nil(t, info)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, info.ProductKey.GetStatus())
			require.Equal(t, token.GetAccessToken(), info.AccessToken.GetAccessToken())
			require.Equal(t, editionID, info.Edition.GetID())
		})
	}
}

func TestEditionAuthorizationAfterProductKeyRevocation(t *testing.T) {
	ctx := context.Background()
	db, token, editionID := newAccessTokenSQLiteFixture(t, schema.ProductKeyStatusActive, true)
	ctrl := gomock.NewController(t)
	editions := mockRepository.NewMockEdition(ctrl)
	auth := serviceV2.NewEditionAuth(db, editions, NewProductKey(db), NewAccessToken(db))
	gameID := values.NewGameID()
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	fileID := values.NewGameFileID()
	editions.EXPECT().GetEditionGameVersionByGameID(ctx, editionID, gameID, repository.LockTypeNone).Return(nil, nil)
	editions.EXPECT().GetEditionGameVersionByImageID(ctx, editionID, imageID, repository.LockTypeNone).Return(nil, nil)
	editions.EXPECT().GetEditionGameVersionByVideoID(ctx, editionID, videoID, repository.LockTypeNone).Return(nil, nil)
	editions.EXPECT().GetEditionGameVersionByFileID(ctx, editionID, fileID, repository.LockTypeNone).Return(nil, nil)

	checks := []struct {
		name string
		run  func() (*domain.LauncherUser, *domain.Edition, error)
	}{
		{"edition", func() (*domain.LauncherUser, *domain.Edition, error) {
			return auth.EditionAuth(ctx, token.GetAccessToken())
		}},
		{"game", func() (*domain.LauncherUser, *domain.Edition, error) {
			return auth.EditionGameAuth(ctx, token.GetAccessToken(), gameID)
		}},
		{"image", func() (*domain.LauncherUser, *domain.Edition, error) {
			return auth.EditionImageAuth(ctx, token.GetAccessToken(), imageID)
		}},
		{"video", func() (*domain.LauncherUser, *domain.Edition, error) {
			return auth.EditionVideoAuth(ctx, token.GetAccessToken(), videoID)
		}},
		{"file", func() (*domain.LauncherUser, *domain.Edition, error) {
			return auth.EditionFileAuth(ctx, token.GetAccessToken(), fileID)
		}},
	}
	for _, check := range checks {
		t.Run("active/"+check.name, func(t *testing.T) {
			key, edition, err := check.run()
			require.NoError(t, err)
			require.Equal(t, values.LauncherUserStatusActive, key.GetStatus())
			require.Equal(t, editionID, edition.GetID())
		})
	}

	// Both status rows remain selectable; revocation changes the product key's status ID.
	require.NoError(t, db.db.Exec("INSERT INTO product_key_statuses (id, name, active) VALUES (?, ?, ?)", 2, schema.ProductKeyStatusInactive, true).Error)
	info, err := NewAccessToken(db).GetAccessTokenInfo(ctx, token.GetAccessToken(), repository.LockTypeNone)
	require.NoError(t, err)
	_, err = auth.RevokeProductKey(ctx, info.ProductKey.GetID())
	require.NoError(t, err)
	require.False(t, token.IsExpired())
	for _, check := range checks {
		t.Run("revoked/"+check.name, func(t *testing.T) {
			key, edition, err := check.run()
			require.ErrorIs(t, err, service.ErrInvalidAccessToken)
			require.Nil(t, key)
			require.Nil(t, edition)
		})
	}
}
