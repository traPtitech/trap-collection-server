package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetGameVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVersionService := mock.NewMockGameVersionV2(ctrl)

	gameVersionHandler := NewGameVersion(mockGameVersionService)

	type test struct {
		description        string
		gameID             values.GameID
		limit              *int
		offset             *int
		expectLimit        uint
		expectOffset       uint
		num                uint
		gameVersions       []*service.GameVersionInfo
		GetGameVersionErr  error
		expectGameVersions openapi.GetGameVersionsResponse
		isErr              bool
		err                error
		statusCode         int
	}

	one := 1
	gameID := values.NewGameID()
	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID1UUID := uuid.UUID(fileID1)
	fileID2UUID := uuid.UUID(fileID2)
	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	strURL := "https://example.com"
	urlLink, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
		return
	}
	urlValue := values.NewGameURLLink(urlLink)
	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(urlValue),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Url:         &strURL,
						ImageID:     uuid.UUID(imageID1),
						VideoID:     uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description:       "GetGameVersionsがErrInvalidGameIDなので404",
			gameID:            gameID,
			GetGameVersionErr: service.ErrInvalidGameID,
			isErr:             true,
			statusCode:        http.StatusNotFound,
		},
		{
			description:       "GetGameVersionsがErrInvalidLimitなので400",
			gameID:            gameID,
			GetGameVersionErr: service.ErrInvalidLimit,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "GetGameVersionsがエラーなので500",
			gameID:            gameID,
			GetGameVersionErr: errors.New("error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description:  "gameVersionが存在しなくてもエラーなし",
			gameID:       gameID,
			num:          0,
			gameVersions: []*service.GameVersionInfo{},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num:      0,
				Versions: []openapi.GameVersion{},
			},
		},
		{
			description: "gameVersionが複数でもエラーなし",
			gameID:      gameID,
			num:         2,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID2,
						values.NewGameVersionName("v1.1.0"),
						values.NewGameVersionDescription("アップデート"),
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(urlValue),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now.Add(-time.Hour),
					),
					Assets: &service.Assets{
						URL: types.NewOption(urlValue),
					},
					ImageID: imageID2,
					VideoID: videoID2,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 2,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID2),
						Name:        "v1.1.0",
						Description: "アップデート",
						CreatedAt:   now,
						Url:         &strURL,
						ImageID:     uuid.UUID(imageID1),
						VideoID:     uuid.UUID(videoID1),
					},
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now.Add(-time.Hour),
						Url:         &strURL,
						ImageID:     uuid.UUID(imageID2),
						VideoID:     uuid.UUID(videoID2),
					},
				},
			},
		},
		{
			description: "windowsのファイルが存在しても問題なし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID1),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
						ImageID: uuid.UUID(imageID1),
						VideoID: uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description: "macのファイルが存在しても問題なし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						Mac: types.NewOption(fileID1),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Files: &openapi.GameVersionFiles{
							Darwin: &fileID1UUID,
						},
						ImageID: uuid.UUID(imageID1),
						VideoID: uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description: "jarのファイルが存在しても問題なし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						Jar: types.NewOption(fileID1),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Files: &openapi.GameVersionFiles{
							Jar: &fileID1UUID,
						},
						ImageID: uuid.UUID(imageID1),
						VideoID: uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description: "ファイルが複数存在しても問題なし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID1),
						Mac:     types.NewOption(fileID2),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Files: &openapi.GameVersionFiles{
							Win32:  &fileID1UUID,
							Darwin: &fileID2UUID,
						},
						ImageID: uuid.UUID(imageID1),
						VideoID: uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description: "ファイルとurlが両方存在しても問題なし",
			gameID:      gameID,
			num:         1,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						URL:     types.NewOption(urlValue),
						Windows: types.NewOption(fileID1),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 1,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Files: &openapi.GameVersionFiles{
							Win32: &fileID1UUID,
						},
						Url:     &strURL,
						ImageID: uuid.UUID(imageID1),
						VideoID: uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description: "limitが存在しても問題なし",
			gameID:      gameID,
			limit:       &one,
			expectLimit: 1,
			num:         2,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(urlValue),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 2,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Url:         &strURL,
						ImageID:     uuid.UUID(imageID1),
						VideoID:     uuid.UUID(videoID1),
					},
				},
			},
		},
		{
			description:  "offsetが存在しても問題なし",
			gameID:       gameID,
			limit:        &one,
			offset:       &one,
			expectLimit:  1,
			expectOffset: 1,
			num:          2,
			gameVersions: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(urlValue),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
			expectGameVersions: openapi.GetGameVersionsResponse{
				Num: 2,
				Versions: []openapi.GameVersion{
					{
						Id:          uuid.UUID(gameVersionID1),
						Name:        "v1.0.0",
						Description: "リリース",
						CreatedAt:   now,
						Url:         &strURL,
						ImageID:     uuid.UUID(imageID1),
						VideoID:     uuid.UUID(videoID1),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/games/%s/versions", uuid.UUID(testCase.gameID)), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockGameVersionService.
				EXPECT().
				GetGameVersions(gomock.Any(), testCase.gameID, &service.GetGameVersionsParams{
					Limit:  testCase.expectLimit,
					Offset: testCase.expectOffset,
				}).
				Return(testCase.num, testCase.gameVersions, testCase.GetGameVersionErr)

			err := gameVersionHandler.GetGameVersion(c, uuid.UUID(testCase.gameID), openapi.GetGameVersionParams{
				Limit:  testCase.limit,
				Offset: testCase.offset,
			})

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
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, http.StatusOK, rec.Code)

			var res openapi.GetGameVersionsResponse
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.expectGameVersions.Num, res.Num)
			assert.Len(t, res.Versions, len(testCase.expectGameVersions.Versions))
			for i, gameVersion := range res.Versions {
				assert.Equal(t, testCase.expectGameVersions.Versions[i].Id, gameVersion.Id)
				assert.Equal(t, testCase.expectGameVersions.Versions[i].Name, gameVersion.Name)
				assert.Equal(t, testCase.expectGameVersions.Versions[i].Description, gameVersion.Description)
				assert.WithinDuration(t, testCase.expectGameVersions.Versions[i].CreatedAt, gameVersion.CreatedAt, 2*time.Second)
				assert.Equal(t, testCase.expectGameVersions.Versions[i].Url, gameVersion.Url)
				assert.Equal(t, testCase.expectGameVersions.Versions[i].ImageID, gameVersion.ImageID)
				assert.Equal(t, testCase.expectGameVersions.Versions[i].VideoID, gameVersion.VideoID)
			}
		})
	}
}

func TestPostGameVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVersionService := mock.NewMockGameVersionV2(ctrl)

	gameVersionHandler := NewGameVersion(mockGameVersionService)

	type test struct {
		description              string
		invalidRequest           bool
		apiGameVersion           *openapi.NewGameVersion
		executeCreateGameVersion bool
		gameID                   values.GameID
		gameVersionName          values.GameVersionName
		gameVersionDescription   values.GameVersionDescription
		imageID                  values.GameImageID
		videoID                  values.GameVideoID
		assets                   *service.Assets
		gameVersion              *service.GameVersionInfo
		CreateGameVersionErr     error
		expectGameVersion        *openapi.GameVersion
		isErr                    bool
		err                      error
		statusCode               int
	}

	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID1UUID := uuid.UUID(fileID1)
	fileID2UUID := uuid.UUID(fileID2)
	invalidURL := " https://example.com"
	strURL := "https://example.com"
	urlLink, err := url.Parse(strURL)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
		return
	}
	urlValue := values.NewGameURLLink(urlLink)
	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					URL: types.NewOption(urlValue),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
		},
		{
			description:    "requestがjsonでないので400",
			invalidRequest: true,
			isErr:          true,
			statusCode:     http.StatusBadRequest,
		},
		{
			description: "バージョン名が不適切なので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.a",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			gameID:     gameID,
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "バージョン名が空文字なので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			gameID:     gameID,
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "descriptionが空文字でもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription(""),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription(""),
					now,
				),
				Assets: &service.Assets{
					URL: types.NewOption(urlValue),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
		},
		{
			description: "urlが不適切なので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &invalidURL,
			},
			gameID:     gameID,
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがErrInvalidGameIDなので404",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			CreateGameVersionErr: service.ErrInvalidGameID,
			isErr:                true,
			statusCode:           http.StatusNotFound,
		},
		{
			description: "CreateGameVersionがErrInvalidGameImageIDなので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			CreateGameVersionErr: service.ErrInvalidGameImageID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがErrInvalidGameVideoIDなので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			CreateGameVersionErr: service.ErrInvalidGameVideoID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがErrNoAssetなので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets:                   &service.Assets{},
			CreateGameVersionErr:     service.ErrNoAsset,
			isErr:                    true,
			statusCode:               http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがエラーなので500",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL: types.NewOption(urlValue),
			},
			CreateGameVersionErr: errors.New("error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
		{
			description: "windowsでもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Windows: types.NewOption(fileID1),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					Windows: types.NewOption(fileID1),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
		},
		{
			description: "macでもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Darwin: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Mac: types.NewOption(fileID1),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					Mac: types.NewOption(fileID1),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Darwin: &fileID1UUID,
				},
			},
		},
		{
			description: "jarでもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Jar: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Jar: types.NewOption(fileID1),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					Jar: types.NewOption(fileID1),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Jar: &fileID1UUID,
				},
			},
		},
		{
			description: "ファイルが複数あってもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32:  &fileID1UUID,
					Darwin: &fileID2UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Windows: types.NewOption(fileID1),
				Mac:     types.NewOption(fileID2),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					Windows: types.NewOption(fileID1),
					Mac:     types.NewOption(fileID2),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32:  &fileID1UUID,
					Darwin: &fileID2UUID,
				},
			},
		},
		{
			description: "ファイルとurlが両方あってもエラーなし",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				URL:     types.NewOption(urlValue),
				Windows: types.NewOption(fileID1),
			},
			gameVersion: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					now,
				),
				Assets: &service.Assets{
					URL:     types.NewOption(urlValue),
					Windows: types.NewOption(fileID1),
				},
				ImageID: imageID,
				VideoID: videoID,
			},
			expectGameVersion: &openapi.GameVersion{
				Id:          uuid.UUID(gameVersionID),
				Name:        "v1.0.0",
				Description: "リリース",
				CreatedAt:   now,
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Url:         &strURL,
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
		},
		{
			description: "CreateGameVersionがErrInvalidGameFileIDなので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Windows: types.NewOption(fileID1),
			},
			CreateGameVersionErr: service.ErrInvalidGameFileID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description: "CreateGameVersionがErrInvalidFileTypeなので400",
			apiGameVersion: &openapi.NewGameVersion{
				Name:        "v1.0.0",
				Description: "リリース",
				ImageID:     uuid.UUID(imageID),
				VideoID:     uuid.UUID(videoID),
				Files: &openapi.GameVersionFiles{
					Win32: &fileID1UUID,
				},
			},
			executeCreateGameVersion: true,
			gameID:                   gameID,
			gameVersionName:          values.NewGameVersionName("v1.0.0"),
			gameVersionDescription:   values.NewGameVersionDescription("リリース"),
			imageID:                  imageID,
			videoID:                  videoID,
			assets: &service.Assets{
				Windows: types.NewOption(fileID1),
			},
			CreateGameVersionErr: service.ErrInvalidGameFileType,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			var req *http.Request
			if testCase.invalidRequest {
				reqBody := bytes.NewBuffer([]byte("invalid"))
				req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/versions", uuid.UUID(testCase.gameID)), reqBody)
				req.Header.Set("Content-Type", echo.MIMETextPlain)
			} else {
				reqBody := bytes.NewBuffer(nil)
				err := json.NewEncoder(reqBody).Encode(testCase.apiGameVersion)
				if err != nil {
					t.Fatalf("failed to encode request body: %v", err)
				}
				req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/versions", uuid.UUID(testCase.gameID)), reqBody)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			}
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
						testCase.imageID,
						testCase.videoID,
						testCase.assets,
					).Return(testCase.gameVersion, testCase.CreateGameVersionErr)
			}

			err = gameVersionHandler.PostGameVersion(c, uuid.UUID(testCase.gameID))

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
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, http.StatusCreated, rec.Code)

			var res openapi.GameVersion
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.expectGameVersion.Id, res.Id)
			assert.Equal(t, testCase.expectGameVersion.Name, res.Name)
			assert.Equal(t, testCase.expectGameVersion.Description, res.Description)
			assert.WithinDuration(t, testCase.expectGameVersion.CreatedAt, res.CreatedAt, 2*time.Second)
			assert.Equal(t, testCase.expectGameVersion.Url, res.Url)
			assert.Equal(t, testCase.expectGameVersion.ImageID, res.ImageID)
			assert.Equal(t, testCase.expectGameVersion.VideoID, res.VideoID)
		})
	}
}
