package v2

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetGameFiles(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFileV2(ctrl)

	gameFile := NewGameFile(mockGameFileService)

	type test struct {
		description     string
		gameID          openapi.GameIDInPath
		files           []*domain.GameFile
		getGameFilesErr error
		resFiles        []openapi.GameFile
		isErr           bool
		err             error
		statusCode      int
	}

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()
	gameFileID4 := values.NewGameFileID()
	gameFileID5 := values.NewGameFileID()

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})
	md5Hash2 := values.NewGameFileHashFromBytes([]byte{0x70, 0x95, 0xba, 0xe0, 0x98, 0x25, 0x9e, 0xd, 0xda, 0x4b, 0x7a, 0xcc, 0x62, 0x4d, 0xe4, 0xe2})

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID1),
					EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
					Md5:        hex.EncodeToString(md5Hash),
					Type:       openapi.Jar,
					CreatedAt:  now,
				},
			},
		},
		{
			description: "win32でもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID2),
					EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
					Md5:        hex.EncodeToString(md5Hash),
					Type:       openapi.Win32,
					CreatedAt:  now,
				},
			},
		},
		{
			description: "darwinでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID3,
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID3),
					EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
					Md5:        hex.EncodeToString(md5Hash),
					Type:       openapi.Darwin,
					CreatedAt:  now,
				},
			},
		},
		{
			description: "jar,win32,darwinのいずれでもないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					values.NewGameFileID(),
					100,
					values.NewGameFileEntryPoint("path/to/file"),
					values.NewGameFileHashFromBytes(md5Hash),
					now,
				),
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:     "GetGameFilesがErrInvalidGameIDなので404",
			gameID:          uuid.UUID(values.NewGameID()),
			getGameFilesErr: service.ErrInvalidGameID,
			isErr:           true,
			statusCode:      http.StatusNotFound,
		},
		{
			description:     "GetGameFilesがエラーなので500",
			gameID:          uuid.UUID(values.NewGameID()),
			getGameFilesErr: errors.New("error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		{
			description: "ファイルがなくても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			files:       []*domain.GameFile{},
			resFiles:    []openapi.GameFile{},
		},
		{
			description: "ファイルが複数あっても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID4,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				domain.NewGameFile(
					gameFileID5,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file2"),
					md5Hash,
					now.Add(-10*time.Hour),
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID4),
					Type:       openapi.Jar,
					EntryPoint: string("path/to/file"),
					Md5:        hex.EncodeToString(md5Hash),
					CreatedAt:  now,
				},
				{
					Id:         uuid.UUID(gameFileID5),
					Type:       openapi.Jar,
					EntryPoint: string("path/to/file2"),
					Md5:        hex.EncodeToString(md5Hash),
					CreatedAt:  now.Add(-10 * time.Hour),
				},
			},
		},
		{
			description: "ファイルサイズが大きくてもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			files: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash2,
					now,
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID1),
					EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
					Md5:        hex.EncodeToString(md5Hash2),
					Type:       openapi.Jar,
					CreatedAt:  now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/files", testCase.gameID), nil)

			mockGameFileService.
				EXPECT().
				GetGameFiles(gomock.Any(), gomock.Any()).
				Return(testCase.files, testCase.getGameFilesErr)

			err := gameFile.GetGameFiles(c, testCase.gameID)

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

			var resFiles []openapi.GameFile
			err = json.NewDecoder(rec.Body).Decode(&resFiles)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			for i, resFile := range resFiles {
				assert.Equal(t, testCase.resFiles[i].Id, resFile.Id)
				assert.Equal(t, testCase.resFiles[i].Type, resFile.Type)
				assert.Equal(t, testCase.resFiles[i].EntryPoint, resFile.EntryPoint)
				assert.Equal(t, testCase.resFiles[i].Md5, resFile.Md5)
				assert.WithinDuration(t, testCase.resFiles[i].CreatedAt, resFile.CreatedAt, time.Second)
			}
		})
	}
}

func TestPostGameFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFileV2(ctrl)

	gameFile := NewGameFile(mockGameFileService)

	type test struct {
		description         string
		gameID              openapi.GameIDInPath
		noFileType          bool
		noEntryPoint        bool
		fileType            openapi.GameFileType
		reader              *bytes.Reader
		executeSaveGameFile bool
		file                *domain.GameFile
		saveGameFileErr     error
		resFile             openapi.GameFile
		isErr               bool
		err                 error
		statusCode          int
	}

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})

	now := time.Now()
	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			gameID:              uuid.UUID(values.NewGameID()),
			fileType:            openapi.Jar,
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			file: domain.NewGameFile(
				gameFileID1,
				values.GameFileTypeJar,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID1),
				Type:       openapi.Jar,
				EntryPoint: string("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			description:         "win32でもエラーなし",
			gameID:              uuid.UUID(values.NewGameID()),
			fileType:            openapi.Win32,
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			file: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeWindows,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID2),
				Type:       openapi.Win32,
				EntryPoint: string("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			description:         "macでもエラーなし",
			gameID:              uuid.UUID(values.NewGameID()),
			fileType:            openapi.Darwin,
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			file: domain.NewGameFile(
				gameFileID3,
				values.GameFileTypeMac,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID3),
				Type:       openapi.Darwin,
				EntryPoint: string("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			// serviceが正しく動作していればあり得ないが、念のため確認
			description:         "win32,darwin,jarのいずれでもないので400",
			gameID:              uuid.UUID(values.NewGameID()),
			fileType:            "invalid",
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: false,
			file: domain.NewGameFile(
				values.NewGameFileID(),
				100,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがErrInvalidGameIDなので404",
			fileType:            openapi.Jar,
			gameID:              uuid.UUID(values.NewGameID()),
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			saveGameFileErr:     service.ErrInvalidGameID,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "SaveGameFileがErrNotZipFileなので400",
			fileType:            openapi.Jar,
			gameID:              uuid.UUID(values.NewGameID()),
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			saveGameFileErr:     service.ErrNotZipFile,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがErrInvalidEntryPointなので400",
			fileType:            openapi.Jar,
			gameID:              uuid.UUID(values.NewGameID()),
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			saveGameFileErr:     service.ErrInvalidEntryPoint,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "SaveGameFileがエラーなので500",
			fileType:            openapi.Jar,
			gameID:              uuid.UUID(values.NewGameID()),
			reader:              bytes.NewReader([]byte("test")),
			executeSaveGameFile: true,
			saveGameFileErr:     errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
		{
			description: "contentがrequest bodyにないので400",
			fileType:    openapi.Jar,
			gameID:      uuid.UUID(values.NewGameID()),
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:  "entryPointがrequest bodyにないので400",
			gameID:       uuid.UUID(values.NewGameID()),
			fileType:     openapi.Jar,
			reader:       bytes.NewReader([]byte("test")),
			noEntryPoint: true,
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		{
			description: "typeがrequest bodyにないので400",
			gameID:      uuid.UUID(values.NewGameID()),
			fileType:    openapi.Jar,
			reader:      bytes.NewReader([]byte("test")),
			noFileType:  true,
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			formDatas := []testFormData{}
			if testCase.reader != nil {
				formDatas = append(formDatas, testFormData{
					fieldName: "content",
					fileName:  "content",
					body:      testCase.reader,
					isFile:    true,
				})
			}
			if !testCase.noEntryPoint {
				formDatas = append(formDatas, testFormData{
					fieldName: "entryPoint",
					value:     "path/to/file",
					isFile:    false,
				})
			}
			if !testCase.noFileType {
				formDatas = append(formDatas, testFormData{
					fieldName: "type",
					value:     string(testCase.fileType),
					isFile:    false,
				})
			}
			c, _, rec := setupTestRequest(t, http.MethodPost, fmt.Sprintf("/api/v2/games/%s/files", testCase.gameID),
				withMultipartFormDataBody(t, formDatas))

			if testCase.executeSaveGameFile {
				mockGameFileService.
					EXPECT().
					SaveGameFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(testCase.file, testCase.saveGameFileErr)
			}

			err := gameFile.PostGameFile(c, testCase.gameID)

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

			var resFile openapi.GameFile
			err = json.NewDecoder(rec.Body).Decode(&resFile)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resFile.Id, resFile.Id)
			assert.Equal(t, testCase.resFile.Type, resFile.Type)
			assert.Equal(t, testCase.resFile.EntryPoint, resFile.EntryPoint)
			assert.Equal(t, testCase.resFile.Md5, resFile.Md5)
			assert.WithinDuration(t, testCase.resFile.CreatedAt, resFile.CreatedAt, time.Second)
		})
	}
}

func TestGetGameFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFileV2(ctrl)

	gameFile := NewGameFile(mockGameFileService)

	type test struct {
		description    string
		gameID         openapi.GameIDInPath
		gameFileID     openapi.GameFileIDInPath
		tmpURL         values.GameFileTmpURL
		getGameFileErr error
		resLocation    string
		isErr          bool
		err            error
		statusCode     int
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode file: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			gameFileID:  uuid.UUID(values.NewGameFileID()),
			tmpURL:      values.NewGameFileTmpURL(urlLink),
			resLocation: "https://example.com",
		},
		{
			description:    "GetGameFileがErrInvalidGameIDなので404",
			gameID:         uuid.UUID(values.NewGameID()),
			gameFileID:     uuid.UUID(values.NewGameFileID()),
			getGameFileErr: service.ErrInvalidGameID,
			isErr:          true,
			statusCode:     http.StatusNotFound,
		},
		{
			description:    "GetGameFileがErrInvalidGameFileIDなので404",
			gameID:         uuid.UUID(values.NewGameID()),
			gameFileID:     uuid.UUID(values.NewGameFileID()),
			getGameFileErr: service.ErrInvalidGameFileID,
			isErr:          true,
			statusCode:     http.StatusNotFound,
		},
		{
			description:    "GetGameFileがエラーなので500",
			gameID:         uuid.UUID(values.NewGameID()),
			gameFileID:     uuid.UUID(values.NewGameFileID()),
			getGameFileErr: errors.New("error"),
			isErr:          true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/files", testCase.gameID), nil)

			mockGameFileService.
				EXPECT().
				GetGameFile(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.tmpURL, testCase.getGameFileErr)

			err := gameFile.GetGameFile(c, testCase.gameID, testCase.gameFileID)

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

			assert.Equal(t, http.StatusSeeOther, rec.Code)

			assert.Equal(t, testCase.resLocation, rec.Header().Get("Location"))
		})
	}
}

func TestGetGameFileMeta(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameFileService := mock.NewMockGameFileV2(ctrl)

	gameFile := NewGameFile(mockGameFileService)

	type test struct {
		description        string
		gameID             openapi.GameIDInPath
		gameFileID         openapi.GameFileIDInPath
		file               *domain.GameFile
		getGameFileMetaErr error
		resFile            openapi.GameFile
		isErr              bool
		err                error
		statusCode         int
	}

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			file: domain.NewGameFile(
				gameFileID1,
				values.GameFileTypeJar,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID1),
				Type:       openapi.Jar,
				EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			description: "win32でもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			file: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeWindows,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID2),
				Type:       openapi.Win32,
				EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			description: "gifでもエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			file: domain.NewGameFile(
				gameFileID3,
				values.GameFileTypeMac,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			resFile: openapi.GameFile{
				Id:         uuid.UUID(gameFileID3),
				Type:       openapi.Darwin,
				EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
				Md5:        openapi.GameFileMd5(hex.EncodeToString(md5Hash)),
				CreatedAt:  now,
			},
		},
		{
			description: "jpeg,png,gifのいずれでもないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			file: domain.NewGameFile(
				gameFileID3,
				100,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:        "GetGameFileMetaがErrInvalidGameIDなので404",
			gameID:             uuid.UUID(values.NewGameID()),
			getGameFileMetaErr: service.ErrInvalidGameID,
			isErr:              true,
			statusCode:         http.StatusNotFound,
		},
		{
			description:        "GetGameFileMetaがErrInvalidGameFileIDなので404",
			gameID:             uuid.UUID(values.NewGameID()),
			getGameFileMetaErr: service.ErrInvalidGameFileID,
			isErr:              true,
			statusCode:         http.StatusNotFound,
		},
		{
			description:        "GetGameFilesがエラーなので500",
			gameID:             uuid.UUID(values.NewGameID()),
			getGameFileMetaErr: errors.New("error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/api/v2/games/%s/files", testCase.gameID), nil)

			mockGameFileService.
				EXPECT().
				GetGameFileMeta(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(testCase.file, testCase.getGameFileMetaErr)

			err := gameFile.GetGameFileMeta(c, testCase.gameID, testCase.gameFileID)

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

			var resFile openapi.GameFile
			err = json.NewDecoder(rec.Body).Decode(&resFile)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.resFile.Id, resFile.Id)
			assert.Equal(t, testCase.resFile.Type, resFile.Type)
			assert.Equal(t, testCase.resFile.EntryPoint, resFile.EntryPoint)
			assert.Equal(t, testCase.resFile.Md5, resFile.Md5)
			assert.WithinDuration(t, testCase.resFile.CreatedAt, resFile.CreatedAt, time.Second)
		})
	}
}
