package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSaveGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description string
		game        *domain.Game
		beforeGames []GameTable
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
			beforeGames: []GameTable{
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
			beforeGames: []GameTable{
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

			var game GameTable
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

func TestUpdateGame(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description string
		game        *domain.Game
		beforeGames []GameTable
		afterGames  []GameTable
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
			beforeGames: []GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			afterGames: []GameTable{
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
			beforeGames: []GameTable{
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
			afterGames: []GameTable{
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
			beforeGames: []GameTable{},
			afterGames:  []GameTable{},
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
					Delete(&GameTable{}).Error
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

			var games []GameTable
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

func TestRemoveGame(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description string
		gameID      values.GameID
		beforeGames []GameTable
		afterGames  []GameTable
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
			beforeGames: []GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			afterGames: []GameTable{
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
			beforeGames: []GameTable{
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
			afterGames: []GameTable{
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
			beforeGames: []GameTable{},
			afterGames:  []GameTable{},
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
					Delete(&GameTable{}).Error
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

			var games []GameTable
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

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		gameTable   []GameTable
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
			gameTable: []GameTable{
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
			gameTable: []GameTable{
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
			gameTable: []GameTable{
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
			gameTable:   []GameTable{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "ゲームが削除済みなのでErrRecordNotFound",
			gameID:      gameID4,
			lockType:    repository.LockTypeNone,
			gameTable: []GameTable{
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
			if len(testCase.gameTable) != 0 {
				err := db.Create(&testCase.gameTable).Error
				if err != nil {
					t.Fatalf("failed to create test data: %+v\n", err)
				}

				for _, game := range testCase.gameTable {
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
			assert.WithinDuration(t, testCase.game.GetCreatedAt(), game.GetCreatedAt(), 2*time.Second)
		})
	}
}

func TestGetGames(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description string
		beforeGames []GameTable
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
			beforeGames: []GameTable{
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
			beforeGames: []GameTable{},
			games:       []*domain.Game{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			beforeGames: []GameTable{
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Unscoped().
					Delete(&GameTable{}).Error
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

			games, err := gameRepository.GetGames(ctx)

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

			for i, game := range testCase.games {
				assert.Equal(t, game.GetID(), games[i].GetID())
				assert.Equal(t, game.GetName(), games[i].GetName())
				assert.Equal(t, game.GetDescription(), games[i].GetDescription())
				assert.WithinDuration(t, game.GetCreatedAt(), games[i].GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetGamesByIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description     string
		beforeGameTable []GameTable
		gameIDs         []values.GameID
		lockType        repository.LockType
		games           []*domain.Game
		isErr           bool
		err             error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			beforeGameTable: []GameTable{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			gameIDs:  []values.GameID{gameID1},
			lockType: repository.LockTypeNone,
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
			description:     "ゲームが存在しないので含まない",
			beforeGameTable: []GameTable{},
			gameIDs:         []values.GameID{gameID2},
			lockType:        repository.LockTypeNone,
			games:           []*domain.Game{},
		},
		{
			description: "ゲームが削除済みなので含まない",
			beforeGameTable: []GameTable{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
						Time:  now,
					},
				},
			},
			gameIDs:  []values.GameID{gameID3},
			lockType: repository.LockTypeNone,
			games:    []*domain.Game{},
		},
		{
			description: "ゲームが複数でも問題なし",
			beforeGameTable: []GameTable{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			gameIDs:  []values.GameID{gameID4, gameID5},
			lockType: repository.LockTypeNone,
			games: []*domain.Game{
				domain.NewGame(
					gameID4,
					"test",
					"test",
					now,
				),
				domain.NewGame(
					gameID5,
					"test",
					"test",
					now,
				),
			},
		},
		{
			description: "含まないゲームが存在してもエラーなし",
			beforeGameTable: []GameTable{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
			gameIDs:  []values.GameID{gameID6},
			lockType: repository.LockTypeNone,
			games: []*domain.Game{
				domain.NewGame(
					gameID6,
					"test",
					"test",
					now,
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.beforeGameTable != nil && len(testCase.beforeGameTable) != 0 {
				err := db.Create(&testCase.beforeGameTable).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
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

			assert.Len(t, games, len(testCase.games))

			gameMap := make(map[values.GameID]*domain.Game, len(testCase.games))
			for _, game := range games {
				gameMap[game.GetID()] = game
			}

			for _, game := range testCase.games {
				actualGame, ok := gameMap[game.GetID()]
				assert.True(t, ok)

				assert.Equal(t, game.GetID(), actualGame.GetID())
				assert.Equal(t, game.GetName(), actualGame.GetName())
				assert.Equal(t, game.GetDescription(), actualGame.GetDescription())
				assert.WithinDuration(t, game.GetCreatedAt(), actualGame.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetGamesByLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description            string
		beforeLauncherVersions []LauncherVersionTable
		launcherVersionID      values.LauncherVersionID
		games                  []*domain.Game
		isErr                  bool
		err                    error
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()
	launcherVersionID3 := values.NewLauncherVersionID()
	launcherVersionID4 := values.NewLauncherVersionID()
	launcherVersionID5 := values.NewLauncherVersionID()
	launcherVersionID6 := values.NewLauncherVersionID()
	launcherVersionID7 := values.NewLauncherVersionID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID1),
					Name: "TestGetGamesByLauncherVersion1",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test1",
							Description: "test1",
							CreatedAt:   now,
						},
					},
				},
			},
			launcherVersionID: launcherVersionID1,
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
			description:            "存在しないランチャーバージョンIDを指定した場合空配列を返す",
			beforeLauncherVersions: []LauncherVersionTable{},
			launcherVersionID:      launcherVersionID2,
			games:                  []*domain.Game{},
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID3),
					Name: "TestGetGamesByLauncherVersion3",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games:     []GameTable{},
				},
			},
			launcherVersionID: launcherVersionID3,
			games:             []*domain.Game{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID4),
					Name: "TestGetGamesByLauncherVersion4",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID2),
							Name:        "test2",
							Description: "test2",
							CreatedAt:   now,
						},
						{
							ID:          uuid.UUID(gameID3),
							Name:        "test3",
							Description: "test3",
							CreatedAt:   now.Add(-time.Hour),
						},
					},
				},
			},
			launcherVersionID: launcherVersionID4,
			games: []*domain.Game{
				domain.NewGame(
					gameID2,
					"test2",
					"test2",
					now,
				),
				domain.NewGame(
					gameID3,
					"test3",
					"test3",
					now.Add(-time.Hour),
				),
			},
		},
		{
			description: "他のランチャーバージョンのゲームは含まない",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID5),
					Name: "TestGetGamesByLauncherVersion5",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID4),
							Name:        "test4",
							Description: "test4",
							CreatedAt:   now,
						},
					},
				},
				{
					ID:   uuid.UUID(launcherVersionID6),
					Name: "TestGetGamesByLauncherVersion6",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID5),
							Name:        "test5",
							Description: "test5",
							CreatedAt:   now,
						},
					},
				},
			},
			launcherVersionID: launcherVersionID5,
			games: []*domain.Game{
				domain.NewGame(
					gameID4,
					"test4",
					"test4",
					now,
				),
			},
		},
		{
			description: "削除されたゲームは含まない",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID7),
					Name: "TestGetGamesByLauncherVersion7",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID6),
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
			},
			launcherVersionID: launcherVersionID7,
			games:             []*domain.Game{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.beforeLauncherVersions != nil && len(testCase.beforeLauncherVersions) != 0 {
				err := db.Create(&testCase.beforeLauncherVersions).Error
				if err != nil {
					t.Fatalf("failed to create test data: %s", err)
				}
			}

			games, err := gameRepository.GetGamesByLauncherVersion(ctx, testCase.launcherVersionID)

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

			for i, game := range testCase.games {
				actualGame := games[i]

				assert.Equal(t, game.GetID(), actualGame.GetID())
				assert.Equal(t, game.GetName(), actualGame.GetName())
				assert.Equal(t, game.GetDescription(), actualGame.GetDescription())
				assert.WithinDuration(t, game.GetCreatedAt(), actualGame.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetGameInfosByLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGame(testDB)

	type test struct {
		description            string
		beforeLauncherVersions []LauncherVersionTable
		launcherVersionID      values.LauncherVersionID
		fileTypes              []values.GameFileType
		gameInfos              []*repository.GameInfo
		isErr                  bool
		err                    error
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()
	launcherVersionID3 := values.NewLauncherVersionID()
	launcherVersionID4 := values.NewLauncherVersionID()
	launcherVersionID5 := values.NewLauncherVersionID()
	launcherVersionID6 := values.NewLauncherVersionID()
	launcherVersionID7 := values.NewLauncherVersionID()
	launcherVersionID8 := values.NewLauncherVersionID()
	launcherVersionID9 := values.NewLauncherVersionID()
	launcherVersionID10 := values.NewLauncherVersionID()
	launcherVersionID11 := values.NewLauncherVersionID()
	launcherVersionID12 := values.NewLauncherVersionID()
	launcherVersionID13 := values.NewLauncherVersionID()
	launcherVersionID14 := values.NewLauncherVersionID()
	launcherVersionID15 := values.NewLauncherVersionID()
	launcherVersionID16 := values.NewLauncherVersionID()
	launcherVersionID17 := values.NewLauncherVersionID()
	launcherVersionID18 := values.NewLauncherVersionID()
	launcherVersionID19 := values.NewLauncherVersionID()
	launcherVersionID20 := values.NewLauncherVersionID()
	launcherVersionID21 := values.NewLauncherVersionID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()
	gameID8 := values.NewGameID()
	gameID9 := values.NewGameID()
	gameID10 := values.NewGameID()
	gameID11 := values.NewGameID()
	gameID12 := values.NewGameID()
	gameID13 := values.NewGameID()
	gameID14 := values.NewGameID()
	gameID15 := values.NewGameID()
	gameID16 := values.NewGameID()
	gameID17 := values.NewGameID()
	gameID18 := values.NewGameID()
	gameID19 := values.NewGameID()
	gameID20 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()
	gameVersionID10 := values.NewGameVersionID()
	gameVersionID11 := values.NewGameVersionID()
	gameVersionID12 := values.NewGameVersionID()
	gameVersionID13 := values.NewGameVersionID()
	gameVersionID14 := values.NewGameVersionID()
	gameVersionID15 := values.NewGameVersionID()
	gameVersionID16 := values.NewGameVersionID()
	gameVersionID17 := values.NewGameVersionID()
	gameVersionID18 := values.NewGameVersionID()
	gameVersionID19 := values.NewGameVersionID()
	gameVersionID20 := values.NewGameVersionID()

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()
	gameFileID4 := values.NewGameFileID()
	gameFileID5 := values.NewGameFileID()
	gameFileID6 := values.NewGameFileID()
	gameFileID7 := values.NewGameFileID()
	gameFileID8 := values.NewGameFileID()
	gameFileID9 := values.NewGameFileID()
	gameFileID10 := values.NewGameFileID()
	gameFileID11 := values.NewGameFileID()
	gameFileID12 := values.NewGameFileID()
	gameFileID13 := values.NewGameFileID()
	gameFileID14 := values.NewGameFileID()
	gameFileID15 := values.NewGameFileID()
	gameFileID16 := values.NewGameFileID()
	gameFileID17 := values.NewGameFileID()
	gameFileID18 := values.NewGameFileID()
	gameFileID19 := values.NewGameFileID()
	gameFileID20 := values.NewGameFileID()

	_, err, _ = fileTypeSetupGroup.Do("setupFileTypeTable", func() (interface{}, error) {
		return nil, setupFileTypeTable(db)
	})
	if err != nil {
		t.Fatalf("failed to setup file type table: %+v\n", err)
	}

	var fileTypes []*GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&fileTypes).Error
	if err != nil {
		t.Fatalf("failed to get file types: %v\n", err)
	}

	fileTypeMap := make(map[string]int, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypeMap[fileType.Name] = fileType.ID
	}

	gameURLID1 := values.NewGameURLID()
	gameURLID2 := values.NewGameURLID()
	gameURLID3 := values.NewGameURLID()
	gameURLID4 := values.NewGameURLID()
	gameURLID5 := values.NewGameURLID()
	gameURLID6 := values.NewGameURLID()
	gameURLID7 := values.NewGameURLID()
	gameURLID8 := values.NewGameURLID()
	gameURLID9 := values.NewGameURLID()
	gameURLID10 := values.NewGameURLID()
	gameURLID11 := values.NewGameURLID()
	gameURLID12 := values.NewGameURLID()
	gameURLID13 := values.NewGameURLID()
	gameURLID14 := values.NewGameURLID()
	gameURLID15 := values.NewGameURLID()
	gameURLID16 := values.NewGameURLID()
	gameURLID17 := values.NewGameURLID()
	gameURLID18 := values.NewGameURLID()
	gameURLID19 := values.NewGameURLID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	gameImageID1 := values.NewGameImageID()
	gameImageID2 := values.NewGameImageID()
	gameImageID3 := values.NewGameImageID()
	gameImageID4 := values.NewGameImageID()
	gameImageID5 := values.NewGameImageID()
	gameImageID6 := values.NewGameImageID()
	gameImageID7 := values.NewGameImageID()
	gameImageID8 := values.NewGameImageID()
	gameImageID9 := values.NewGameImageID()
	gameImageID10 := values.NewGameImageID()
	gameImageID11 := values.NewGameImageID()
	gameImageID12 := values.NewGameImageID()
	gameImageID13 := values.NewGameImageID()
	gameImageID14 := values.NewGameImageID()
	gameImageID15 := values.NewGameImageID()
	gameImageID16 := values.NewGameImageID()
	gameImageID17 := values.NewGameImageID()
	gameImageID18 := values.NewGameImageID()
	gameImageID19 := values.NewGameImageID()
	gameImageID20 := values.NewGameImageID()

	_, err, _ = imageTypeSetupGroup.Do("setupImageTypeTable", func() (interface{}, error) {
		return nil, setupImageTypeTable(db)
	})
	if err != nil {
		t.Fatalf("failed to setup image type table: %v\n", err)
	}

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

	gameVideoID1 := values.NewGameVideoID()
	gameVideoID2 := values.NewGameVideoID()
	gameVideoID3 := values.NewGameVideoID()
	gameVideoID4 := values.NewGameVideoID()
	gameVideoID5 := values.NewGameVideoID()
	gameVideoID6 := values.NewGameVideoID()
	gameVideoID7 := values.NewGameVideoID()
	gameVideoID8 := values.NewGameVideoID()
	gameVideoID9 := values.NewGameVideoID()
	gameVideoID10 := values.NewGameVideoID()
	gameVideoID11 := values.NewGameVideoID()
	gameVideoID12 := values.NewGameVideoID()
	gameVideoID13 := values.NewGameVideoID()
	gameVideoID14 := values.NewGameVideoID()
	gameVideoID15 := values.NewGameVideoID()
	gameVideoID16 := values.NewGameVideoID()
	gameVideoID17 := values.NewGameVideoID()
	gameVideoID18 := values.NewGameVideoID()
	gameVideoID19 := values.NewGameVideoID()
	gameVideoID20 := values.NewGameVideoID()

	_, err, _ = videoTypeSetupGroup.Do("setupVideoTypeTable", func() (interface{}, error) {
		return nil, setupVideoTypeTable(db)
	})
	if err != nil {
		t.Fatalf("failed to setup video type table: %v\n", err)
	}

	var videoTypes []*GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&videoTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	videoTypeMap := make(map[string]int, len(videoTypes))
	for _, videoType := range videoTypes {
		videoTypeMap[videoType.Name] = videoType.ID
	}

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID1),
					Name: "Tggiblv1",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID1),
							Name:        "test1",
							Description: "test1",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID1),
									Name:        "test1",
									Description: "test1",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID1),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID1),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID1),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID1),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID1,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"test1",
						"test1",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						"test1",
						"test1",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID1,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description:            "ランチャーバージョンが存在しないのでRecordNotFound",
			beforeLauncherVersions: []LauncherVersionTable{},
			launcherVersionID:      launcherVersionID2,
			fileTypes:              []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			isErr:                  true,
			err:                    repository.ErrRecordNotFound,
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID3),
					Name: "Tggiblv3",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games:     []GameTable{},
				},
			},
			launcherVersionID: launcherVersionID3,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos:         []*repository.GameInfo{},
		},
		{
			description: "バージョンが存在しないゲームは除外する",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID4),
					Name: "Tggiblv4",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:           uuid.UUID(gameID2),
							Name:         "test2",
							Description:  "test2",
							CreatedAt:    now,
							GameVersions: []GameVersionTable{},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID2),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID2),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID4,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos:         []*repository.GameInfo{},
		},
		{
			description: "ゲームファイルが存在しなくてもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID5),
					Name: "Tggiblv5",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID3),
							Name:        "test3",
							Description: "test3",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID2),
									Name:        "test2",
									Description: "test2",
									CreatedAt:   now,
									GameFiles:   []GameFileTable{},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID2),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID3),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID3),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID5,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID3,
						"test3",
						"test3",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						"test2",
						"test2",
						now,
					),
					LatestFiles: []*domain.GameFile{},
					LatestURL: domain.NewGameURL(
						gameURLID2,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID3,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID3,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "URLが存在しなくてもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID6),
					Name: "Tggiblv6",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID4),
							Name:        "test4",
							Description: "test4",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID3),
									Name:        "test3",
									Description: "test3",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID2),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID4),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID4),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID6,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID4,
						"test4",
						"test4",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID3,
						"test3",
						"test3",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID2,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID4,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID4,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "Imageが存在しない場合除外",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID7),
					Name: "Tggiblv7",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID5),
							Name:        "test5",
							Description: "test5",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID4),
									Name:        "test4",
									Description: "test4",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID3),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID3),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID5),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID7,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos:         []*repository.GameInfo{},
		},
		{
			description: "ゲーム紹介動画が存在しなくてもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID8),
					Name: "Tggiblv8",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID6),
							Name:        "test6",
							Description: "test6",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID5),
									Name:        "test5",
									Description: "test5",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID4),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID4),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID5),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID8,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID6,
						"test6",
						"test6",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID5,
						"test5",
						"test5",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID4,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID4,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID5,
						values.GameImageTypePng,
						now,
					),
				},
			},
		},
		{
			description: "他のランチャーバージョンが存在してもエラーなし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID9),
					Name: "Tggiblv9",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID7),
							Name:        "test7",
							Description: "test7",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID6),
									Name:        "test6",
									Description: "test6",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID5),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID5),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID6),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID6),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
				{
					ID:   uuid.UUID(launcherVersionID10),
					Name: "Tggiblv10",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID8),
							Name:        "test1",
							Description: "test1",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID7),
									Name:        "test7",
									Description: "test7",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID6),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID6),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID7),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID7),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID9,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID7,
						"test7",
						"test7",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID6,
						"test6",
						"test6",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID5,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID5,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID6,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID6,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ゲームが複数存在しても問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID11),
					Name: "Tggiblv11",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID9),
							Name:        "test9",
							Description: "test9",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID8),
									Name:        "test8",
									Description: "test8",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID7),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID7),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID8),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID8),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
						{
							ID:          uuid.UUID(gameID10),
							Name:        "test10",
							Description: "test10",
							CreatedAt:   now.Add(-time.Hour),
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID9),
									Name:        "test9",
									Description: "test9",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID8),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID8),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID9),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID9),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID11,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID9,
						"test9",
						"test9",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID8,
						"test8",
						"test8",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID7,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID7,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID8,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID8,
						values.GameVideoTypeMp4,
						now,
					),
				},
				{
					Game: domain.NewGame(
						gameID10,
						"test10",
						"test10",
						now.Add(-time.Hour),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID9,
						"test9",
						"test9",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID8,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID8,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID9,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID9,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ゲームバージョンが複数の場合、最新のものを使用する",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID12),
					Name: "Tggiblv12",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID11),
							Name:        "test11",
							Description: "test11",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID10),
									Name:        "test10",
									Description: "test10",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID9),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID9),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
								{
									ID:          uuid.UUID(gameVersionID11),
									Name:        "test11",
									Description: "test11",
									CreatedAt:   now.Add(-time.Hour),
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID10),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID10),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID10),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID10),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID12,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID11,
						"test11",
						"test11",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID10,
						"test10",
						"test10",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID9,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID9,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID10,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID10,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ファイルが複数存在する場合、すべて含む",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID13),
					Name: "Tggiblv13",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID12),
							Name:        "test12",
							Description: "test12",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID12),
									Name:        "test12",
									Description: "test12",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID11),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
										{
											ID:         uuid.UUID(gameFileID12),
											FileTypeID: fileTypeMap[gameFileTypeWindows],
											Hash:       "68617368",
											EntryPoint: "main.exe",
											CreatedAt:  now.Add(-time.Hour),
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID11),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID11),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID11),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID13,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID12,
						"test12",
						"test12",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID12,
						"test12",
						"test12",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID11,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
						domain.NewGameFile(
							gameFileID12,
							values.GameFileTypeWindows,
							"main.exe",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now.Add(-time.Hour),
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID11,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID11,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID11,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "imageが複数の場合、最新のもののみ取得される",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID14),
					Name: "Tggiblv14",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID13),
							Name:        "test13",
							Description: "test13",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID13),
									Name:        "test13",
									Description: "test13",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID13),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID12),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID12),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
								{
									ID:          uuid.UUID(gameImageID13),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now.Add(-time.Hour),
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID12),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID14,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID13,
						"test13",
						"test13",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID13,
						"test13",
						"test13",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID13,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID12,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID12,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID12,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ゲーム紹介動画が複数の場合、最新のもののみ取得される",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID15),
					Name: "Tggiblv15",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID14),
							Name:        "test14",
							Description: "test14",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID14),
									Name:        "test14",
									Description: "test14",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID14),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID13),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID14),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID13),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
								{
									ID:          uuid.UUID(gameVideoID14),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now.Add(-time.Hour),
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID15,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID14,
						"test14",
						"test14",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID14,
						"test14",
						"test14",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID14,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID13,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID14,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID13,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ファイルがwindows用でも問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID16),
					Name: "Tggiblv16",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID15),
							Name:        "test15",
							Description: "test15",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID15),
									Name:        "test15",
									Description: "test15",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID15),
											FileTypeID: fileTypeMap[gameFileTypeWindows],
											Hash:       "68617368",
											EntryPoint: "main.exe",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID14),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID15),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID15),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID16,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID15,
						"test15",
						"test15",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID15,
						"test15",
						"test15",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID15,
							values.GameFileTypeWindows,
							"main.exe",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID14,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID15,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID15,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "ファイルがmac用でも問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID17),
					Name: "Tggiblv17",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID16),
							Name:        "test16",
							Description: "test16",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID16),
									Name:        "test16",
									Description: "test16",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID16),
											FileTypeID: fileTypeMap[gameFileTypeMac],
											Hash:       "68617368",
											EntryPoint: "main.app",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID15),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID16),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID16),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID17,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID16,
						"test16",
						"test16",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID16,
						"test16",
						"test16",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID16,
							values.GameFileTypeMac,
							"main.app",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID15,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID16,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID16,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "画像がjpegでも問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID18),
					Name: "Tggiblv18",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID17),
							Name:        "test17",
							Description: "test17",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID17),
									Name:        "test17",
									Description: "test17",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID17),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID16),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID17),
									ImageTypeID: imageTypeMap[gameImageTypeJpeg],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID17),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID18,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID17,
						"test17",
						"test17",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID17,
						"test17",
						"test17",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID17,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID16,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID17,
						values.GameImageTypeJpeg,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID17,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "画像がgifでも問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID19),
					Name: "Tggiblv19",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID18),
							Name:        "test18",
							Description: "test18",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID18),
									Name:        "test18",
									Description: "test18",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID18),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID17),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID18),
									ImageTypeID: imageTypeMap[gameImageTypeGif],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID18),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID19,
			fileTypes:         []values.GameFileType{values.GameFileTypeJar, values.GameFileTypeWindows, values.GameFileTypeMac},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID18,
						"test18",
						"test18",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID18,
						"test18",
						"test18",
						now,
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID18,
							values.GameFileTypeJar,
							"main.jar",
							values.NewGameFileHashFromBytes([]byte("hash")),
							now,
						),
					},
					LatestURL: domain.NewGameURL(
						gameURLID17,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID18,
						values.GameImageTypeGif,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID18,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "fileTypeで絞り込みを入れても問題なし",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID20),
					Name: "Tggiblv20",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID19),
							Name:        "test19",
							Description: "test19",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID19),
									Name:        "test19",
									Description: "test19",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID19),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID18),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID19),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID19),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID20,
			fileTypes:         []values.GameFileType{values.GameFileTypeWindows},
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID19,
						"test19",
						"test19",
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID19,
						"test19",
						"test19",
						now,
					),
					LatestFiles: []*domain.GameFile{},
					LatestURL: domain.NewGameURL(
						gameURLID18,
						values.NewGameURLLink(urlLink),
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID19,
						values.GameImageTypePng,
						now,
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID19,
						values.GameVideoTypeMp4,
						now,
					),
				},
			},
		},
		{
			description: "誤ったfileTypeなのでエラー",
			beforeLauncherVersions: []LauncherVersionTable{
				{
					ID:   uuid.UUID(launcherVersionID21),
					Name: "Tggiblv21",
					QuestionnaireURL: sql.NullString{
						Valid:  true,
						String: "https://example.com",
					},
					CreatedAt: now,
					Games: []GameTable{
						{
							ID:          uuid.UUID(gameID20),
							Name:        "test20",
							Description: "test20",
							CreatedAt:   now,
							GameVersions: []GameVersionTable{
								{
									ID:          uuid.UUID(gameVersionID20),
									Name:        "test20",
									Description: "test20",
									CreatedAt:   now,
									GameFiles: []GameFileTable{
										{
											ID:         uuid.UUID(gameFileID20),
											FileTypeID: fileTypeMap[gameFileTypeJar],
											Hash:       "68617368",
											EntryPoint: "main.jar",
											CreatedAt:  now,
										},
									},
									GameURL: GameURLTable{
										ID:        uuid.UUID(gameURLID19),
										URL:       "https://example.com",
										CreatedAt: now,
									},
								},
							},
							GameImages: []GameImageTable{
								{
									ID:          uuid.UUID(gameImageID20),
									ImageTypeID: imageTypeMap[gameImageTypePng],
									CreatedAt:   now,
								},
							},
							GameVideos: []GameVideoTable{
								{
									ID:          uuid.UUID(gameVideoID20),
									VideoTypeID: videoTypeMap[gameVideoTypeMp4],
									CreatedAt:   now,
								},
							},
						},
					},
				},
			},
			launcherVersionID: launcherVersionID21,
			fileTypes:         []values.GameFileType{100},
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.beforeLauncherVersions != nil && len(testCase.beforeLauncherVersions) != 0 {
				err := db.Create(&testCase.beforeLauncherVersions).Error
				if err != nil {
					t.Fatalf("failed to create launcher version: %s", err)
				}
			}

			gameInfos, err := gameRepository.GetGameInfosByLauncherVersion(ctx, testCase.launcherVersionID, testCase.fileTypes)

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

			assert.Len(t, gameInfos, len(testCase.gameInfos))

			for i, gameInfo := range testCase.gameInfos {
				actualGameInfo := gameInfos[i]

				assert.Equal(t, gameInfo.Game.GetID(), actualGameInfo.Game.GetID())
				assert.Equal(t, gameInfo.Game.GetName(), actualGameInfo.Game.GetName())
				assert.Equal(t, gameInfo.Game.GetDescription(), actualGameInfo.Game.GetDescription())
				assert.WithinDuration(t, gameInfo.Game.GetCreatedAt(), actualGameInfo.Game.GetCreatedAt(), time.Second)

				assert.Equal(t, gameInfo.LatestVersion.GetID(), actualGameInfo.LatestVersion.GetID())
				assert.Equal(t, gameInfo.LatestVersion.GetName(), actualGameInfo.LatestVersion.GetName())
				assert.Equal(t, gameInfo.LatestVersion.GetDescription(), actualGameInfo.LatestVersion.GetDescription())
				assert.WithinDuration(t, gameInfo.LatestVersion.GetCreatedAt(), actualGameInfo.LatestVersion.GetCreatedAt(), time.Second)

				assert.Len(t, actualGameInfo.LatestFiles, len(gameInfo.LatestFiles))

				for j, gameFile := range gameInfo.LatestFiles {
					actualGameFile := actualGameInfo.LatestFiles[j]

					assert.Equal(t, gameFile.GetID(), actualGameFile.GetID())
					assert.Equal(t, gameFile.GetFileType(), actualGameFile.GetFileType())
					assert.Equal(t, gameFile.GetEntryPoint(), actualGameFile.GetEntryPoint())
					assert.Equal(t, gameFile.GetHash(), actualGameFile.GetHash())
					assert.WithinDuration(t, gameFile.GetCreatedAt(), actualGameFile.GetCreatedAt(), time.Second)
				}

				if gameInfo.LatestURL == nil {
					assert.Nil(t, actualGameInfo.LatestURL)
				} else {
					assert.Equal(t, gameInfo.LatestURL.GetID(), actualGameInfo.LatestURL.GetID())
					assert.Equal(t, gameInfo.LatestURL.GetLink(), actualGameInfo.LatestURL.GetLink())
					assert.WithinDuration(t, gameInfo.LatestURL.GetCreatedAt(), actualGameInfo.LatestURL.GetCreatedAt(), time.Second)
				}

				if gameInfo.LatestImage == nil {
					assert.Nil(t, actualGameInfo.LatestImage)
				} else {
					assert.Equal(t, gameInfo.LatestImage.GetID(), actualGameInfo.LatestImage.GetID())
					assert.Equal(t, gameInfo.LatestImage.GetType(), actualGameInfo.LatestImage.GetType())
					assert.WithinDuration(t, gameInfo.LatestImage.GetCreatedAt(), actualGameInfo.LatestImage.GetCreatedAt(), time.Second)
				}

				if gameInfo.LatestVideo == nil {
					assert.Nil(t, actualGameInfo.LatestVideo)
				} else {
					assert.Equal(t, gameInfo.LatestVideo.GetID(), actualGameInfo.LatestVideo.GetID())
					assert.Equal(t, gameInfo.LatestVideo.GetType(), actualGameInfo.LatestVideo.GetType())
					assert.WithinDuration(t, gameInfo.LatestVideo.GetCreatedAt(), actualGameInfo.LatestVideo.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}
