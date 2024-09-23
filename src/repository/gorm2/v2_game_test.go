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

func TestSaveGameV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description string
		game        *domain.Game
		beforeGames []migrate.GameTable2
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	now := time.Now()

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			game: domain.NewGame(
				gameID1,
				"test",
				"test",
				values.GameVisibilityTypeLimited,
				now,
			),
		},
		{
			description: "別のゲームが存在してもエラーなし",
			game: domain.NewGame(
				gameID2,
				"test",
				"test",
				values.GameVisibilityTypeLimited,
				now,
			),
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID3),
					Name:             "test",
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
		},
		{
			description: "同じIDを持つゲームがあるのでエラー",
			game: domain.NewGame(
				gameID4,
				"test",
				"test",
				values.GameVisibilityTypeLimited,
				now,
			),
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID4),
					Name:             "test",
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.SaveGame(ctx, testCase.game)

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

			var game migrate.GameTable2
			err = db.
				Session(&gorm.Session{}).
				Where("id = ?", uuid.UUID(testCase.game.GetID())).
				First(&game).Error
			if err != nil {
				t.Fatalf("failed to get game: %+v\n", err)
			}

			assert.Equal(t, uuid.UUID(testCase.game.GetID()), game.ID)
			assert.Equal(t, string(testCase.game.GetName()), game.Name)
			assert.Equal(t, string(testCase.game.GetDescription()), game.Description)
			assert.WithinDuration(t, testCase.game.GetCreatedAt(), game.CreatedAt, time.Second)
		})
	}
}

func TestUpdateGameV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description string
		game        *domain.Game
		beforeGames []migrate.GameTable2
		afterGames  []migrate.GameTable2
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	var gameVisibilityTypes []migrate.GameVisibilityTypeTable
	err = db.
		Find(&gameVisibilityTypes).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	var (
		gameVisibilityTypeIDPublic  int
		gameVisibilityTypeIDLimited int
		// gameVisibilityTypeIDPrivate int
	)
	for i := range gameVisibilityTypes {
		switch gameVisibilityTypes[i].Name {
		case migrate.GameVisibilityTypePublic:
			gameVisibilityTypeIDPublic = gameVisibilityTypes[i].ID
		case migrate.GameVisibilityTypeLimited:
			gameVisibilityTypeIDLimited = gameVisibilityTypes[i].ID
		case migrate.GameVisibilityTypePrivate:
			_ = gameVisibilityTypes[i].ID
		default:
			t.Fatalf("unknown game visibility type: %s", gameVisibilityTypes[i].Name)
		}
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			game: domain.NewGame(
				gameID1,
				"test2",
				"test2",
				values.GameVisibilityTypeLimited,
				now,
			),
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test1",
					Description:      "test1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			afterGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDLimited,
				},
			},
		},
		{
			description: "別のゲームが存在してもエラーなし",
			game: domain.NewGame(
				gameID1,
				"test3",
				"test3",
				values.GameVisibilityTypeLimited,
				now,
			),
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test1",
					Description:      "test1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			afterGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test3",
					Description:      "test3",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDLimited,
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
		},
		{
			description: "ゲームが存在しないのでErrNoRecordUpdated",
			game: domain.NewGame(
				gameID1,
				"test2",
				"test2",
				values.GameVisibilityTypeLimited,
				now,
			),
			beforeGames: []migrate.GameTable2{},
			afterGames:  []migrate.GameTable2{},
			isErr:       true,
			err:         repository.ErrNoRecordUpdated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Unscoped().
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&migrate.GameTable2{VisibilityTypeID: gameVisibilityTypeIDPublic}).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}
			}()

			if len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.UpdateGame(ctx, testCase.game)

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

			var games []migrate.GameTable2
			err = db.
				Session(&gorm.Session{}).
				Order("created_at desc").
				Find(&games).Error
			if err != nil {
				t.Fatalf("failed to get game: %+v\n", err)
			}

			assert.Len(t, games, len(testCase.afterGames))

			for i, game := range testCase.afterGames {
				assert.Equal(t, game.ID, games[i].ID)
				assert.Equal(t, game.Name, games[i].Name)
				assert.Equal(t, game.Description, games[i].Description)
				assert.Equal(t, game.VisibilityTypeID, games[i].VisibilityTypeID)
				assert.WithinDuration(t, game.CreatedAt, games[i].CreatedAt, time.Second)
			}
		})
	}
}

