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

func TestSaveGameImageV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository := NewGameImageV2(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		image        *domain.GameImage
		beforeImages []migrate.GameImageTable2
		expectImages []migrate.GameImageTable2
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

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

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
			beforeImages: []migrate.GameImageTable2{},
			expectImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
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
			beforeImages: []migrate.GameImageTable2{},
			expectImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
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
			beforeImages: []migrate.GameImageTable2{},
			expectImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeGif],
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
			beforeImages: []migrate.GameImageTable2{},
			expectImages: []migrate.GameImageTable2{},
			isErr:        true,
		},
		{
			description: "既に画像が存在しても問題なし",
			gameID:      gameID5,
			image: domain.NewGameImage(
				imageID5,
				values.GameImageTypeJpeg,
				now,
			),
			beforeImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
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
			beforeImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImages: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID8),
					GameID:      uuid.UUID(gameID6),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable2{
				ID:               uuid.UUID(testCase.gameID),
				Name:             "test",
				Description:      "test",
				CreatedAt:        time.Now(),
				GameImage2s:      testCase.beforeImages,
				VisibilityTypeID: gameVisibilityTypeIDPublic,
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

			var images []migrate.GameImageTable2
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&images).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, images, len(testCase.expectImages))

			imageMap := make(map[uuid.UUID]migrate.GameImageTable2)
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

func TestGetGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository := NewGameImageV2(testDB)

	type test struct {
		description string
		imageID     values.GameImageID
		lockType    repository.LockType
		images      []migrate.GameImageTable2
		expectImage repository.GameImageInfo
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()

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

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			imageID:     imageID1,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImage: repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID1,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID1,
			},
		},
		{
			description: "pngでも問題なし",
			imageID:     imageID2,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now,
				},
			},
			expectImage: repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID2,
					values.GameImageTypePng,
					now,
				),
				GameID: gameID2,
			},
		},
		{
			description: "gifでも問題なし",
			imageID:     imageID3,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeGif],
					CreatedAt:   now,
				},
			},
			expectImage: repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID3,
					values.GameImageTypeGif,
					now,
				),
				GameID: gameID3,
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			imageID:     imageID4,
			lockType:    repository.LockTypeRecord,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID4),
					GameID:      uuid.UUID(gameID4),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImage: repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID4,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID4,
			},
		},
		{
			description: "複数の画像があっても問題なし",
			imageID:     imageID5,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImage: repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID5,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID5,
			},
		},
		{
			description: "画像が存在しないのでRecordNotFound",
			imageID:     imageID7,
			lockType:    repository.LockTypeNone,
			images:      []migrate.GameImageTable2{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, image := range testCase.images {
				if game, ok := gameIDMap[image.GameID]; ok {
					game.GameImage2s = append(game.GameImage2s, image)
				} else {
					gameIDMap[image.GameID] = &migrate.GameTable2{
						ID:               image.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameImage2s:      []migrate.GameImageTable2{image},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			image, err := gameImageRepository.GetGameImage(ctx, testCase.imageID, testCase.lockType)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.expectImage.GetID(), image.GetID())
			assert.Equal(t, testCase.expectImage.GetType(), image.GetType())
			assert.WithinDuration(t, testCase.expectImage.GetCreatedAt(), image.GetCreatedAt(), time.Second)
			assert.Equal(t, testCase.expectImage.GameID, image.GameID)
		})
	}
}

func TestGetGameImages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameImageRepository := NewGameImageV2(testDB)

	type test struct {
		description  string
		gameID       values.GameID
		lockType     repository.LockType
		images       []migrate.GameImageTable2
		expectImages []*domain.GameImage
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

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID1),
					GameID:      uuid.UUID(gameID1),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImages: []*domain.GameImage{
				domain.NewGameImage(
					imageID1,
					values.GameImageTypeJpeg,
					now,
				),
			},
		},
		{
			description: "pngでも問題なし",
			gameID:      gameID2,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID2),
					GameID:      uuid.UUID(gameID2),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now,
				},
			},
			expectImages: []*domain.GameImage{
				domain.NewGameImage(
					imageID2,
					values.GameImageTypePng,
					now,
				),
			},
		},
		{
			description: "gifでも問題なし",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID3),
					GameID:      uuid.UUID(gameID3),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeGif],
					CreatedAt:   now,
				},
			},
			expectImages: []*domain.GameImage{
				domain.NewGameImage(
					imageID3,
					values.GameImageTypeGif,
					now,
				),
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			gameID:      gameID4,
			lockType:    repository.LockTypeRecord,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID4),
					GameID:      uuid.UUID(gameID4),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
			},
			expectImages: []*domain.GameImage{
				domain.NewGameImage(
					imageID4,
					values.GameImageTypeJpeg,
					now,
				),
			},
		},
		{
			description: "複数の画像があっても問題なし",
			gameID:      gameID5,
			images: []migrate.GameImageTable2{
				{
					ID:          uuid.UUID(imageID5),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypeJpeg],
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(imageID6),
					GameID:      uuid.UUID(gameID5),
					ImageTypeID: imageTypeMap[migrate.GameImageTypePng],
					CreatedAt:   now.Add(-10 * time.Hour),
				},
			},
			expectImages: []*domain.GameImage{
				domain.NewGameImage(
					imageID5,
					values.GameImageTypeJpeg,
					now,
				),
				domain.NewGameImage(
					imageID6,
					values.GameImageTypePng,
					now.Add(-10*time.Hour),
				),
			},
		},
		{
			description:  "画像が存在しなくても問題なし",
			gameID:       gameID6,
			lockType:     repository.LockTypeNone,
			images:       []migrate.GameImageTable2{},
			expectImages: []*domain.GameImage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, image := range testCase.images {
				if game, ok := gameIDMap[image.GameID]; ok {
					game.GameImage2s = append(game.GameImage2s, image)
				} else {
					gameIDMap[image.GameID] = &migrate.GameTable2{
						ID:               image.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameImage2s:      []migrate.GameImageTable2{image},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			images, err := gameImageRepository.GetGameImages(ctx, testCase.gameID, testCase.lockType)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			for i, expectImage := range testCase.expectImages {
				assert.Equal(t, expectImage.GetID(), images[i].GetID())
				assert.Equal(t, expectImage.GetType(), images[i].GetType())
				assert.WithinDuration(t, expectImage.GetCreatedAt(), images[i].GetCreatedAt(), time.Second)
			}
		})
	}
}
