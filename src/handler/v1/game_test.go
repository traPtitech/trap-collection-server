package v1

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	session := NewSession("key", "secret")
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description    string
		strGameID      string
		executeGetGame bool
		game           *service.GameInfo
		GetGameErr     error
		apiGame        openapi.Game
		isErr          bool
		err            error
		statusCode     int
	}

	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	now := time.Now()

	testCases := []test{
		{
			description:    "特に問題ないのでエラーなし",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			game: &service.GameInfo{
				Game: domain.NewGame(
					gameID,
					values.NewGameName("test"),
					values.NewGameDescription("test"),
					now,
				),
				LatestVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("test"),
					values.NewGameVersionDescription("test"),
					now,
				),
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Version: &openapi.GameVersion{
					Id:          uuid.UUID(gameVersionID).String(),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "gameIDがuuidでないので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:    "ゲームが存在しないので400",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			GetGameErr:     service.ErrNoGame,
			isErr:          true,
			statusCode:     http.StatusBadRequest,
		},
		{
			description:    "GetGameがエラーなので500",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			GetGameErr:     errors.New("error"),
			isErr:          true,
			statusCode:     http.StatusInternalServerError,
		},
		{
			description:    "gameVersionがnilでもエラーなし",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			game: &service.GameInfo{
				Game: domain.NewGame(
					gameID,
					values.NewGameName("test"),
					values.NewGameDescription("test"),
					now,
				),
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any()).
					Return(testCase.game, testCase.GetGameErr)
			}

			game, err := gameHandler.GetGame(testCase.strGameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, testCase.apiGame, *game)
		})
	}
}
