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
	"gorm.io/gorm/logger"
)

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

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

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
					Games: []*migrate.GameTable2{
						{
							ID:               uuid.UUID(gameID1),
							Name:             "test",
							Description:      "test",
							CreatedAt:        now.Add(-time.Hour),
							VisibilityTypeID: gameVisibilityTypeIDPublic,
						},
					},
				},
				{
					ID:        uuid.UUID(genreID7),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
					Games: []*migrate.GameTable2{
						{
							ID:               uuid.UUID(gameID1),
							Name:             "test",
							Description:      "test",
							CreatedAt:        now.Add(-time.Hour),
							VisibilityTypeID: gameVisibilityTypeIDPublic,
						},
					},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(genreID7),
					Name:      "test2",
					CreatedAt: now.Add(-time.Hour * 2),
					Games: []*migrate.GameTable2{
						{
							ID:               uuid.UUID(gameID1),
							Name:             "test",
							Description:      "test",
							CreatedAt:        now.Add(-time.Hour),
							VisibilityTypeID: gameVisibilityTypeIDPublic,
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

				err = _db.Delete(&migrate.GameTable2{VisibilityTypeID: gameVisibilityTypeIDPublic}).Error
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

func TestGetGameGenresWithNames(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		gameGenreNames     []values.GameGenreName
		gameGenres         []migrate.GameGenreTable
		isErr              bool
		expectedGameGenres []*domain.GameGenre
		expectedErr        error
	}

	gameGenreID1 := uuid.New()
	gameGenreID2 := uuid.New()

	now := time.Now()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameGenreNames: []values.GameGenreName{values.NewGameGenreName("ジャンル")},
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        gameGenreID1,
					Name:      "ジャンル",
					CreatedAt: now.Add(-time.Hour),
				},
			},
			expectedGameGenres: []*domain.GameGenre{
				domain.NewGameGenre(values.GameGenreIDFromUUID(gameGenreID1), values.NewGameGenreName("ジャンル"), now.Add(-time.Hour)),
			},
		},
		"ジャンルが複数でも問題なし": {
			gameGenreNames: []values.GameGenreName{
				values.NewGameGenreName("ジャンル1"),
				values.NewGameGenreName("ジャンル2"),
			},
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        gameGenreID1,
					Name:      "ジャンル1",
					CreatedAt: now.Add(-time.Hour),
				},
				{
					ID:        gameGenreID2,
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour * 2),
				},
			},
			expectedGameGenres: []*domain.GameGenre{
				domain.NewGameGenre(values.GameGenreIDFromUUID(gameGenreID1), values.NewGameGenreName("ジャンル1"), now.Add(-time.Hour)),
				domain.NewGameGenre(values.GameGenreIDFromUUID(gameGenreID2), values.NewGameGenreName("ジャンル2"), now.Add(-time.Hour*2)),
			},
		},
		"関係ないジャンルがDBに合っても問題なし": {
			gameGenreNames: []values.GameGenreName{
				values.NewGameGenreName("ジャンル1"),
			},
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        gameGenreID1,
					Name:      "ジャンル1",
					CreatedAt: now.Add(-time.Hour),
				},
				{
					ID:        gameGenreID2,
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour * 2),
				},
			},
			expectedGameGenres: []*domain.GameGenre{
				domain.NewGameGenre(values.GameGenreIDFromUUID(gameGenreID1), values.NewGameGenreName("ジャンル1"), now.Add(-time.Hour)),
			},
		},
		"該当するジャンルが存在しないのでErrRecordNotFound": {
			gameGenreNames: []values.GameGenreName{
				values.NewGameGenreName("ジャンル1"),
			},
			gameGenres: []migrate.GameGenreTable{
				{
					ID:        gameGenreID2,
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour * 2),
				},
			},
			isErr:       true,
			expectedErr: repository.ErrRecordNotFound,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer cleanupGameGenresTable(t)

			if testCase.gameGenres != nil && len(testCase.gameGenres) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).Create(testCase.gameGenres).Error
				if err != nil {
					t.Fatalf("failed to create game genres: %v", err)
				}
			}

			genres, err := gameGenreRepository.GetGameGenresWithNames(ctx, testCase.gameGenreNames)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					if !errors.Is(err, testCase.expectedErr) {
						t.Fatalf("expected: %v, actual: %v", testCase.expectedErr, err)
					}
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil || testCase.isErr {
				return
			}

			assert.Len(t, genres, len(testCase.expectedGameGenres))

			for i := range genres {
				assert.Equal(t, testCase.expectedGameGenres[i].GetID(), genres[i].GetID())
				assert.Equal(t, testCase.expectedGameGenres[i].GetName(), genres[i].GetName())
				assert.WithinDuration(t, testCase.expectedGameGenres[i].GetCreatedAt(), genres[i].GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestSaveGameGenres(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		gameGenres       []*domain.GameGenre
		beforeGameGenres []migrate.GameGenreTable
		afterGameGenres  []migrate.GameGenreTable
		isErr            bool
		expectedErr      error
	}

	gameGenreID1 := uuid.New()
	gameGenreID2 := uuid.New()

	gameGenreName1 := "ジャンル1"
	gameGenreName2 := "ジャンル2"

	now := time.Now()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameGenres:      []*domain.GameGenre{domain.NewGameGenre(values.GameGenreID(gameGenreID1), values.GameGenreName(gameGenreName1), now)},
			afterGameGenres: []migrate.GameGenreTable{{ID: gameGenreID1, Name: gameGenreName1, CreatedAt: now}},
		},
		"他にジャンルがあってもエラー無し": {
			gameGenres:       []*domain.GameGenre{domain.NewGameGenre(values.GameGenreID(gameGenreID1), values.GameGenreName(gameGenreName1), now)},
			beforeGameGenres: []migrate.GameGenreTable{{ID: gameGenreID2, Name: gameGenreName2, CreatedAt: now.Add(-time.Hour)}},
			afterGameGenres: []migrate.GameGenreTable{
				{ID: gameGenreID1, Name: gameGenreName1, CreatedAt: now},
				{ID: gameGenreID2, Name: gameGenreName2, CreatedAt: now.Add(-time.Hour)},
			},
		},
		"複数ジャンルの作成でもエラー無し": {
			gameGenres: []*domain.GameGenre{
				domain.NewGameGenre(values.GameGenreID(gameGenreID1), values.GameGenreName(gameGenreName1), now),
				domain.NewGameGenre(values.GameGenreID(gameGenreID2), values.GameGenreName(gameGenreName2), now.Add(-time.Second)),
			},
			afterGameGenres: []migrate.GameGenreTable{
				{ID: gameGenreID1, Name: gameGenreName1, CreatedAt: now},
				{ID: gameGenreID2, Name: gameGenreName2, CreatedAt: now.Add(-time.Second)},
			},
		},
		"ジャンルが重複しているのでErrDuplicatedUniqueKey": {
			gameGenres:       []*domain.GameGenre{domain.NewGameGenre(values.GameGenreID(gameGenreID1), values.GameGenreName(gameGenreName1), now)},
			beforeGameGenres: []migrate.GameGenreTable{{ID: gameGenreID2, Name: gameGenreName1, CreatedAt: now.Add(-time.Hour)}},
			afterGameGenres:  []migrate.GameGenreTable{{ID: gameGenreID2, Name: gameGenreName1, CreatedAt: now.Add(-time.Hour)}},
			isErr:            true,
			expectedErr:      repository.ErrDuplicatedUniqueKey,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer cleanupGameGenresTable(t)

			if testCase.beforeGameGenres != nil && len(testCase.beforeGameGenres) > 0 {
				err := db.Create(&testCase.beforeGameGenres).Error
				if err != nil {
					t.Fatalf("failed to create game genres before sub test: %v", err)
				}
			}

			err := gameGenreRepository.SaveGameGenres(ctx, testCase.gameGenres)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var genres []migrate.GameGenreTable

			err = db.Order("created_at desc").Find(&genres).Error
			if err != nil {
				t.Fatalf("failed to get game genres: %v", err)
			}

			assert.Len(t, genres, len(testCase.afterGameGenres))

			for i, genre := range genres {
				assert.Equal(t, testCase.afterGameGenres[i].ID, genre.ID)
				assert.Equal(t, testCase.afterGameGenres[i].Name, genre.Name)
				assert.WithinDuration(t, testCase.afterGameGenres[i].CreatedAt, genre.CreatedAt, time.Second)
			}
		})
	}
}

