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
