package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetGenresByGameID(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		gameID             values.GameID
		gameGenres         []migrate.GameGenreTable
		expectedGameGenres []*domain.GameGenre
		isErr              bool
		expectedErr        error
	}

	now := time.Now()

	gameID1 := values.NewGameID()
	genreID1 := values.NewGameGenreID()
	gameID2 := values.NewGameID()
	genreID2 := values.NewGameGenreID()
	gameID3 := values.NewGameID()
	genreID3 := values.NewGameGenreID()
	gameID4 := values.NewGameID()
	// genreID4 := values.NewGameGenreID()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameID: gameID1,
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID1),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test",
							Description: "test",
						},
					},
				},
			},
			expectedGameGenres: []*domain.GameGenre{
				domain.NewGameGenre(genreID1, "test", now.Add(-time.Hour)),
			},
		},
		"複数のジャンルがあってもエラーなし": {
			gameID: gameID2,
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID2),
					Name:      "test1",
					CreatedAt: now.Add(-time.Hour),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID2),
							Name:        "test2",
							Description: "test2",
						},
					},
				},
				{
					ID:        uuid.UUID(genreID3),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID2),
							Name:        "test2",
							Description: "test2",
						},
					},
				},
			},
			expectedGameGenres: []*domain.GameGenre{
				domain.NewGameGenre(genreID2, "test1", now.Add(-time.Hour)),
				domain.NewGameGenre(genreID3, "test2", now.Add(-time.Hour*2)),
			},
		},
		"ジャンルが空でもエラーなし": {
			gameID:             gameID3,
			gameGenres:         []migrate.GameGenreTable{},
			expectedGameGenres: []*domain.GameGenre{},
		},
		"違うゲームがあってもエラーなし": {
			gameID: gameID4,
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID1),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test",
							Description: "test",
						},
					},
				},
			},
			expectedGameGenres: []*domain.GameGenre{},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer func() {
				_db := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					})

				var genres []migrate.GameGenreTable
				err := _db.Find(&genres).Error
				if err != nil {
					t.Fatalf("failed to get genres")
				}

				err = _db.
					Select("Games").
					Delete(&genres).Error
				if err != nil {
					t.Fatalf("failed to delete genres: %+v\n", err)
				}

				err = _db.Delete(&migrate.GameTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete games: %+v\n", err)
				}
			}()

			if testCase.gameGenres != nil && len(testCase.gameGenres) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.gameGenres).Error
				if err != nil {
					t.Fatalf("failed to create genre: %+v\n", err)
				}
			}

			gameGenres, err := gameGenreRepository.GetGenresByGameID(ctx, testCase.gameID)

			if testCase.isErr {
				if testCase.expectedErr == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Len(t, gameGenres, len(testCase.expectedGameGenres))
			for i, genre := range gameGenres {
				assert.Equal(t, testCase.expectedGameGenres[i].GetID(), genre.GetID())
				assert.Equal(t, testCase.expectedGameGenres[i].GetName(), genre.GetName())
				assert.WithinDuration(t, testCase.expectedGameGenres[i].GetCreatedAt(), genre.GetCreatedAt(), time.Second)
			}

		})
	}

}

func TestRemoveGameGenre(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		genreID          values.GameGenreID
		beforeGameGenres []migrate.GameGenreTable
		afterGameGenres  []migrate.GameGenreTable
		isErr            bool
		expectedErr      error
	}

	now := time.Now()

	genreID1 := values.NewGameGenreID()
	genreID2 := values.NewGameGenreID()
	genreID3 := values.NewGameGenreID()
	genreID4 := values.NewGameGenreID()
	genreID5 := values.NewGameGenreID()
	genreID6 := values.NewGameGenreID()
	genreID7 := values.NewGameGenreID()

	gameID1 := values.NewGameID()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			genreID: genreID1,
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID1),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
				},
			},
			afterGameGenres: []migrate.GameGenreTable{},
		},
		"該当するジャンルが存在しないのでErrNoRecordDeleted": {
			genreID: genreID2,
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID3),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID3),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
				},
			},
			isErr:       true,
			expectedErr: repository.ErrNoRecordDeleted,
		},
		"ジャンルが複数あっても問題なし": {
			genreID: genreID4,
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID4),
					Name:      "test1",
					CreatedAt: now.Add(-time.Hour),
				},
				{
					ID:        uuid.UUID(genreID5),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID5),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
				},
			},
		},
		"ゲームが紐づいていてもエラー無し": {
			genreID: genreID6,
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID6),
					Name:      "test",
					CreatedAt: now.Add(-time.Hour),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test",
							Description: "test",
							CreatedAt:   now.Add(-time.Hour),
						},
					},
				},
				{
					ID:        uuid.UUID(genreID7),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test",
							Description: "test",
							CreatedAt:   now.Add(-time.Hour),
						},
					},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID7),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
					Games: []migrate.GameTable2{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test",
							Description: "test",
							CreatedAt:   now.Add(-time.Hour),
						},
					},
				},
			},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			// 1個テストケースを実行したらテーブルの中身全部削除
			defer func() {
				_db := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					})

				var genres []migrate.GameGenreTable
				err := _db.Find(&genres).Error
				if err != nil {
					t.Fatalf("failed to get genres")
				}

				err = _db.
					Select("Games").
					Delete(&genres).Error
				if err != nil {
					t.Fatalf("failed to delete genres: %+v\n", err)
				}

				err = _db.Delete(&migrate.GameTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete games: %+v\n", err)
				}
			}()

			if testCase.beforeGameGenres != nil && len(testCase.beforeGameGenres) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeGameGenres).Error
				if err != nil {
					t.Fatalf("failed to create genre: %+v\n", err)
				}
			}

			err = gameGenreRepository.RemoveGameGenre(ctx, testCase.genreID)

			if testCase.isErr {
				if testCase.expectedErr == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			var genres []migrate.GameGenreTable
			err = db.
				Preload("Games").
				Find(&genres).Error
			if err != nil {
				t.Fatalf("failed to get genres: %+v", err)
			}

			assert.Len(t, genres, len(testCase.afterGameGenres))

			for i, genre := range genres {
				assert.Equal(t, testCase.afterGameGenres[i].ID, genre.ID)
				assert.Equal(t, testCase.afterGameGenres[i].Name, genre.Name)
				assert.WithinDuration(t, testCase.afterGameGenres[i].CreatedAt, genre.CreatedAt, time.Second)

				assert.Len(t, genre.Games, len(testCase.afterGameGenres[i].Games))
				for j, game := range genre.Games {
					assert.Equal(t, testCase.afterGameGenres[i].Games[j].ID, game.ID)
					assert.Equal(t, testCase.afterGameGenres[i].Games[j].Name, game.Name)
					assert.Equal(t, testCase.afterGameGenres[i].Games[j].Description, game.Description)
					assert.WithinDuration(t, testCase.afterGameGenres[i].Games[j].CreatedAt, game.CreatedAt, time.Second)
				}
			}
		})
	}
}
