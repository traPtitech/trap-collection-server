package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description     string
		name            values.GameName
		gameDescription values.GameDescription
		SaveGameErr     error
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "特に問題ないのでエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "nameが空でもエラーなし",
			name:            values.GameName(""),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "descriptionが空でもエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription(""),
		},
		{
			description:     "CreateGameがエラーなのでエラー",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			SaveGameErr:     errors.New("test"),
			isErr:           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				SaveGame(gomock.Any(), gomock.Any()).
				Return(testCase.SaveGameErr)

			game, err := gameVersionService.CreateGame(ctx, testCase.name, testCase.gameDescription)

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

			assert.Equal(t, testCase.name, game.GetName())
			assert.Equal(t, testCase.gameDescription, game.GetDescription())
			assert.WithinDuration(t, time.Now(), game.GetCreatedAt(), time.Second)
		})
	}
}

func TestUpdateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description       string
		gameID            values.GameID
		name              values.GameName
		gameDescription   values.GameDescription
		game              *domain.Game
		GetGameErr        error
		executeUpdateGame bool
		UpdateGameErr     error
		isErr             bool
		err               error
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:     "特に問題ないのでエラーなし",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "nameの変更なしでもエラーなし",
			gameID:          gameID,
			name:            values.GameName("before"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "descriptionの変更なしでもエラーなし",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("before"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "変更なしでも問題なし",
			gameID:          gameID,
			name:            values.GameName("before"),
			gameDescription: values.GameDescription("before"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
		},
		{
			description:     "ゲームが存在しないのでErrNoGame",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			GetGameErr:      repository.ErrRecordNotFound,
			isErr:           true,
			err:             service.ErrNoGame,
		},
		{
			description:     "GetGameがエラーなのでエラー",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			GetGameErr:      errors.New("error"),
			isErr:           true,
		},
		{
			description:     "UpdateGameがエラーなのでエラー",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
			UpdateGameErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(testCase.game, testCase.GetGameErr)

			if testCase.executeUpdateGame {
				mockGameRepository.
					EXPECT().
					UpdateGame(gomock.Any(), testCase.game).
					Return(testCase.UpdateGameErr)
			}

			game, err := gameVersionService.UpdateGame(ctx, testCase.gameID, testCase.name, testCase.gameDescription)

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

			assert.Equal(t, testCase.game, game)

			assert.Equal(t, testCase.game.GetID(), game.GetID())
			assert.Equal(t, testCase.name, game.GetName())
			assert.Equal(t, testCase.gameDescription, game.GetDescription())
			assert.Equal(t, testCase.game.GetCreatedAt(), game.GetCreatedAt())
		})
	}
}
