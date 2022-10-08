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
		beforeGames []migrate.GameTable
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			game: domain.NewGame(
				gameID1,
				"test",
				"test",
				now,
			),
		},
		{
			description: "別のゲームが存在してもエラーなし",
			game: domain.NewGame(
				gameID2,
				"test",
				"test",
				now,
			),
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "同じIDを持つゲームがあるのでエラー",
			game: domain.NewGame(
				gameID4,
				"test",
				"test",
				now,
			),
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.beforeGames != nil && len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.SaveGameV2(ctx, testCase.game)

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

			var game migrate.GameTable
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
		beforeGames []migrate.GameTable
		afterGames  []migrate.GameTable
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			game: domain.NewGame(
				gameID1,
				"test2",
				"test2",
				now,
			),
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			afterGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "別のゲームが存在してもエラーなし",
			game: domain.NewGame(
				gameID1,
				"test3",
				"test3",
				now,
			),
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			afterGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test3",
					Description: "test3",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			description: "ゲームが存在しないのでErrNoRecordUpdated",
			game: domain.NewGame(
				gameID1,
				"test2",
				"test2",
				now,
			),
			beforeGames: []migrate.GameTable{},
			afterGames:  []migrate.GameTable{},
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
					Delete(&migrate.GameTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}
			}()

			if testCase.beforeGames != nil && len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.UpdateGameV2(ctx, testCase.game)

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

			var games []migrate.GameTable
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
		beforeGames []migrate.GameTable
		afterGames  []migrate.GameTable
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			afterGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
				},
			},
		},
		{
			description: "別のゲームが存在してもエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			afterGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			description: "ゲームが存在しないのでErrNoRecordDeleted",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable{},
			afterGames:  []migrate.GameTable{},
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
					Delete(&migrate.GameTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete game: %+v\n", err)
				}
			}()

			if testCase.beforeGames != nil && len(testCase.beforeGames) != 0 {
				err := db.
					Session(&gorm.Session{
						Logger: logger.Default.LogMode(logger.Info),
					}).
					Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %+v\n", err)
				}
			}

			err := gameRepository.RemoveGameV2(ctx, testCase.gameID)

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

			var games []migrate.GameTable
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
		GameTable   []migrate.GameTable
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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			GameTable: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			game: domain.NewGame(
				gameID1,
				"test",
				"test",
				now,
			),
		},
		{
			description: "行ロックでもエラーなし",
			gameID:      gameID2,
			lockType:    repository.LockTypeRecord,
			GameTable: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			game: domain.NewGame(
				gameID2,
				"test",
				"test",
				now,
			),
		},
		{
			description: "ロックの種類が不正なのでエラー",
			gameID:      gameID5,
			lockType:    100,
			GameTable: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			game: domain.NewGame(
				gameID5,
				"test",
				"test",
				now,
			),
			isErr: true,
		},
		{
			description: "ゲームが存在しないのでErrRecordNotFound",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			GameTable:   []migrate.GameTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "ゲームが削除済みなのでErrRecordNotFound",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			GameTable: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Time:  now,
						Valid: true,
					},
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

			game, err := gameRepository.GetGameV2(ctx, testCase.gameID, testCase.lockType)

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
		description string
		limit       int
		offset      int
		beforeGames []migrate.GameTable
		games       []*domain.Game
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			limit:       -1,
			offset:      0,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"test",
					"test",
					now,
				),
			},
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			limit:       -1,
			offset:      -1,
			beforeGames: []migrate.GameTable{},
			games:       []*domain.Game{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			limit:       -1,
			offset:      0,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"test1",
					"test1",
					now,
				),
				domain.NewGame(
					gameID2,
					"test2",
					"test2",
					now.Add(-time.Hour),
				),
			},
		},
		{
			description: "limitが設定されてもエラーなし",
			limit:       1,
			offset:      0,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"test1",
					"test1",
					now,
				),
			},
		},
		{
			description: "offsetが設定されてもエラーなし",
			limit:       -1,
			offset:      1,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			games: []*domain.Game{
				domain.NewGame(
					gameID2,
					"test2",
					"test2",
					now.Add(-time.Hour),
				),
			},
		},
		{
			description: "limitとoffset両方が設定されてもエラーなし",
			limit:       1,
			offset:      1,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			games: []*domain.Game{
				domain.NewGame(
					gameID2,
					"test2",
					"test2",
					now.Add(-time.Hour),
				),
			},
		},
		{
			description: "limitが-1より小さいのでエラー",
			limit:       -2,
			offset:      0,
			beforeGames: []migrate.GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			isErr: true,
			err:   repository.ErrNegativeLimit,
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
					Delete(&migrate.GameTable{}).Error
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

			games, n, err := gameRepository.GetGamesV2(ctx, testCase.limit, testCase.offset)

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
			assert.Len(t, testCase.beforeGames, n)

			for i, game := range testCase.games {
				assert.Equal(t, game.GetID(), games[i].GetID())
				assert.Equal(t, game.GetName(), games[i].GetName())
				assert.Equal(t, game.GetDescription(), games[i].GetDescription())
				assert.WithinDuration(t, game.GetCreatedAt(), games[i].GetCreatedAt(), time.Second)
			}
		})
	}
}
