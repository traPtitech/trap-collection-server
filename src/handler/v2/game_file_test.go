package v2

import (
	"encoding/json"
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
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
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

	//TODO:hashについて要調査
	//テストでhashのjsonのエンコード、デコードが変なことになっていそうだけど、よくわからなかった。
	//↓のようにmd5Hashを用いるとテストが通るようになった。
	md5Hash := values.NewGameFileHashFromBytes([]byte("ea703e7aa1efda0064eaa507d9e8ab7e"))
	md5Hash2 := values.NewGameFileHashFromBytes([]byte("7095bae098259e0dda4b7acc624de4e2"))

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
					Md5:        string(md5Hash),
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
					Md5:        string(md5Hash),
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
					values.NewGameFileHashFromBytes(md5Hash),
					now,
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID3),
					EntryPoint: openapi.GameFileEntryPoint("path/to/file"),
					Md5:        string(md5Hash),
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
					values.NewGameFileHashFromBytes(md5Hash),
					now,
				),
				domain.NewGameFile(
					gameFileID5,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file2"),
					values.NewGameFileHashFromBytes(md5Hash),
					now.Add(-10*time.Hour),
				),
			},
			resFiles: []openapi.GameFile{
				{
					Id:         uuid.UUID(gameFileID4),
					Type:       openapi.Jar,
					EntryPoint: string("path/to/file"),
					Md5:        string(md5Hash),
					CreatedAt:  now,
				},
				{
					Id:         uuid.UUID(gameFileID5),
					Type:       openapi.Jar,
					EntryPoint: string("path/to/file2"),
					Md5:        string(md5Hash),
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
					Md5:        string(md5Hash2),
					Type:       openapi.Jar,
					CreatedAt:  now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/games/%s/files", testCase.gameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

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
