package gorm2

import (
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
)

func newAccessTokenFixture(t *testing.T, statusName string, statusRowActive bool) (*DB, *domain.LauncherSession, values.EditionID) {
	t.Helper()
	tx := testDB.db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	repositoryDB := &DB{db: tx}
	db, err := repositoryDB.getDB(t.Context())
	require.NoError(t, err)

	var status schema.ProductKeyStatusTable
	require.NoError(t, db.Where("name = ?", statusName).
		FirstOrCreate(&status, schema.ProductKeyStatusTable{Name: statusName}).Error)

	require.NoError(t, db.Model(&status).Update("active", statusRowActive).Error)

	now := time.Now()
	editionID := values.NewEditionID()
	keyID := values.NewLauncherUserID()
	require.NoError(t, db.Create(&schema.EditionTable{
		ID: uuid.UUID(editionID), Name: "synthetic edition", CreatedAt: now,
	}).Error)
	productKey, err := values.NewLauncherUserProductKey()
	require.NoError(t, err)
	require.NoError(t, db.Create(&schema.ProductKeyTable{
		ID: uuid.UUID(keyID), EditionID: uuid.UUID(editionID),
		ProductKey: string(productKey), StatusID: status.ID, CreatedAt: now,
	}).Error)

	accessToken, err := values.NewLauncherSessionAccessToken()
	require.NoError(t, err)
	token := domain.NewLauncherSession(
		values.NewLauncherSessionID(),
		accessToken,
		now.Add(time.Hour),
	)
	require.NoError(t, NewAccessToken(repositoryDB).SaveAccessToken(t.Context(), keyID, token))
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
			db, token, editionID := newAccessTokenFixture(t, tc.status, tc.rowActive)
			info, err := NewAccessToken(db).GetAccessTokenInfo(t.Context(), token.GetAccessToken(), repository.LockTypeNone)
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
	ctx := t.Context()
	db, token, editionID := newAccessTokenFixture(t, schema.ProductKeyStatusActive, true)
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
