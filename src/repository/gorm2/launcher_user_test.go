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
	"github.com/traPtitech/trap-collection-server/src/repository"
)

func TestCreateLauncherUsers(t *testing.T) {
	t.Parallel()

	launcherUserRepository := NewLauncherUser(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	productKey1, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey2, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey3, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey4, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	productKey5, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	launcherVersionID := values.NewLauncherVersionID()
	dbLauncherVersion := LauncherVersionTable{
		ID:        uuid.UUID(launcherVersionID),
		Name:      "TestCreateLauncherUsers",
		CreatedAt: time.Now(),
	}

	err = db.Create(&dbLauncherVersion).Error
	if err != nil {
		t.Errorf("failed to create launcher version: %v", err)
	}

	type test struct {
		description       string
		launcherVersionID values.LauncherVersionID
		launcherUsers     []*domain.LauncherUser
		isErr             bool
		err               error
	}

	testCases := []test{
		{
			description:       "入出力問題ないのでエラーなし",
			launcherVersionID: launcherVersionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey1,
				),
			},
		},
		{
			description:       "ユーザーが空でもエラーなし",
			launcherVersionID: launcherVersionID,
			launcherUsers:     []*domain.LauncherUser{},
		},
		{
			description:       "ユーザーが複数人でもエラーなし",
			launcherVersionID: launcherVersionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey2,
				),
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey3,
				),
			},
		},
		{
			description:       "プロダクトキーが同一なのでエラー",
			launcherVersionID: launcherVersionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey4,
				),
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey4,
				),
			},
			isErr: true,
		},
		{
			description:       "ランチャーバージョンが存在しないのでエラー",
			launcherVersionID: values.NewLauncherVersionID(),
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey5,
				),
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherUsers, err := launcherUserRepository.CreateLauncherUsers(ctx, testCase.launcherVersionID, testCase.launcherUsers)

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

			assert.ElementsMatch(t, testCase.launcherUsers, launcherUsers)
		})
	}
}

func TestDeleteLauncherUser(t *testing.T) {
	t.Parallel()

	launcherUserRepository := NewLauncherUser(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	launcherVersionID := values.NewLauncherVersionID()
	dbLauncherVersion := LauncherVersionTable{
		ID:        uuid.UUID(launcherVersionID),
		Name:      "TestDeleteLauncherUser",
		CreatedAt: time.Now(),
	}

	err = db.Create(&dbLauncherVersion).Error
	if err != nil {
		t.Errorf("failed to create launcher version: %v", err)
	}

	type test struct {
		description         string
		validLauncherUserID bool
		isErr               bool
		err                 error
	}

	testCases := []test{
		{
			description:         "ユーザーが存在するのでエラーなし",
			validLauncherUserID: true,
		},
		{
			description:         "ユーザーが存在しないのでエラー",
			validLauncherUserID: false,
			isErr:               true,
			err:                 repository.ErrNoRecordDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherUserID := values.NewLauncherUserID()
			if testCase.validLauncherUserID {
				productKey, err := values.NewLauncherUserProductKey()
				if err != nil {
					t.Errorf("failed to create product key: %v", err)
				}

				dbLauncherUser := LauncherUserTable{
					ID:                uuid.UUID(launcherUserID),
					ProductKey:        string(productKey),
					LauncherVersionID: uuid.UUID(launcherVersionID),
				}
				err = db.Create(&dbLauncherUser).Error
				if err != nil {
					t.Errorf("failed to create launcher user: %v", err)
				}
			}

			err := launcherUserRepository.DeleteLauncherUser(ctx, launcherUserID)

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
