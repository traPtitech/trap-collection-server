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
	"gorm.io/gorm"
)

func TestSetupImageTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description      string
		beforeImageTypes []string
		isErr            bool
		err              error
	}

	testCases := []test{
		{
			description:      "何も存在しない場合問題なし",
			beforeImageTypes: []string{},
		},
		{
			description: "1つのみ存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
			},
		},
		{
			description: "2つ存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
				gameImageTypePng,
			},
		},
		{
			description: "全て存在する場合問題なし",
			beforeImageTypes: []string{
				gameImageTypeJpeg,
				gameImageTypePng,
				gameImageTypeGif,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&GameImageTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeImageTypes) != 0 {
				imageTypes := make([]*GameImageTypeTable, 0, len(testCase.beforeImageTypes))
				for _, imageType := range testCase.beforeImageTypes {
					imageTypes = append(imageTypes, &GameImageTypeTable{
						Name: imageType,
					})
				}

				err := db.Create(imageTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupImageTypeTable(db)

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

			var imageTypes []*GameImageTypeTable
			err = db.
				Select("name").
				Find(&imageTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			imageTypeNames := make([]string, 0, len(imageTypes))
			for _, imageType := range imageTypes {
				imageTypeNames = append(imageTypeNames, imageType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameImageTypeJpeg,
				gameImageTypePng,
				gameImageTypeGif,
			}, imageTypeNames)
		})
	}
}

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository, err := NewGameImage(testDB)
	if err != nil {
		t.Fatalf("failed to create game management role repository: %+v\n", err)
	}

	type test struct {
		description  string
		gameID       values.GameID
		image        *domain.GameImage
		beforeImages []GameImageTable
		expectImages []GameImageTable
		isErr        bool
		err          error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()
	imageID8 := values.NewGameImageID()

	var imageTypes []*GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&imageTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	imageTypeMap := make(map[string]int, len(imageTypes))
	for _, imageType := range imageTypes {
		imageTypeMap[imageType.Name] = imageType.ID
	}

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			image: domain.NewGameImage(
				imageID1,
				values.GameImageTypeJpeg,
			),
			beforeImages: []GameImageTable{},
			expectImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "pngでも問題なし",
			gameID:      gameID2,
			image: domain.NewGameImage(
				imageID2,
				values.GameImageTypePng,
			),
			beforeImages: []GameImageTable{},
			expectImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "gifでも問題なし",
			gameID:      gameID3,
			image: domain.NewGameImage(
				imageID3,
				values.GameImageTypeGif,
			),
			beforeImages: []GameImageTable{},
			expectImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[gameImageTypeGif],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "想定外の画像の種類なのでエラー",
			gameID:      gameID4,
			image: domain.NewGameImage(
				imageID4,
				100,
			),
			beforeImages: []GameImageTable{},
			expectImages: []GameImageTable{},
			isErr:        true,
		},
		{
			description: "既に画像画像が存在しても問題なし",
			gameID:      gameID5,
			image: domain.NewGameImage(
				imageID5,
				values.GameImageTypeJpeg,
			),
			beforeImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			expectImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   time.Now(),
				},
			},
		},
		{
			description: "エラーの場合変更なし",
			gameID:      gameID6,
			image: domain.NewGameImage(
				imageID7,
				100,
			),
			beforeImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			expectImages: []GameImageTable{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now().Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameImages:  testCase.beforeImages,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameImageRepository.SaveGameImage(ctx, testCase.gameID, testCase.image)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var images []GameImageTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&images).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, images, len(testCase.expectImages))

			imageMap := make(map[uuid.UUID]GameImageTable)
			for _, image := range images {
				imageMap[image.ID] = image
			}

			for _, expectImage := range testCase.expectImages {
				actualImage, ok := imageMap[expectImage.ID]
				if !ok {
					t.Errorf("not found image: %+v", expectImage)
				}

				assert.Equal(t, expectImage.GameID, actualImage.GameID)
				assert.Equal(t, expectImage.ImageTypeID, actualImage.ImageTypeID)
				assert.WithinDuration(t, expectImage.CreatedAt, actualImage.CreatedAt, 2*time.Second)
			}
		})
	}
}

func TestGetLatestGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository, err := NewGameImage(testDB)
	if err != nil {
		t.Fatalf("failed to create game management role repository: %+v\n", err)
	}

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		images      []GameImageTable
		expectImage *domain.GameImage
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()

	var imageTypes []*GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&imageTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	imageTypeMap := make(map[string]int, len(imageTypes))
	for _, imageType := range imageTypes {
		imageTypeMap[imageType.Name] = imageType.ID
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			images: []GameImageTable{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   time.Now(),
				},
			},
			expectImage: domain.NewGameImage(
				imageID1,
				values.GameImageTypeJpeg,
			),
		},
		{
			description: "pngでもエラーなし",
			gameID:      gameID2,
			lockType:    repository.LockTypeNone,
			images: []GameImageTable{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now(),
				},
			},
			expectImage: domain.NewGameImage(
				imageID2,
				values.GameImageTypePng,
			),
		},
		{
			description: "gifでもエラーなし",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			images: []GameImageTable{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[gameImageTypeGif],
					CreatedAt:   time.Now(),
				},
			},
			expectImage: domain.NewGameImage(
				imageID3,
				values.GameImageTypeGif,
			),
		},
		{
			description: "画像がないのでエラー",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			images:      []GameImageTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "複数でも正しい画像が返る",
			gameID:      gameID5,
			lockType:    repository.LockTypeNone,
			images: []GameImageTable{
				{
					ID:          uuid.UUID(imageID4),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   time.Now().Add(-24 * time.Hour),
				},
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   time.Now(),
				},
			},
			expectImage: domain.NewGameImage(
				imageID5,
				values.GameImageTypePng,
			),
		},
		{
			description: "行ロックをとっても問題なし",
			gameID:      gameID6,
			lockType:    repository.LockTypeRecord,
			images: []GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   time.Now(),
				},
			},
			expectImage: domain.NewGameImage(
				imageID6,
				values.GameImageTypeJpeg,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:          uuid.UUID(testCase.gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameImages:  testCase.images,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			image, err := gameImageRepository.GetLatestGameImage(ctx, testCase.gameID, testCase.lockType)

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

			assert.Equal(t, *testCase.expectImage, *image)
		})
	}
}