func TestRemoveGameV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description string
		gameID      values.GameID
		beforeGames []migrate.GameTable2
		afterGames  []migrate.GameTable2
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test1",
					Description:      "test1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			afterGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
		},
		{
			description: "別のゲームが存在してもエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test1",
					Description:      "test1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			afterGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
		},
		{
			description: "ゲームが存在しないのでErrNoRecordDeleted",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable2{},
			afterGames:  []migrate.GameTable2{},
			isErr:       true,
			err:         repository.ErrNoRecordDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Delete(&migrate.GameTable2{VisibilityTypeID: gameVisibilityTypeIDPublic}).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}
			}()

			if len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.RemoveGame(ctx, testCase.gameID)

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

			var games []migrate.GameTable2
			err = db.
				Unscoped().
				Session(&gorm.Session{}).
				Order("created_at desc").
				Find(&games).Error
			if err != nil {
				t.Fatalf("failed to get games: %+v\n", err)
			}

			assert.Len(t, games, len(testCase.afterGames))

			for i, game := range games {
				assert.Equal(t, testCase.afterGames[i].ID, game.ID)
				assert.Equal(t, testCase.afterGames[i].Name, game.Name)
				assert.Equal(t, testCase.afterGames[i].Description, game.Description)
				assert.WithinDuration(t, testCase.afterGames[i].CreatedAt, game.CreatedAt, time.Second)
				assert.Equal(t, testCase.afterGames[i].DeletedAt.Valid, game.DeletedAt.Valid)
				assert.WithinDuration(t, testCase.afterGames[i].DeletedAt.Time, game.DeletedAt.Time, time.Second)
			}
		})
	}
}

func TestGetGameV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		GameTable   []migrate.GameTable2
		game        *domain.Game
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()

	now := time.Now()

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			GameTable: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test",
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			game: domain.NewGame(
				gameID1,
				"test",
				"test",
				values.GameVisibilityTypePublic,
				now,
			),
		},
		{
			description: "行ロックでもエラーなし",
			gameID:      gameID2,
			lockType:    repository.LockTypeRecord,
			GameTable: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test",
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			game: domain.NewGame(
				gameID2,
				"test",
				"test",
				values.GameVisibilityTypePublic,
				now,
			),
		},
		{
			description: "ロックの種類が不正なのでエラー",
			gameID:      gameID5,
			lockType:    100,
			GameTable: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID5),
					Name:             "test",
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			game: domain.NewGame(
				gameID5,
				"test",
				"test",
				values.GameVisibilityTypePublic,
				now,
			),
			isErr: true,
		},
		{
			description: "ゲームが存在しないのでErrRecordNotFound",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			GameTable:   []migrate.GameTable2{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "ゲームが削除済みなのでErrRecordNotFound",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			GameTable: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Time:  now,
						Valid: true,
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.GameTable) != 0 {
				err := db.Create(&testCase.GameTable).Error
				if err != nil {
					t.Fatalf("failed to create test data: %+v\n", err)
				}

				for _, game := range testCase.GameTable {
					if game.DeletedAt.Valid {
						err = db.Delete(&game).Error
						if err != nil {
							t.Fatalf("failed to delete test data: %+v\n", err)
						}
					}
				}
			}

			game, err := gameRepository.GetGame(ctx, testCase.gameID, testCase.lockType)

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

			assert.Equal(t, testCase.game.GetID(), game.GetID())
			assert.Equal(t, testCase.game.GetName(), game.GetName())
			assert.Equal(t, testCase.game.GetDescription(), game.GetDescription())
			assert.Equal(t, testCase.game.GetVisibility(), game.GetVisibility())
			assert.WithinDuration(t, testCase.game.GetCreatedAt(), game.GetCreatedAt(), 2*time.Second)
		})
	}
}

