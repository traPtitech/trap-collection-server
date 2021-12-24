package v1

import (
	"bytes"
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

func TestPostFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFile(ctrl)

	gameFileHandler := NewGameFile(mockGameFileService)

	type test struct {
		description         string
		strGameID           string
		reader              *bytes.Reader
		strFileType         string
		strEntryPoint       string
		executeSaveGameFile bool
		gameID              values.GameID
		fileType            values.GameFileType
		entryPoint          values.GameFileEntryPoint
		gameFile            *domain.GameFile
		SaveGameFileErr     error
		apiGameFile         *openapi.GameFile
		isErr               bool
		err                 error
		statusCode          int
	}

	gameID := values.NewGameID()
	gameFileID := values.NewGameFileID()

	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "jar",
			strEntryPoint:       "main.jar",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeJar,
			entryPoint:          values.NewGameFileEntryPoint("main.jar"),
			gameFile: domain.NewGameFile(
				gameFileID,
				values.GameFileTypeJar,
				values.NewGameFileEntryPoint("main.jar"),
				[]byte("a"),
				time.Now(),
			),
			apiGameFile: &openapi.GameFile{
				Id:         uuid.UUID(gameFileID).String(),
				Type:       "jar",
				EntryPoint: "main.jar",
			},
		},
		{
			description:         "fileTypeがwindowsでもエラーなし",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "windows",
			strEntryPoint:       "main.exe",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeWindows,
			entryPoint:          values.NewGameFileEntryPoint("main.exe"),
			gameFile: domain.NewGameFile(
				gameFileID,
				values.GameFileTypeWindows,
				values.NewGameFileEntryPoint("main.exe"),
				[]byte("a"),
				time.Now(),
			),
			apiGameFile: &openapi.GameFile{
				Id:         uuid.UUID(gameFileID).String(),
				Type:       "windows",
				EntryPoint: "main.exe",
			},
		},
		{
			description:         "fileTypeがmacでもエラーなし",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "mac",
			strEntryPoint:       "main.app",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeMac,
			entryPoint:          values.NewGameFileEntryPoint("main.app"),
			gameFile: domain.NewGameFile(
				gameFileID,
				values.GameFileTypeMac,
				values.NewGameFileEntryPoint("main.app"),
				[]byte("a"),
				time.Now(),
			),
			apiGameFile: &openapi.GameFile{
				Id:         uuid.UUID(gameFileID).String(),
				Type:       "mac",
				EntryPoint: "main.app",
			},
		},
		{
			description:   "fileTypeが誤っているのでエラー",
			strGameID:     uuid.UUID(gameID).String(),
			reader:        bytes.NewReader([]byte("a")),
			strFileType:   "invalid",
			strEntryPoint: "main.jar",
			isErr:         true,
			statusCode:    http.StatusBadRequest,
		},
		{
			description:   "entryPointが空文字なのでエラー",
			strGameID:     uuid.UUID(gameID).String(),
			reader:        bytes.NewReader([]byte("a")),
			strFileType:   "jar",
			strEntryPoint: "",
			isErr:         true,
			statusCode:    http.StatusBadRequest,
		},
		{
			description:   "gameIDが不正なので400",
			strGameID:     "invalid",
			reader:        bytes.NewReader([]byte("a")),
			strFileType:   "jar",
			strEntryPoint: "main.jar",
			isErr:         true,
			statusCode:    http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがErrInvalidGameIDなので400",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "jar",
			strEntryPoint:       "main.jar",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeJar,
			entryPoint:          values.NewGameFileEntryPoint("main.jar"),
			SaveGameFileErr:     service.ErrInvalidGameID,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがErrNoGameVersionなので400",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "jar",
			strEntryPoint:       "main.jar",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeJar,
			entryPoint:          values.NewGameFileEntryPoint("main.jar"),
			SaveGameFileErr:     service.ErrNoGameVersion,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがErrGameFileAlreadyExistsなので400",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "jar",
			strEntryPoint:       "main.jar",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeJar,
			entryPoint:          values.NewGameFileEntryPoint("main.jar"),
			SaveGameFileErr:     service.ErrGameFileAlreadyExists,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがエラーなので500",
			strGameID:           uuid.UUID(gameID).String(),
			reader:              bytes.NewReader([]byte("a")),
			strFileType:         "jar",
			strEntryPoint:       "main.jar",
			executeSaveGameFile: true,
			gameID:              gameID,
			fileType:            values.GameFileTypeJar,
			entryPoint:          values.NewGameFileEntryPoint("main.jar"),
			SaveGameFileErr:     errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			r := &NopCloser{testCase.reader}

			if testCase.executeSaveGameFile {
				mockGameFileService.
					EXPECT().
					SaveGameFile(gomock.Any(), r, testCase.gameID, testCase.fileType, testCase.entryPoint).
					Return(testCase.gameFile, testCase.SaveGameFileErr)
			}

			gameFile, err := gameFileHandler.PostFile(testCase.strGameID, testCase.strEntryPoint, r, testCase.strFileType)

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

			assert.Equal(t, *testCase.apiGameFile, *gameFile)
		})
	}
}

func TestGetGameFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFile(ctrl)

	gameFileHandler := NewGameFile(mockGameFileService)

	type test struct {
		description        string
		strGameID          string
		strOperatingSystem string
		executeGetGameFile bool
		gameID             values.GameID
		GetGameFileErr     error
		isErr              bool
		err                error
		statusCode         int
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:        "特に問題ないのでエラーなし",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "win32",
			executeGetGameFile: true,
			gameID:             gameID,
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:        "macでも問題なし",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "darwin",
			executeGetGameFile: true,
			gameID:             gameID,
		},
		{
			description:        "osが不正なので400",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "invalid",
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "GetGameFileがErrInvalidGameIDなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "win32",
			executeGetGameFile: true,
			gameID:             gameID,
			GetGameFileErr:     service.ErrInvalidGameID,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "GetGameFileがErrNoGameVersionなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "win32",
			executeGetGameFile: true,
			gameID:             gameID,
			GetGameFileErr:     service.ErrNoGameVersion,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "GetGameFileがErrNoGameFileなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "win32",
			executeGetGameFile: true,
			gameID:             gameID,
			GetGameFileErr:     service.ErrNoGameFile,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "GetGameFileがエラーなので500",
			strGameID:          uuid.UUID(gameID).String(),
			strOperatingSystem: "win32",
			executeGetGameFile: true,
			gameID:             gameID,
			GetGameFileErr:     errors.New("error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			r := bytes.NewReader([]byte("a"))

			if testCase.executeGetGameFile {
				mockGameFileService.
					EXPECT().
					GetGameFile(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(r, nil, testCase.GetGameFileErr)
			}

			res, err := gameFileHandler.GetGameFile(testCase.strGameID, testCase.strOperatingSystem)

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

			assert.Equal(t, r, res)
		})
	}
}
