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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestSaveProductKeys(t *testing.T) {
	productKeyRepository := NewProductKey(testDB)

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description       string
		editionID         values.LauncherVersionID
		productKeys       []*domain.LauncherUser
		beforeProductKeys []migrate.ProductKeyTable2
		expectProductKeys []migrate.ProductKeyTable2
		isErr             bool
		err               error
	}

	editionID1 := values.NewLauncherVersionID()
	editionID2 := values.NewLauncherVersionID()
	editionID3 := values.NewLauncherVersionID()
	editionID4 := values.NewLauncherVersionID()
	editionID5 := values.NewLauncherVersionID()
	editionID6 := values.NewLauncherVersionID()
	editionID7 := values.NewLauncherVersionID()

	productKeyID1 := values.NewLauncherUserID()
	productKeyID2 := values.NewLauncherUserID()
	productKeyID3 := values.NewLauncherUserID()
	productKeyID4 := values.NewLauncherUserID()
	productKeyID5 := values.NewLauncherUserID()
	productKeyID6 := values.NewLauncherUserID()
	productKeyID7 := values.NewLauncherUserID()
	productKeyID8 := values.NewLauncherUserID()
	productKeyID9 := values.NewLauncherUserID()
	productKeyID10 := values.NewLauncherUserID()
	productKeyID11 := values.NewLauncherUserID()
	productKeyID12 := values.NewLauncherUserID()
	productKeyID13 := values.NewLauncherUserID()

	var status []*migrate.ProductKeyStatusTable2
	err = db.
		Session(&gorm.Session{}).
		Find(&status).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	statusMap := make(map[string]int, len(status))
	for _, statusValue := range status {
		statusMap[statusValue.Name] = statusValue.ID
	}

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			editionID:   editionID1,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID1,
					values.NewLauncherUserProductKeyFromString("xxxxx-xxxxx-xxxxx-xxxxx-xxxxx"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID1),
					EditionID:  uuid.UUID(editionID1),
					ProductKey: "xxxxx-xxxxx-xxxxx-xxxxx-xxxxx",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now,
				},
			},
		},
		{
			description: "statusがinactiveでもエラーなし",
			editionID:   editionID2,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID2,
					values.NewLauncherUserProductKeyFromString("yyyyy-yyyyy-yyyyy-yyyyy-yyyyy"),
					values.LauncherUserStatusInactive,
					now,
				),
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID2),
					EditionID:  uuid.UUID(editionID2),
					ProductKey: "yyyyy-yyyyy-yyyyy-yyyyy-yyyyy",
					StatusID:   statusMap[migrate.ProductKeyStatusInactive],
					CreatedAt:  now,
				},
			},
		},
		{
			description: "既にproduct keyが存在してもエラーなし",
			editionID:   editionID3,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID3,
					values.NewLauncherUserProductKeyFromString("zzzzz-zzzzz-zzzzz-zzzzz-zzzzz"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			beforeProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID4),
					EditionID:  uuid.UUID(editionID3),
					ProductKey: "vvvvv-vvvvv-vvvvv-vvvvv-vvvvv",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID4),
					EditionID:  uuid.UUID(editionID3),
					ProductKey: "vvvvv-vvvvv-vvvvv-vvvvv-vvvvv",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
				{
					ID:         uuid.UUID(productKeyID3),
					EditionID:  uuid.UUID(editionID3),
					ProductKey: "zzzzz-zzzzz-zzzzz-zzzzz-zzzzz",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now,
				},
			},
		},
		{
			description: "複数product keyが存在してもエラーなし",
			editionID:   editionID4,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID5,
					values.NewLauncherUserProductKeyFromString("aaaaa-aaaaa-aaaaa-aaaaa-aaaaa"),
					values.LauncherUserStatusActive,
					now,
				),
				domain.NewProductKey(
					productKeyID6,
					values.NewLauncherUserProductKeyFromString("bbbbb-bbbbb-bbbbb-bbbbb-bbbbb"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID5),
					EditionID:  uuid.UUID(editionID4),
					ProductKey: "aaaaa-aaaaa-aaaaa-aaaaa-aaaaa",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now,
				},
				{
					ID:         uuid.UUID(productKeyID6),
					EditionID:  uuid.UUID(editionID4),
					ProductKey: "bbbbb-bbbbb-bbbbb-bbbbb-bbbbb",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now,
				},
			},
		},
		{
			description: "既に同一のproduct keyが存在するのでエラー",
			editionID:   editionID5,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID7,
					values.NewLauncherUserProductKeyFromString("wwwww-wwwww-wwwww-wwwww-wwwww"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			beforeProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID8),
					EditionID:  uuid.UUID(editionID5),
					ProductKey: "wwwww-wwwww-wwwww-wwwww-wwwww",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID8),
					EditionID:  uuid.UUID(editionID5),
					ProductKey: "wwwww-wwwww-wwwww-wwwww-wwwww",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			isErr: true,
		},
		{
			description: "既に同一のproduct keyが存在するので全てのproduct keyが登録されない",
			editionID:   editionID6,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID9,
					values.NewLauncherUserProductKeyFromString("ccccc-ccccc-ccccc-ccccc-ccccc"),
					values.LauncherUserStatusActive,
					now,
				),
				domain.NewProductKey(
					productKeyID10,
					values.NewLauncherUserProductKeyFromString("ddddd-ddddd-ddddd-ddddd-ddddd"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			beforeProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID11),
					EditionID:  uuid.UUID(editionID6),
					ProductKey: "ddddd-ddddd-ddddd-ddddd-ddddd",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			expectProductKeys: []migrate.ProductKeyTable2{
				{
					ID:         uuid.UUID(productKeyID11),
					EditionID:  uuid.UUID(editionID6),
					ProductKey: "ddddd-ddddd-ddddd-ddddd-ddddd",
					StatusID:   statusMap[migrate.ProductKeyStatusActive],
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			isErr: true,
		},
		{
			description: "同一のproduct keyを含むのでエラー",
			editionID:   editionID7,
			productKeys: []*domain.LauncherUser{
				domain.NewProductKey(
					productKeyID12,
					values.NewLauncherUserProductKeyFromString("ccccc-ccccc-ccccc-ccccc-ccccc"),
					values.LauncherUserStatusActive,
					now,
				),
				domain.NewProductKey(
					productKeyID13,
					values.NewLauncherUserProductKeyFromString("ccccc-ccccc-ccccc-ccccc-ccccc"),
					values.LauncherUserStatusActive,
					now,
				),
			},
			expectProductKeys: []migrate.ProductKeyTable2{},
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Where("edition_id = ?", uuid.UUID(testCase.editionID)).
					Delete(&migrate.ProductKeyTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete table: %v", err)
				}

				err = db.
					Unscoped().
					Where("id = ?", uuid.UUID(testCase.editionID)).
					Delete(&migrate.EditionTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete table: %v", err)
				}
			}()
			err := db.Create(&migrate.EditionTable2{
				ID:          uuid.UUID(testCase.editionID),
				Name:        "test",
				CreatedAt:   now,
				ProductKeys: testCase.beforeProductKeys,
			}).Error
			if err != nil {
				t.Fatalf("failed to create table: %v", err)
			}

			err = productKeyRepository.SaveProductKeys(ctx, testCase.editionID, testCase.productKeys)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var productKeys []migrate.ProductKeyTable2
			err = db.
				Where("edition_id = ?", uuid.UUID(testCase.editionID)).
				Find(&productKeys).Error
			if err != nil {
				t.Fatalf("failed to find table: %v", err)
			}

			assert.Len(t, productKeys, len(testCase.expectProductKeys))

			productKeyMap := make(map[uuid.UUID]migrate.ProductKeyTable2)
			for _, productKey := range productKeys {
				productKeyMap[productKey.ID] = productKey
			}

			for _, expectProductKey := range testCase.expectProductKeys {
				actualProductKey, ok := productKeyMap[expectProductKey.ID]
				if !ok {
					t.Errorf("not found product key: %v", expectProductKey.ID)
				}

				assert.Equal(t, expectProductKey.ID, actualProductKey.ID)
				assert.Equal(t, expectProductKey.EditionID, actualProductKey.EditionID)
				assert.Equal(t, expectProductKey.ProductKey, actualProductKey.ProductKey)
				assert.Equal(t, expectProductKey.StatusID, actualProductKey.StatusID)
				assert.WithinDuration(t, expectProductKey.CreatedAt, actualProductKey.CreatedAt, 2*time.Second)
			}
		})
	}
}
