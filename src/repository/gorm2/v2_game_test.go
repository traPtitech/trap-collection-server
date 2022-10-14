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
			beforeGames: []migrate.GameTable2{
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
			beforeGames: []migrate.GameTable2{
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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			game: domain.NewGame(
				gameID1,
				"test2",
				"test2",
				now,
			),
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			afterGames: []migrate.GameTable2{
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
			beforeGames: []migrate.GameTable2{
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
			afterGames: []migrate.GameTable2{
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
					Delete(&migrate.GameTable2{}).Error
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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
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
				},
			},
		},
		{
			description: "別のゲームが存在してもエラーなし",
			gameID:      gameID1,
			beforeGames: []migrate.GameTable2{
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
					Delete(&migrate.GameTable2{}).Error
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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			GameTable: []migrate.GameTable2{
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
			GameTable: []migrate.GameTable2{
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
			GameTable: []migrate.GameTable2{
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
		beforeGames []migrate.GameTable2
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
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
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
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{},
			games:       []*domain.Game{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
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
			beforeGames: []migrate.GameTable2{
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
			description: "offsetだけなのでエラー", //これはserviceで除かれるはず
			limit:       0,
			offset:      1,
			beforeGames: []migrate.GameTable2{
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
			err:   repository.ErrBadLimitAndOffset,
		},
		{
			description: "limitとoffset両方が設定されてもエラーなし",
			limit:       1,
			offset:      1,
			beforeGames: []migrate.GameTable2{
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
			description: "limitが0より小さいのでエラー",
			limit:       -2,
			offset:      0,
			beforeGames: []migrate.GameTable2{
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
					Delete(&migrate.GameTable2{}).Error
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

			games, n, err := gameRepository.GetGames(ctx, testCase.limit, testCase.offset)

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

func TestGetGamesByUserV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameRepository := NewGameV2(testDB)

	type test struct {
		description        string
		userID             values.TraPMemberID
		limit              int
		offset             int
		beforeGames        []migrate.GameTable2
		expectedGameNumber int
		games              []*domain.Game
		isErr              bool
		err                error
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
	gameID10 := values.NewGameID()
	gameID11 := values.NewGameID()
	gameID12 := values.NewGameID()
	gameID13 := values.NewGameID()
	gameID14 := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())
	userID5 := values.NewTrapMemberID(uuid.New())
	userID6 := values.NewTrapMemberID(uuid.New())
	userID7 := values.NewTrapMemberID(uuid.New())
	userID8 := values.NewTrapMemberID(uuid.New())
	userID9 := values.NewTrapMemberID(uuid.New())
	userID10 := values.NewTrapMemberID(uuid.New())
	userID11 := values.NewTrapMemberID(uuid.New())

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

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			userID:      userID1,
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID1),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 1,
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
			description: "ゲームが存在しなくてもエラーなし",
			userID:      userID2,
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{},
			games:       []*domain.Game{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			userID:      userID3,
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID3),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test3",
					Description: "test3",
					CreatedAt:   now.Add(-time.Hour),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID3),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 2,
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
			description: "他のユーザーのゲームは取得しない",
			userID:      userID4,
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test4",
					Description: "test4",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID5),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 0,
			games:              []*domain.Game{},
		},
		{
			description: "collaboratorでもゲームを取得できる",
			userID:      userID6,
			limit:       0,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test5",
					Description: "test5",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID6),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeCollaborator],
						},
					},
				},
			},
			expectedGameNumber: 1,
			games: []*domain.Game{
				domain.NewGame(
					gameID5,
					"test5",
					"test5",
					now,
				),
			},
		},
		{
			description: "削除されたゲームは取得しない",
			userID:      userID7,
			limit:       0,
			offset:      0,
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
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID7),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 0,
			games:              []*domain.Game{},
		},
		{
			description: "limitが0より小さいのでエラー",
			userID:      userID8,
			limit:       -2,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test7",
					Description: "test7",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID8),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrNegativeLimit,
		},
		{
			description: "limitを設定してもエラーなし",
			userID:      userID9,
			limit:       1,
			offset:      0,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test8",
					Description: "test8",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID9),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID9),
					Name:        "test9",
					Description: "test9",
					CreatedAt:   now.Add(-time.Hour),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID9),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 2,
			games: []*domain.Game{
				domain.NewGame(
					gameID8,
					"test8",
					"test8",
					now,
				),
			},
		},
		{
			description: "offsetだけなのでエラー", //これはserviceで除かれるはず
			userID:      userID10,
			limit:       0,
			offset:      1,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID10),
					Name:        "test10",
					Description: "test10",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID10),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID11),
					Name:        "test11",
					Description: "test11",
					CreatedAt:   now.Add(-time.Hour),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID10),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			isErr: true,
			err:   repository.ErrBadLimitAndOffset,
		},
		{
			description: "limitとoffset両方設定してもエラーなし",
			userID:      userID11,
			limit:       1,
			offset:      1,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID12),
					Name:        "test12",
					Description: "test12",
					CreatedAt:   now,
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID11),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID13),
					Name:        "test13",
					Description: "test13",
					CreatedAt:   now.Add(-time.Hour),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID11),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
				{
					ID:          uuid.UUID(gameID14),
					Name:        "test14",
					Description: "test14",
					CreatedAt:   now.Add(-time.Hour * 2),
					GameManagementRoles: []migrate.GameManagementRoleTable{
						{
							UserID:     uuid.UUID(userID11),
							RoleTypeID: roleTypeMap[gameManagementRoleTypeAdministrator],
						},
					},
				},
			},
			expectedGameNumber: 3,
			games: []*domain.Game{
				domain.NewGame(
					gameID13,
					"test13",
					"test13",
					now.Add(-time.Hour),
				),
			},
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

			games, n, err := gameRepository.GetGamesByUser(ctx, testCase.userID, testCase.limit, testCase.offset)

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
			assert.Equal(t, testCase.expectedGameNumber, n)

			for i, game := range testCase.games {
				assert.Equal(t, game.GetID(), games[i].GetID())
				assert.Equal(t, game.GetName(), games[i].GetName())
				assert.Equal(t, game.GetDescription(), games[i].GetDescription())
				assert.WithinDuration(t, game.GetCreatedAt(), games[i].GetCreatedAt(), time.Second)
			}
		})
	}
}