func TestGetGamesV2(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		// 引数

		limit        int
		offset       int
		sort         repository.GamesSortType
		visibilities []values.GameVisibility
		userID       *values.TraPMemberID
		gameGenres   []values.GameGenreID
		gameName     string

		// テストデータ

		beforeGames []migrate.GameTable2

		// 返り値

		games       []*domain.GameWithGenres
		expectedNum int
		isErr       bool
		err         error
	}

	now := time.Now()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()

	gameName1 := values.NewGameName("test")
	gameName2 := values.NewGameName("テスト2")
	gameName3 := values.NewGameName("3テスト")

	game1 := domain.NewGame(gameID1, gameName1, "test", values.GameVisibilityTypePublic, now.Add(-time.Hour*2))
	game2 := domain.NewGame(gameID2, gameName2, "test", values.GameVisibilityTypeLimited, now.Add(-time.Hour))
	game3 := domain.NewGame(gameID3, gameName3, "test", values.GameVisibilityTypePrivate, now)

	gameGenreID1 := values.NewGameGenreID()
	gameGenreID2 := values.NewGameGenreID()

	gameGenreName1 := values.NewGameGenreName("ジャンル1")
	gameGenreName2 := values.NewGameGenreName("ジャンル2")

	gameGenre1 := domain.NewGameGenre(gameGenreID1, gameGenreName1, now.Add(-time.Hour))
	gameGenre2 := domain.NewGameGenre(gameGenreID2, gameGenreName2, now)

	memberUUID1 := uuid.New()
	memberUUID2 := uuid.New()
	trapMemberID1 := values.NewTrapMemberID(memberUUID1)

	var gameVisibilityTypes []migrate.GameVisibilityTypeTable
	err = db.
		Find(&gameVisibilityTypes).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	var (
		gameVisibilityTypeIDPublic  int
		gameVisibilityTypeIDLimited int
		gameVisibilityTypeIDPrivate int
	)
	for i := range gameVisibilityTypes {
		switch gameVisibilityTypes[i].Name {
		case migrate.GameVisibilityTypePublic:
			gameVisibilityTypeIDPublic = gameVisibilityTypes[i].ID
		case migrate.GameVisibilityTypeLimited:
			gameVisibilityTypeIDLimited = gameVisibilityTypes[i].ID
		case migrate.GameVisibilityTypePrivate:
			gameVisibilityTypeIDPrivate = gameVisibilityTypes[i].ID
		default:
			t.Fatalf("unknown game visibility type: %s", gameVisibilityTypes[i].Name)
		}
	}

	var gameRoleTypes []migrate.GameManagementRoleTypeTable
	err = db.Find(&gameRoleTypes).Error
	if err != nil {
		t.Fatalf("failed to get game management role type: %v\n", err)
	}
	var (
		gameRoleTypeIDOwner        int
		gameRoleTypeIDCollaborator int
	)
	for i := range gameRoleTypes {
		switch gameRoleTypes[i].Name {
		case migrate.GameManagementRoleTypeAdministrator:
			gameRoleTypeIDOwner = gameRoleTypes[i].ID
		case migrate.GameManagementRoleTypeCollaborator:
			gameRoleTypeIDCollaborator = gameRoleTypes[i].ID
		default:
			t.Fatalf("unknown game management role type: %s", gameRoleTypes[i].Name)
		}
	}

	testCases := map[string]test{
		"特に問題ないのでエラーなし": {
			limit:        1,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
			},
			games:       []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			expectedNum: 1,
		},
		"複数ゲームがあってもエラーなし": {
			limit:        2,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game2, []*domain.GameGenre{gameGenre2}), // 新しいゲームが先
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
			},
			expectedNum: 2,
		},
		"limitedとoffsetがあってもエラーなし": {
			limit:        1,
			offset:       1,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
			},
			expectedNum: 2,
		},
		"順番が最新バージョン順でもエラーなし": {
			limit:        2,
			offset:       0,
			sort:         repository.GamesSortTypeLatestVersion,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:                     uuid.UUID(gameID1),
					Name:                   string(gameName1),
					Description:            "test",
					CreatedAt:              now.Add(-time.Hour * 2),
					LatestVersionUpdatedAt: now,
					VisibilityTypeID:       gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
				{
					ID:                     uuid.UUID(gameID2),
					Name:                   string(gameName2),
					Description:            "test",
					CreatedAt:              now.Add(-time.Hour),
					LatestVersionUpdatedAt: now.Add(-time.Hour),
					VisibilityTypeID:       gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
				domain.NewGameWithGenres(game2, []*domain.GameGenre{gameGenre2}),
			},
			expectedNum: 2,
		},
		"visibilityの制限があっても問題なし": {
			limit:        3,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited},
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
				{
					ID:                     uuid.UUID(gameID3),
					Name:                   string(gameName3),
					Description:            "test",
					CreatedAt:              now.Add(-time.Hour),
					LatestVersionUpdatedAt: now.Add(-time.Hour),
					VisibilityTypeID:       gameVisibilityTypeIDPrivate,
					GameGenres: []*migrate.GameGenreTable{
						{
							ID:        uuid.UUID(gameGenreID2),
							Name:      string(gameGenreName2),
							CreatedAt: now,
						},
						{
							ID:        uuid.UUID(gameGenreID1),
							Name:      string(gameGenreName1),
							CreatedAt: now.Add(-time.Hour),
						}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game2, []*domain.GameGenre{gameGenre2}),
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
			},
			expectedNum: 2,
		},
		"ユーザーの指定があってもエラーなし": {
			limit:        2,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       &trapMemberID1,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							GameID:     uuid.UUID(gameID1),
							UserID:     memberUUID1,
							RoleTypeID: gameRoleTypeIDCollaborator,
						},
					},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							GameID:     uuid.UUID(gameID2),
							UserID:     memberUUID2,
							RoleTypeID: gameRoleTypeIDOwner,
						},
					},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
			},
			expectedNum: 1,
		},
		"ゲームジャンルの指定があってもエラーなし": {
			limit:        2,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   []values.GameGenreID{gameGenreID1},
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID1),
						Name:      string(gameGenreName1),
						CreatedAt: now.Add(-time.Hour),
					}},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
			},
			expectedNum: 1,
		},
		"ゲームジャンルの指定が複数あってもエラーなし": {
			limit:        2,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   []values.GameGenreID{gameGenreID1, gameGenreID2},
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
					GameGenres: []*migrate.GameGenreTable{
						{
							ID:        uuid.UUID(gameGenreID1),
							Name:      string(gameGenreName1),
							CreatedAt: now.Add(-time.Hour),
						},
						{
							ID:        uuid.UUID(gameGenreID2),
							Name:      string(gameGenreName2),
							CreatedAt: now,
						},
					},
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
					GameGenres: []*migrate.GameGenreTable{{
						ID:        uuid.UUID(gameGenreID2),
						Name:      string(gameGenreName2),
						CreatedAt: now,
					}},
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1, gameGenre2}),
			},
			expectedNum: 1,
		},
		"ゲーム名の指定があってもエラーなし": {
			limit:        3,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "テスト",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID2),
					Name:             string(gameName2),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDLimited,
				},
				{
					ID:               uuid.UUID(gameID3),
					Name:             string(gameName3),
					Description:      "test",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPrivate,
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game3, []*domain.GameGenre{}),
				domain.NewGameWithGenres(game2, []*domain.GameGenre{}),
			},
			expectedNum: 2,
		},
		"条件に合うゲームが無くてもエラー無し": {
			limit:        3,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   []values.GameGenreID{gameGenreID1},
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			games:       []*domain.GameWithGenres{},
			expectedNum: 0,
		},
		"limitが0(上限なし)でもエラー無し": {
			limit:        0,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: nil,
			userID:       nil,
			gameGenres:   nil,
			gameName:     "",
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             string(gameName1),
					Description:      "test",
					CreatedAt:        now.Add(-time.Hour * 2),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{}),
			},
			expectedNum: 1,
		},
		"limitが負なのでErrNegativeLimit": {
			limit:  -1,
			offset: 0,
			sort:   repository.GamesSortTypeCreatedAt,
			isErr:  true,
			err:    repository.ErrNegativeLimit,
		},
		"limitが0なのにoffsetが正なのでエラー": {
			limit:  0,
			offset: 1,
			sort:   repository.GamesSortTypeCreatedAt,
			isErr:  true,
		},
		"sortの値がおかしいのでエラーs": {
			limit:  1,
			offset: 0,
			sort:   100,
			isErr:  true,
		},
		"visibilityの値がおかしいのでエラー": {
			limit:        1,
			offset:       0,
			sort:         repository.GamesSortTypeCreatedAt,
			visibilities: []values.GameVisibility{100},
			isErr:        true,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			defer func() {
				var gameIDs []migrate.GameTable2
				err := db.Model(&migrate.GameTable2{}).Select("id").Find(&gameIDs).Error
				if err != nil {
					t.Fatalf("failed to get game ids: %+v\n", err)
				}

				// ゲームとジャンルとロールの削除
				err = db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Select("GameGenres", "GameManagementRoles").
					Unscoped().
					Delete(gameIDs).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}
			}()

			if len(testCase.beforeGames) != 0 {
				err := db.Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create test data: %+v\n", err)
				}
			}

			games, n, err := gameRepository.GetGames(
				ctx, testCase.limit, testCase.offset, testCase.sort,
				testCase.visibilities, testCase.userID, testCase.gameGenres, testCase.gameName)

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

			assert.Len(t, games, len(testCase.games))
			assert.Equal(t, testCase.expectedNum, n)

			for i := range testCase.games {
				assert.Equal(t, games[i].GetGame().GetID(), testCase.games[i].GetGame().GetID())
				assert.Equal(t, games[i].GetGame().GetName(), testCase.games[i].GetGame().GetName())
				assert.Equal(t, games[i].GetGame().GetDescription(), testCase.games[i].GetGame().GetDescription())
				assert.Equal(t, games[i].GetGame().GetVisibility(), testCase.games[i].GetGame().GetVisibility())
				assert.WithinDuration(t, games[i].GetGame().GetCreatedAt(), testCase.games[i].GetGame().GetCreatedAt(), time.Second)

				testCaseGameGenres := testCase.games[i].GetGenres()
				assert.Len(t, games[i].GetGenres(), len(testCaseGameGenres))

				// ゲームジャンルの順番は保証していないので、mapに直してから比較
				testCaseGameGenresMap := make(map[values.GameGenreID]*domain.GameGenre, len(testCaseGameGenres))
				for j := range testCase.games[i].GetGenres() {
					testCaseGameGenresMap[testCaseGameGenres[j].GetID()] = testCaseGameGenres[j]
				}
				for j := range testCase.games[i].GetGenres() {
					assert.Contains(t, testCaseGameGenresMap, games[i].GetGenres()[j].GetID())

					genre := testCaseGameGenresMap[games[i].GetGenres()[j].GetID()]
					//存在しなかったら上のContainsでテスト失敗するので、存在確認はしない
					assert.Equal(t, genre.GetName(), games[i].GetGenres()[j].GetName())
					assert.WithinDuration(t, genre.GetCreatedAt(), games[i].GetGenres()[j].GetCreatedAt(), time.Second)
				}
			}
		})
	}
}

