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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
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

	editionID := values.NewEditionID()
	dbLauncherVersion := schema.LauncherVersionTable{
		ID:        uuid.UUID(editionID),
		Name:      "TestCreateLauncherUsers",
		CreatedAt: time.Now(),
	}

	err = db.Create(&dbLauncherVersion).Error
	if err != nil {
		t.Errorf("failed to create launcher version: %v", err)
	}

	type test struct {
		description   string
		editionID     values.EditionID
		launcherUsers []*domain.LauncherUser
		isErr         bool
		err           error
	}

	testCases := []test{
		{
			description: "入出力問題ないのでエラーなし",
			editionID:   editionID,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey1,
				),
			},
		},
		{
			description:   "ユーザーが空でもエラーなし",
			editionID:     editionID,
			launcherUsers: []*domain.LauncherUser{},
		},
		{
			description: "ユーザーが複数人でもエラーなし",
			editionID:   editionID,
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
			description: "プロダクトキーが同一なのでエラー",
			editionID:   editionID,
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
			description: "ランチャーバージョンが存在しないのでエラー",
			editionID:   values.NewEditionID(),
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
			launcherUsers, err := launcherUserRepository.CreateLauncherUsers(ctx, testCase.editionID, testCase.launcherUsers)

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

	editionID := values.NewEditionID()
	dbEdition := schema.LauncherVersionTable{
		ID:        uuid.UUID(editionID),
		Name:      "TestDeleteLauncherUser",
		CreatedAt: time.Now(),
	}

	err = db.Create(&dbEdition).Error
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

				dbLauncherUser := schema.LauncherUserTable{
					ID:                uuid.UUID(launcherUserID),
					ProductKey:        string(productKey),
					LauncherVersionID: uuid.UUID(editionID),
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

func TestGetLauncherUserByProductKey(t *testing.T) {
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

	editionID := values.NewEditionID()
	launcherUserID := values.NewLauncherUserID()
	launcherUser := domain.NewLauncherUser(
		launcherUserID,
		productKey1,
	)
	dbLauncherVersion := schema.LauncherVersionTable{
		ID:        uuid.UUID(editionID),
		Name:      "TestGetLauncherUserByProductKey",
		CreatedAt: time.Now(),
		LauncherUsers: []schema.LauncherUserTable{
			{
				ID:         uuid.UUID(launcherUserID),
				ProductKey: string(productKey1),
				CreatedAt:  time.Now(),
			},
		},
	}

	err = db.Create(&dbLauncherVersion).Error
	if err != nil {
		t.Errorf("failed to create launcher version: %v", err)
	}

	type test struct {
		description  string
		productKey   values.LauncherUserProductKey
		launcherUser *domain.LauncherUser
		isErr        bool
		err          error
	}

	testCases := []test{
		{
			description:  "ユーザーが存在するのでエラーなし",
			productKey:   productKey1,
			launcherUser: launcherUser,
		},
		{
			description: "ユーザーが存在しないのでエラー",
			productKey:  productKey2,
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherUser, err := launcherUserRepository.GetLauncherUserByProductKey(ctx, testCase.productKey)

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

			assert.Equal(t, *testCase.launcherUser, *launcherUser)
		})
	}
}
