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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
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