func TestRegisterGenresToGame(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		gameID           values.GameID
		gameGenreIDs     []values.GameGenreID
		games            []migrate.GameTable2
		beforeGameGenres []migrate.GameGenreTable
		afterGameGenres  []migrate.GameGenreTable
		isErr            bool
		expectedErr      error
	}

	gameID1 := values.NewGameID()

	game1 := migrate.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "test",
		Description:      "test",
		CreatedAt:        time.Now(),
		VisibilityTypeID: 1,
	}

	gameGenreID1 := values.NewGameGenreID()
	gameGenreID2 := values.NewGameGenreID()

	now := time.Now()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameID:       gameID1,
			gameGenreIDs: []values.GameGenreID{gameGenreID1},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
		},
		"違うジャンルが紐づいていても問題なし": {
			gameID:       gameID1,
			gameGenreIDs: []values.GameGenreID{gameGenreID2},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{&game1},
				},
			},
		},
		"ジャンルの追加でも問題なし": {
			gameID:       gameID1,
			gameGenreIDs: []values.GameGenreID{gameGenreID1, gameGenreID2},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "ジャンル2",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{&game1},
				},
			},
		},
		"存在しないゲームなのでエラー": {
			gameID:       values.NewGameID(),
			gameGenreIDs: []values.GameGenreID{gameGenreID1},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
			isErr:       true,
			expectedErr: repository.ErrRecordNotFound,
		},
		"存在しないジャンルなのでエラー": {
			gameID:       gameID1,
			gameGenreIDs: []values.GameGenreID{values.NewGameGenreID()},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
			isErr:       true,
			expectedErr: repository.ErrIncludeInvalidArgs,
		},
		"ジャンルが空でもエラー無し": {
			gameID:       gameID1,
			gameGenreIDs: []values.GameGenreID{},
			games:        []migrate.GameTable2{game1},
			beforeGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{&game1},
				},
			},
			afterGameGenres: []migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "ジャンル1",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{},
				},
			},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer func() {
				cleanupGameGenresTable(t)
				err := db.
					Session(&gorm.Session{AllowGlobalUpdate: true}).
					Unscoped().
					Delete(&migrate.GameTable2{}).Error
				if err != nil {
					t.Fatalf("failed to delete games: %+v\n", err)
				}
			}()

			if testCase.games != nil && len(testCase.games) > 0 {
				err := db.Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create games before sub test: %v", err)
				}
			}

			if testCase.beforeGameGenres != nil && len(testCase.games) > 0 {
				err := db.Create(testCase.beforeGameGenres).Error
				if err != nil {
					t.Fatalf("failed to create game genres before sub test: %v", err)
				}
			}

			err := gameGenreRepository.RegisterGenresToGame(ctx, testCase.gameID, testCase.gameGenreIDs)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var genres []migrate.GameGenreTable

			genreIDs := make([]uuid.UUID, 0, len(testCase.afterGameGenres))
			for i := range testCase.afterGameGenres {
				genreIDs = append(genreIDs, uuid.UUID(testCase.afterGameGenres[i].ID))
			}
			err = db.Preload("Games").Where("`game_genres`.`id` in ?", genreIDs).Order("created_at desc").Find(&genres).Error
			if err != nil {
				t.Fatalf("failed to get game genres: %v", err)
			}

			assert.Len(t, genres, len(testCase.afterGameGenres))

			for i, genre := range genres {
				assert.Equal(t, testCase.afterGameGenres[i].ID, genre.ID)
				assert.Equal(t, testCase.afterGameGenres[i].Name, genre.Name)
				assert.WithinDuration(t, testCase.afterGameGenres[i].CreatedAt, genre.CreatedAt, time.Second)

				if testCase.afterGameGenres[i].Games != nil {
					assert.Len(t, genre.Games, len(testCase.afterGameGenres[i].Games))
					for j, game := range genre.Games {
						assert.Equal(t, testCase.afterGameGenres[i].Games[j].ID, game.ID)
						assert.Equal(t, testCase.afterGameGenres[i].Games[j].Name, game.Name)
						assert.Equal(t, testCase.afterGameGenres[i].Games[j].Description, game.Description)
						assert.WithinDuration(t, testCase.afterGameGenres[i].Games[j].CreatedAt, game.CreatedAt, time.Second)
					}
				}

				assert.WithinDuration(t, testCase.afterGameGenres[i].CreatedAt, genre.CreatedAt, time.Second)
			}
		})
	}
}

