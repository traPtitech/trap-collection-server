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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository := NewGameImage(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		image        *domain.GameImage
		beforeImages []migrate.GameImageTable
		expectImages []migrate.GameImageTable
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

	var imageTypes []*migrate.GameImageTypeTable
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

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			image: domain.NewGameImage(
				imageID1,
				values.GameImageTypeJpeg,
				now,
			),
			beforeImages: []migrate.GameImageTable{},
			expectImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "pngでも問題なし",
			gameID:      gameID2,
			image: domain.NewGameImage(
				imageID2,
				values.GameImageTypePng,
				now,
			),
			beforeImages: []migrate.GameImageTable{},
			expectImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "gifでも問題なし",
			gameID:      gameID3,
			image: domain.NewGameImage(
				imageID3,
				values.GameImageTypeGif,
				now,
			),
			beforeImages: []migrate.GameImageTable{},
			expectImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[gameImageTypeGif],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "想定外の画像の種類なのでエラー",
			gameID:      gameID4,
			image: domain.NewGameImage(
				imageID4,
				100,
				now,
			),
			beforeImages: []migrate.GameImageTable{},
			expectImages: []migrate.GameImageTable{},
			isErr:        true,
		},
		{
			description: "既に画像画像が存在しても問題なし",
			gameID:      gameID5,
			image: domain.NewGameImage(
				imageID5,
				values.GameImageTypeJpeg,
				now,
			),
			beforeImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
		},
		{
			description: "エラーの場合変更なし",
			gameID:      gameID6,
			image: domain.NewGameImage(
				imageID7,
				100,
				now,
			),
			beforeImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImages: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
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

			var images []migrate.GameImageTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&images).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, images, len(testCase.expectImages))

			imageMap := make(map[uuid.UUID]migrate.GameImageTable)
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

	gameImageRepository := NewGameImage(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		images      []migrate.GameImageTable
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

	var imageTypes []*migrate.GameImageTypeTable
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

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImage: domain.NewGameImage(
				imageID1,
				values.GameImageTypeJpeg,
				now,
			),
		},
		{
			description: "pngでもエラーなし",
			gameID:      gameID2,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now,
				},
			},
			expectImage: domain.NewGameImage(
				imageID2,
				values.GameImageTypePng,
				now,
			),
		},
		{
			description: "gifでもエラーなし",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[gameImageTypeGif],
					CreatedAt:   now,
				},
			},
			expectImage: domain.NewGameImage(
				imageID3,
				values.GameImageTypeGif,
				now,
			),
		},
		{
			description: "画像がないのでエラー",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			images:      []migrate.GameImageTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "複数でも正しい画像が返る",
			gameID:      gameID5,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID4),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   now.Add(-24 * time.Hour),
				},
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[gameImageTypePng],
					CreatedAt:   now,
				},
			},
			expectImage: domain.NewGameImage(
				imageID5,
				values.GameImageTypePng,
				time.Now(),
			),
		},
		{
			description: "行ロックをとっても問題なし",
			gameID:      gameID6,
			lockType:    repository.LockTypeRecord,
			images: []migrate.GameImageTable{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[gameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImage: domain.NewGameImage(
				imageID6,
				values.GameImageTypeJpeg,
				now,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
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

			assert.Equal(t, testCase.expectImage.GetID(), image.GetID())
			assert.Equal(t, testCase.expectImage.GetType(), image.GetType())
			assert.WithinDuration(t, testCase.expectImage.GetCreatedAt(), image.GetCreatedAt(), time.Second)
		})
	}
}
