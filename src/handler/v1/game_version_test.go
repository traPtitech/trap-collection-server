package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostGameVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mockConfig.NewMockApp(ctrl)
	mockAppConfig.
		EXPECT().
		FeatureV1Write().
		Return(true)
	mockGameVersionService := mock.NewMockGameVersion(ctrl)

	gameVersionHandler := NewGameVersion(mockAppConfig, mockGameVersionService)

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
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/versions", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

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

			gameVersion, err := gameVersionHandler.PostGameVersion(c, testCase.strGameID, testCase.apiGameVersion)

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

func TestGetGameVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mockConfig.NewMockApp(ctrl)
	mockAppConfig.
		EXPECT().
		FeatureV1Write().
		Return(true)
	mockGameVersionService := mock.NewMockGameVersion(ctrl)

	gameVersionHandler := NewGameVersion(mockAppConfig, mockGameVersionService)

	type test struct {
		description           string
		strGameID             string
		executeGetGameVersion bool
		gameID                values.GameID
		gameVersions          []*domain.GameVersion
		GetGameVersionErr     error
		expectGameVersions    []*openapi.GameVersion
		isErr                 bool
		err                   error
		statusCode            int
	}

	gameID := values.NewGameID()
	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	now := time.Now()

	testCases := []test{
		{
			description:           "特に問題ないのでエラーなし",
			strGameID:             uuid.UUID(gameID).String(),
			executeGetGameVersion: true,
			gameID:                gameID,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
			},
			expectGameVersions: []*openapi.GameVersion{
				{
					Id:          uuid.UUID(gameVersionID1).String(),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "gameIDがUUIDでないので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:           "GetGameVersionsがErrInvalidGameIDなので400",
			strGameID:             uuid.UUID(gameID).String(),
			executeGetGameVersion: true,
			gameID:                gameID,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
			},
			GetGameVersionErr: service.ErrInvalidGameID,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:           "GetGameVersionsがエラーなので500",
			strGameID:             uuid.UUID(gameID).String(),
			executeGetGameVersion: true,
			gameID:                gameID,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
			},
			GetGameVersionErr: errors.New("error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description:           "gameVersionが存在しなくてもエラーなし",
			strGameID:             uuid.UUID(gameID).String(),
			executeGetGameVersion: true,
			gameID:                gameID,
			gameVersions:          []*domain.GameVersion{},
			expectGameVersions:    []*openapi.GameVersion{},
		},
		{
			description:           "gameVersionが複数でもエラーなし",
			strGameID:             uuid.UUID(gameID).String(),
			executeGetGameVersion: true,
			gameID:                gameID,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					gameVersionID2,
					values.NewGameVersionName("v1.1.0"),
					values.NewGameVersionDescription("アップデート"),
					now,
				),
				domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now.Add(-time.Hour),
				),
			},
			expectGameVersions: []*openapi.GameVersion{
				{
					Id:          uuid.UUID(gameVersionID2).String(),
					Name:        "v1.1.0",
					Description: "アップデート",
					CreatedAt:   now,
				},
				{
					Id:          uuid.UUID(gameVersionID1).String(),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/version", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetGameVersion {
				mockGameVersionService.
					EXPECT().
					GetGameVersions(gomock.Any(), testCase.gameID).
					Return(testCase.gameVersions, testCase.GetGameVersionErr)
			}

			gameVersions, err := gameVersionHandler.GetGameVersion(c, testCase.strGameID)

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

			assert.Len(t, gameVersions, len(testCase.expectGameVersions))

			for i, gameVersion := range gameVersions {
				assert.Equal(t, testCase.expectGameVersions[i].Id, gameVersion.Id)
				assert.Equal(t, testCase.expectGameVersions[i].Name, gameVersion.Name)
				assert.Equal(t, testCase.expectGameVersions[i].Description, gameVersion.Description)
				assert.WithinDuration(t, testCase.expectGameVersions[i].CreatedAt, gameVersion.CreatedAt, 2*time.Second)
			}
		})
	}
}
