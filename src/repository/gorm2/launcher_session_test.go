package gorm2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

func TestCreateLauncherSession(t *testing.T) {
	t.Parallel()

	launcherSesionRepository := NewLauncherSession(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	productKey, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	accessToken1, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	accessToken2, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	editionID := values.NewEditionID()
	launcherUserID := values.NewLauncherUserID()
	dbEdition := schema.LauncherVersionTable{
		ID:        uuid.UUID(editionID),
		Name:      "TestCreateLauncherSession",
		CreatedAt: time.Now(),
		LauncherUsers: []schema.LauncherUserTable{
			{
				ID:         uuid.UUID(launcherUserID),
				ProductKey: string(productKey),
				CreatedAt:  time.Now(),
			},
		},
	}

	err = db.Create(&dbEdition).Error
	if err != nil {
		t.Errorf("failed to create launcher version: %v", err)
	}

	type test struct {
		description     string
		launcherUserID  values.LauncherUserID
		launcherSession *domain.LauncherSession
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:    "入出力問題ないのでエラーなし",
			launcherUserID: launcherUserID,
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken1,
				time.Now().Add(time.Hour),
			),
		},
		{
			description:    "同一のアクセストークンが存在するのでエラー",
			launcherUserID: launcherUserID,
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken1,
				time.Now().Add(time.Hour),
			),
			isErr: true,
		},
		{
			description:    "ユーザーIDが存在しないのでエラー",
			launcherUserID: values.NewLauncherUserID(),
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken2,
				time.Now().Add(time.Hour),
			),
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherSession, err := launcherSesionRepository.CreateLauncherSession(ctx, testCase.launcherUserID, testCase.launcherSession)

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

			assert.Equal(t, testCase.launcherSession, launcherSession)

			var dbLauncherSession schema.LauncherSessionTable
			err = db.
				Where("id = ?", uuid.UUID(testCase.launcherSession.GetID())).
				Take(&dbLauncherSession).Error
			if err != nil {
				t.Errorf("failed to get launcher session: %v", err)
			}

			assert.Equal(t, uuid.UUID(testCase.launcherSession.GetID()), dbLauncherSession.ID)
			assert.Equal(t, uuid.UUID(testCase.launcherUserID), dbLauncherSession.LauncherUserID)
			assert.Equal(t, string(testCase.launcherSession.GetAccessToken()), dbLauncherSession.AccessToken)
			assert.WithinDuration(t, testCase.launcherSession.GetExpiresAt(), dbLauncherSession.ExpiresAt, time.Second)
		})
	}
}