func TestGetGamesByIDsV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description   string
		gameIDs       []values.GameID
		lockType      repository.LockType
		beforeGames   []migrate.GameTable2
		expectedGames []*domain.Game
		isErr         bool
		err           error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()
	gameID8 := values.NewGameID()
	gameID9 := values.NewGameID()

	now := time.Now()

	var roleTypes []*migrate.GameManagementRoleTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&roleTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	roleTypeMap := make(map[string]int, len(roleTypes))
	for _, roleType := range roleTypes {
		roleTypeMap[roleType.Name] = roleType.ID
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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameIDs:     []values.GameID{gameID1},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test1",
					Description:      "test1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			expectedGames: []*domain.Game{
				domain.NewGame(gameID1, "test1", "test1", values.GameVisibilityTypeLimited, now),
			},
		},
		{
			description: "gameIDが複数でもエラーなし",
			gameIDs:     []values.GameID{gameID2, gameID3},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test2",
					Description:      "test2",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID3),
					Name:             "test3",
					Description:      "test3",
					CreatedAt:        now.Add(-time.Hour),
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			expectedGames: []*domain.Game{
				domain.NewGame(gameID2, "test2", "test2", values.GameVisibilityTypeLimited, now),
				domain.NewGame(gameID3, "test3", "test3", values.GameVisibilityTypeLimited, now.Add(-time.Hour)),
			},
		},
		{
			description: "違うgameIDのゲームは取らない",
			gameIDs:     []values.GameID{gameID4},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID4),
					Name:             "test4",
					Description:      "test4",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
				{
					ID:               uuid.UUID(gameID5),
					Name:             "test5",
					Description:      "test5",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			expectedGames: []*domain.Game{
				domain.NewGame(gameID4, "test4", "test4", values.GameVisibilityTypeLimited, now),
			},
		},
		{
			description: "削除されたゲームは取らない",
			gameIDs:     []values.GameID{gameID6},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test6",
					Description: "test6",
					CreatedAt:   now.Add(-time.Hour),
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			expectedGames: []*domain.Game{},
		},
		{
			description:   "ゲームが存在しなくても問題なし",
			gameIDs:       []values.GameID{gameID7},
			lockType:      repository.LockTypeNone,
			beforeGames:   []migrate.GameTable2{},
			expectedGames: []*domain.Game{},
		},
		{
			description: "lockTypeがrecordでも問題なし",
			gameIDs:     []values.GameID{gameID8},
			lockType:    repository.LockTypeRecord,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID8),
					Name:             "test8",
					Description:      "test8",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			expectedGames: []*domain.Game{
				domain.NewGame(gameID8, "test8", "test8", values.GameVisibilityTypeLimited, now),
			},
		},
		{
			description: "lockTypeが無効",
			gameIDs:     []values.GameID{gameID9},
			lockType:    100,
			beforeGames: []migrate.GameTable2{
				{
					ID:               uuid.UUID(gameID9),
					Name:             "test9",
					Description:      "test9",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGames) != 0 {
				err := db.Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create test data: %+v\n", err)
				}
			}

			games, err := gameRepository.GetGamesByIDs(ctx, testCase.gameIDs, testCase.lockType)

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

			assert.Len(t, games, len(testCase.expectedGames))

			gameMap := make(map[values.GameID]*domain.Game)
			for _, game := range games {
				gameMap[game.GetID()] = game
			}

			for _, expectedGame := range testCase.expectedGames {
				game, ok := gameMap[expectedGame.GetID()]
				if !ok {
					t.Errorf("game must be %+v, but actual is nil", expectedGame)
					continue
				}

				assert.Equal(t, expectedGame.GetID(), game.GetID())
				assert.Equal(t, expectedGame.GetName(), game.GetName())
				assert.Equal(t, expectedGame.GetDescription(), game.GetDescription())
				assert.WithinDuration(t, expectedGame.GetCreatedAt(), game.GetCreatedAt(), time.Second)
			}
		})
	}
}
