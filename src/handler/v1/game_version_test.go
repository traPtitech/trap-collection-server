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

func TestPostGameVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVersionService := mock.NewMockGameVersion(ctrl)

	gameVersionHandler := NewGameVersion(mockGameVersionService)

	type test struct {
		description              string
		strGameID                string
		apiGameVersion           *openapi.NewGameVersion
		executeCreateGameVersion bool
		gameID                   values.GameID
		gameVersionName          values.GameVersionName
		gameVersionDescription   values.GameVersionDescription
		gameVersion              *domain.GameVersion
		CreateGameVersionErr     error
		expectGameVersion        *openapi.GameVersion
		isErr                    bool
		err                      error
		statusCode               int
	}

	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			gameVersion: domain.NewGameVersion(
				gameVersionID,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID).String(),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
			},
		},
		{
			description: "gameIDがUUIDでないので400",
			strGameID:   "invalid",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "gameIDが空文字なので400",
			strGameID:   "",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "バージョン名が不適切なので400",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.a",
				Description: "リリース",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "バージョン名が空文字なので400",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "",
				Description: "リリース",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "descriptionが空文字でもエラーなし",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "",
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription(""),
			gameVersion: domain.NewGameVersion(
				gameVersionID,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription(""),
				now,
			),
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID).String(),
				Name:        "v1.0.0",
				Description: "",
				CreatedAt:   now,
			},
		},
		{
			description: "CreateGameVersionがErrInvalidGameIDなので400",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			CreateGameVersionErr:     service.ErrInvalidGameID,
			isErr:                    true,
			statusCode:               http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがエラーなので500",
			strGameID:   uuid.UUID(gameID).String(),
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			CreateGameVersionErr:     errors.New("error"),
			isErr:                    true,
			statusCode:               http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeCreateGameVersion {
				mockGameVersionService.
					EXPECT().
					CreateGameVersion(
						gomock.Any(),
						testCase.gameID,
						testCase.gameVersionName,
						testCase.gameVersionDescription,
					).Return(testCase.gameVersion, testCase.CreateGameVersionErr)
			}

			gameVersion, err := gameVersionHandler.PostGameVersion(testCase.strGameID, testCase.apiGameVersion)

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

			assert.Equal(t, testCase.expectGameVersion.Id, gameVersion.Id)
			assert.Equal(t, testCase.expectGameVersion.Name, gameVersion.Name)
			assert.Equal(t, testCase.expectGameVersion.Description, gameVersion.Description)
			assert.WithinDuration(t, testCase.expectGameVersion.CreatedAt, gameVersion.CreatedAt, 2*time.Second)
		})
	}
}