func TestGetGameGenres(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameGenreRepository := NewGameGenre(testDB)

	type test struct {
		visibilities       []values.GameVisibility
		gameGenres         []*migrate.GameGenreTable
		expectedGenresInfo []*repository.GameGenreInfo
		isErr              bool
		expectedErr        error
	}

	now := time.Now()

	var visibilities []migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&visibilities).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}

	var gameVisibilityTypeIDPublic int
	var gameVisibilityTypeIDPrivate int
	for i := range visibilities {
		switch visibilities[i].Name {
		case migrate.GameVisibilityTypePublic:
			gameVisibilityTypeIDPublic = visibilities[i].ID
		case migrate.GameVisibilityTypePrivate:
			gameVisibilityTypeIDPrivate = visibilities[i].ID
		}
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	game1 := &migrate.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "game1",
		VisibilityTypeID: gameVisibilityTypeIDPublic,
		CreatedAt:        now.Add(-time.Hour),
	}
	game2 := &migrate.GameTable2{
		ID:               uuid.UUID(gameID2),
		Name:             "game2",
		VisibilityTypeID: gameVisibilityTypeIDPrivate,
		CreatedAt:        now.Add(-time.Hour * 2),
	}

	gameGenreID1 := values.NewGameGenreID()
	gameGenreID2 := values.NewGameGenreID()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate},
			gameGenres: []*migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "3D",
					CreatedAt: now,
					Games:     []*migrate.GameTable2{game1},
				},
			},
			expectedGenresInfo: []*repository.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, "3D", now), Num: 1},
			},
		},
		"ジャンルが無くてもエラー無し": {
			visibilities:       []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate},
			gameGenres:         []*migrate.GameGenreTable{},
			expectedGenresInfo: []*repository.GameGenreInfo{},
		},
		"ジャンルがたくさんあってもエラー無し": {
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate},
			gameGenres: []*migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "3D",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{game1},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "2D",
					CreatedAt: now.Add(-time.Hour * 2),
					Games:     []*migrate.GameTable2{game2},
				},
			},
			expectedGenresInfo: []*repository.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, "3D", now.Add(-time.Hour)), Num: 1},
				{GameGenre: *domain.NewGameGenre(gameGenreID2, "2D", now.Add(-time.Hour*2)), Num: 1},
			},
		},
		"1つのジャンルにゲームがたくさんあってもエラー無し": {
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate},
			gameGenres: []*migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "3D",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{game1, game2},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "2D",
					CreatedAt: now.Add(-time.Hour * 2),
					Games:     []*migrate.GameTable2{game1, game2},
				},
			},
			expectedGenresInfo: []*repository.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, "3D", now.Add(-time.Hour)), Num: 2},
				{GameGenre: *domain.NewGameGenre(gameGenreID2, "2D", now.Add(-time.Hour*2)), Num: 2},
			},
		},
		"1つのジャンルにゲームが無くてもエラー無し": {
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate},
			gameGenres: []*migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "3D",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{game1, game2},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "2D",
					CreatedAt: now.Add(-time.Hour * 2),
					Games:     []*migrate.GameTable2{},
				},
			},
			expectedGenresInfo: []*repository.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, "3D", now.Add(-time.Hour)), Num: 2},
			},
		},
		"全てのvisibilityでなくてもok": {
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited},
			gameGenres: []*migrate.GameGenreTable{
				{
					ID:        uuid.UUID(gameGenreID1),
					Name:      "3D",
					CreatedAt: now.Add(-time.Hour),
					Games:     []*migrate.GameTable2{game1, game2},
				},
				{
					ID:        uuid.UUID(gameGenreID2),
					Name:      "2D",
					CreatedAt: now.Add(-time.Hour * 2),
					Games:     []*migrate.GameTable2{},
				},
			},
			expectedGenresInfo: []*repository.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, "3D", now.Add(-time.Hour)), Num: 1},
			},
		},
		"visibilityの値がおかしいのでエラー": {
			visibilities: []values.GameVisibility{100},
			isErr:        true,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer func() {
				cleanupGameGenresTable(t)
				err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
					Delete(&migrate.GameTable2{}).Error
				if err != nil {
					t.Fatalf("failed to clean up games table: %v\n", err)
				}
			}()

			if testCase.gameGenres != nil && len(testCase.gameGenres) > 0 {
				err := db.Create(testCase.gameGenres).Error
				if err != nil {
					t.Fatalf("failed to create game genres: %v\n", err)
				}
			}

			genreInfos, err := gameGenreRepository.GetGameGenres(ctx, testCase.visibilities)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil || testCase.isErr {
				return
			}

			assert.Len(t, genreInfos, len(testCase.expectedGenresInfo))

			for i := range genreInfos {
				assert.Equal(t, testCase.expectedGenresInfo[i].GetID(), genreInfos[i].GetID())
				assert.Equal(t, testCase.expectedGenresInfo[i].GetName(), genreInfos[i].GetName())
				assert.WithinDuration(t, testCase.expectedGenresInfo[i].GetCreatedAt(), genreInfos[i].GetCreatedAt(), time.Second)
				assert.Equal(t, testCase.expectedGenresInfo[i].Num, genreInfos[i].Num)
			}
		})

	}
}

// game_genresテーブルとgame_genre_relationsテーブルを削除する。gamesテーブルは削除されない。
func cleanupGameGenresTable(t *testing.T) {
	t.Helper()
	db, err := testDB.getDB(context.Background())
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	db = db.Session(&gorm.Session{AllowGlobalUpdate: true})

	var genres []migrate.GameGenreTable
	err = db.Find(&genres).Error
	if err != nil {
		t.Fatalf("failed to get genres")
	}

	err = db.
		Select("Games").
		Unscoped().
		Delete(&genres).Error
	if err != nil {
		t.Fatalf("failed to delete genres: %+v\n", err)
	}
}
