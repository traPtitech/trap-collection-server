package v1

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	mockStorage "github.com/traPtitech/trap-collection-server/src/storage/mock"
)

func TestSaveGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFile(ctrl)

	type test struct {
		description                   string
		reader                        *bytes.Buffer
		gameID                        values.GameID
		fileType                      values.GameFileType
		entryPoint                    values.GameFileEntryPoint
		GetGameErr                    error
		executeGetLatestGameVersion   bool
		gameVersion                   *domain.GameVersion
		GetLatestGameVersionErr       error
		executeRepositorySaveGameFile bool
		repositorySaveGameFileErr     error
		executeStorageSaveGameVersion bool
		storageSaveGameVersionErr     error
		hash                          values.GameFileHash
		isErr                         bool
		err                           error
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:                 "特に問題ないのでエラーなし",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			hash:                          values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
		},
		{
			description: "ゲームが存在しないのでエラー",
			reader:      bytes.NewBufferString("test"),
			gameID:      gameID,
			fileType:    values.GameFileTypeJar,
			entryPoint:  values.NewGameFileEntryPoint("/path/to/file"),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			reader:      bytes.NewBufferString("test"),
			gameID:      gameID,
			fileType:    values.GameFileTypeJar,
			entryPoint:  values.NewGameFileEntryPoint("/path/to/file"),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                 "ゲームバージョンが存在しないのでエラー",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     repository.ErrRecordNotFound,
			isErr:                       true,
			err:                         service.ErrNoGameVersion,
		},
		{
			description:                 "GetLatestGameVersionがエラーなのでエラー",
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
		{
			description:                 "repositoryのSaveGameFileがエラーなのでエラー",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			repositorySaveGameFileErr:     errors.New("error"),
			isErr:                         true,
		},
		{
			description:                 "storageのSaveGameVersionがエラーなのでエラー",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			storageSaveGameVersionErr:     errors.New("error"),
			isErr:                         true,
		},
		{
			description:                 "fileTypeがwindowsでもエラーなし",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeWindows,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			hash:                          values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
		},
		{
			description:                 "fileTypeがmacでもエラーなし",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeMac,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			hash:                          values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
		},
		{
			description:                 "entryPointが空でもエラーなし",
			reader:                      bytes.NewBufferString("test"),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint(""),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			hash:                          values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
		},
		{
			description:                 "データが大きくてもエラーなし",
			reader:                      bytes.NewBufferString(strings.Repeat("a", 1e7)),
			gameID:                      gameID,
			fileType:                    values.GameFileTypeJar,
			entryPoint:                  values.NewGameFileEntryPoint("/path/to/file"),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameVersion: true,
			hash:                          values.NewGameFileHashFromBytes([]byte{0x70, 0x95, 0xba, 0xe0, 0x98, 0x25, 0x9e, 0xd, 0xda, 0x4b, 0x7a, 0xcc, 0x62, 0x4d, 0xe4, 0xe2}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			mockGameFileStorage := mockStorage.NewGameFile(ctrl, buf)

			gameFileService := NewGameFile(
				mockDB,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
				mockGameFileStorage,
			)

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), gameID, repository.LockTypeRecord).
					Return(testCase.gameVersion, testCase.GetLatestGameVersionErr)
			}

			if testCase.executeRepositorySaveGameFile {
				mockGameFileRepository.
					EXPECT().
					SaveGameFile(gomock.Any(), testCase.gameVersion.GetID(), gomock.Any()).
					Return(testCase.repositorySaveGameFileErr)
			}

			if testCase.executeStorageSaveGameVersion {
				mockGameFileStorage.
					EXPECT().
					SaveGameFile(gomock.Any(), gomock.Any()).
					Return(testCase.storageSaveGameVersionErr)
			}

			var expectBytes []byte
			if testCase.reader != nil {
				expectBytes = testCase.reader.Bytes()
			}

			gameFile, err := gameFileService.SaveGameFile(ctx, testCase.reader, testCase.gameID, testCase.fileType, testCase.entryPoint)

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

			assert.Equal(t, testCase.fileType, gameFile.GetFileType())
			assert.Equal(t, testCase.entryPoint, gameFile.GetEntryPoint())
			assert.Equal(t, testCase.hash, gameFile.GetHash())
			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFile(ctrl)

	type test struct {
		description                  string
		fileContent                  *bytes.Buffer
		gameID                       values.GameID
		environment                  *values.LauncherEnvironment
		GetGameErr                   error
		executeGetLatestGameVersion  bool
		gameVersion                  *domain.GameVersion
		GetLatestGameVersionErr      error
		executeRepositoryGetGameFile bool
		gameFiles                    []*domain.GameFile
		repositoryGetGameFileErr     error
		gameFile                     *domain.GameFile
		executeStorageGetGameFile    bool
		storageGetGameFileErr        error
		isErr                        bool
		err                          error
	}

	gameID := values.NewGameID()
	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()

	testCases := []test{
		{
			description:                 "特に問題ないのでエラーなし",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID1,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
		},
		{
			description: "ゲームが存在しないのでエラー",
			fileContent: bytes.NewBufferString("test"),
			gameID:      gameID,
			environment: values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			fileContent: bytes.NewBufferString("test"),
			gameID:      gameID,
			environment: values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                 "ゲームバージョンが存在しないのでエラー",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     repository.ErrRecordNotFound,
			isErr:                       true,
			err:                         service.ErrNoGameVersion,
		},
		{
			description:                 "GetLatestGameVersionがエラーなのでエラー",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
		{
			description:                 "repositoryのGetGameFileがエラーなのでエラー",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			repositoryGetGameFileErr:     errors.New("error"),
			isErr:                        true,
		},
		{
			description:                 "storageのGetGameFileがエラーなのでエラー",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID1,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
			storageGetGameFileErr:     errors.New("error"),
			isErr:                     true,
		},
		{
			description:                 "ゲームファイルが存在しないのでエラー",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles:                    []*domain.GameFile{},
			isErr:                        true,
			err:                          service.ErrNoGameFile,
		},
		{
			description:                 "windows用のファイルが存在すればjarよりそちらを優先する",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeWindows,
					"/path/to/game.exe",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeWindows,
				"/path/to/game.exe",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
		},
		{
			description:                 "順番に関わらずwindows用のファイルが存在すればjarよりそちらを優先する",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeWindows,
					"/path/to/game.exe",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeWindows,
				"/path/to/game.exe",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
		},
		{
			description:                 "mac用のファイルが存在すればjarよりそちらを優先する",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeMac,
					"/path/to/game.app",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeMac,
				"/path/to/game.app",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
		},
		{
			description:                 "順番に関わらずmac用のファイルが存在すればjarよりそちらを優先する",
			fileContent:                 bytes.NewBufferString("test"),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeMac,
					"/path/to/game.app",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID2,
				values.GameFileTypeMac,
				"/path/to/game.app",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
			),
			executeStorageGetGameFile: true,
		},
		{
			description:                 "ファイルが大きくても問題なし",
			fileContent:                 bytes.NewBufferString(strings.Repeat("a", 1e7)),
			gameID:                      gameID,
			environment:                 values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeRepositoryGetGameFile: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					values.NewGameFileHashFromBytes([]byte{0x70, 0x95, 0xba, 0xe0, 0x98, 0x25, 0x9e, 0xd, 0xda, 0x4b, 0x7a, 0xcc, 0x62, 0x4d, 0xe4, 0xe2}),
				),
			},
			gameFile: domain.NewGameFile(
				gameFileID1,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x70, 0x95, 0xba, 0xe0, 0x98, 0x25, 0x9e, 0xd, 0xda, 0x4b, 0x7a, 0xcc, 0x62, 0x4d, 0xe4, 0xe2}),
			),
			executeStorageGetGameFile: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameFileStorage := mockStorage.NewGameFile(ctrl, testCase.fileContent)

			gameFileService := NewGameFile(
				mockDB,
				mockGameRepository,
				mockGameVersionRepository,
				mockGameFileRepository,
				mockGameFileStorage,
			)

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), gameID, repository.LockTypeNone).
					Return(testCase.gameVersion, testCase.GetLatestGameVersionErr)
			}

			if testCase.executeRepositoryGetGameFile {
				mockGameFileRepository.
					EXPECT().
					GetGameFiles(gomock.Any(), testCase.gameVersion.GetID(), gomock.Any()).
					Return(testCase.gameFiles, testCase.repositoryGetGameFileErr)
			}

			if testCase.executeStorageGetGameFile {
				mockGameFileStorage.
					EXPECT().
					GetGameFile(gomock.Any(), gomock.Any()).
					Return(testCase.storageGetGameFileErr)
			}

			var expectBytes []byte
			if testCase.fileContent != nil {
				expectBytes = testCase.fileContent.Bytes()
			}

			buf := bytes.NewBuffer(nil)
			gameFile, err := gameFileService.GetGameFile(ctx, buf, testCase.gameID, testCase.environment)

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

			assert.Equal(t, *testCase.gameFile, *gameFile)
			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}
