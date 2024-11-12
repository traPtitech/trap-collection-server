package v2

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	mockStorage "github.com/traPtitech/trap-collection-server/src/storage/mock"
	"github.com/traPtitech/trap-collection-server/testdata"
)

func Test_checkZip(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		readerFunc func(t *testing.T) io.Reader
		wantOk     bool
		isErr      bool
		err        error
	}{
		"zipファイルなのでエラー無し": {
			readerFunc: func(t *testing.T) io.Reader {
				t.Helper()

				r, err := testdata.FS.Open("a.zip")
				require.NoError(t, err)
				t.Cleanup(func() {
					r.Close()
				})

				return r
			},
			wantOk: true,
		},
		"zipファイルではないのでfalse": {
			readerFunc: func(t *testing.T) io.Reader {
				t.Helper()

				return strings.NewReader("test")
			},
			wantOk: false,
		},
	}

	gameFile := &GameFile{}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, ok, err := gameFile.checkZip(context.Background(), testCase.readerFunc(t))
			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.wantOk, ok)
		})
	}
}

func Test_checkEntryPointExist(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		entryPoint  values.GameFileEntryPoint
		zipFileName string
		result      bool
		isErr       bool
		err         error
	}{
		"特に問題ないのでエラーなし": {
			entryPoint:  values.NewGameFileEntryPoint("a/b/file"),
			zipFileName: "a.zip",
			result:      true,
		},
		"存在しないパスなのでfalse": {
			entryPoint:  values.NewGameFileEntryPoint("a/b/not_exist"),
			zipFileName: "a.zip",
			result:      false,
		},
		".で始まる相対パスはfalse": {
			entryPoint:  values.NewGameFileEntryPoint("./a/b/file"),
			zipFileName: "a.zip",
			result:      false,
		},
		"ディレクトリを指定しているのでfalse": {
			entryPoint:  values.NewGameFileEntryPoint("a/b/"),
			zipFileName: "a.zip",
			result:      false,
		},
		"ファイルのzipでもエラー無し": {
			entryPoint:  values.NewGameFileEntryPoint("b.txt"),
			zipFileName: "b.zip",
			result:      true,
		},
	}

	gameFile := &GameFile{}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b, err := testdata.FS.ReadFile(testCase.zipFileName)
			require.NoError(t, err)

			r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
			require.NoError(t, err)

			result, err := gameFile.checkEntryPointExist(context.Background(), r, testCase.entryPoint)
			if testCase.isErr {
				if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.result, result)
		})
	}
}

func TestSaveGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)

	type test struct {
		description                   string
		readerFunc                    func(t *testing.T) io.Reader
		gameID                        values.GameID
		fileType                      values.GameFileType
		entryPoint                    values.GameFileEntryPoint
		GetGameErr                    error
		executeRepositorySaveGameFile bool
		repositorySaveGameFileErr     error
		executeStorageSaveGameFile    bool
		storageSaveGameFileErr        error
		hash                          values.GameFileHash
		isErr                         bool
		err                           error
	}

	gameID := values.NewGameID()

	testdataZipReaderFunc := func(t *testing.T) io.Reader {
		t.Helper()

		r, err := testdata.FS.Open("a.zip")
		require.NoError(t, err)
		t.Cleanup(func() {
			r.Close()
		})

		return r
	}
	testdataZipHash := values.NewGameFileHashFromBytes([]byte{0x02, 0x4d, 0xc4, 0x46, 0xbe, 0x7a, 0xb9, 0x1e, 0x64, 0xb9, 0x50, 0x10, 0x2a, 0x94, 0xb7, 0xbd})

	testCases := []test{
		{
			description:                   "特に問題ないのでエラーなし",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			hash:                          testdataZipHash,
		},
		{
			description: "ゲームが存在しないのでエラー",
			readerFunc:  testdataZipReaderFunc,
			gameID:      gameID,
			fileType:    values.GameFileTypeJar,
			entryPoint:  values.NewGameFileEntryPoint("a/b/file"),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			readerFunc:  testdataZipReaderFunc,
			gameID:      gameID,
			fileType:    values.GameFileTypeJar,
			entryPoint:  values.NewGameFileEntryPoint("a/b/file"),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                   "repositoryのSaveGameFileがエラーなのでエラー",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			repositorySaveGameFileErr:     errors.New("error"),
			isErr:                         true,
		},
		{
			description:                   "storageのSaveGameFileがエラーなのでエラー",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			storageSaveGameFileErr:        errors.New("error"),
			isErr:                         true,
		},
		{
			description:                   "fileTypeがwindowsでもエラーなし",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeWindows,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			hash:                          testdataZipHash,
		},
		{
			description:                   "fileTypeがmacでもエラーなし",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeMac,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			hash:                          testdataZipHash,
		},
		{
			description:                   "entryPointが空なので、ErrInvalidEntryPoint",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint(""),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			isErr:                         true,
			err:                           service.ErrInvalidEntryPoint,
		},
		{
			description:                   "無効なentryPointなので、ErrInvalidEntryPoint",
			readerFunc:                    testdataZipReaderFunc,
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/not_exist"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			isErr:                         true,
			err:                           service.ErrInvalidEntryPoint,
		},
		{
			description:                   "zipではないので、ErrNotZipFile",
			readerFunc:                    func(t *testing.T) io.Reader { t.Helper(); return strings.NewReader("test") },
			gameID:                        gameID,
			fileType:                      values.GameFileTypeJar,
			entryPoint:                    values.NewGameFileEntryPoint("a/b/file"),
			executeRepositorySaveGameFile: true,
			executeStorageSaveGameFile:    true,
			isErr:                         true,
			err:                           service.ErrNotZipFile,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			mockGameFileStorage := mockStorage.NewGameFile(ctrl, buf)

			gameFileService := NewGameFile(
				mockDB,
				mockGameRepository,
				mockGameFileRepository,
				mockGameFileStorage,
			)

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeRepositorySaveGameFile {
				mockGameFileRepository.
					EXPECT().
					SaveGameFile(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(testCase.repositorySaveGameFileErr)
				mockGameFileStorage.
					EXPECT().
					SaveGameFile(gomock.Any(), gomock.Any()).
					Return(testCase.storageSaveGameFileErr)
			}

			var expectBytes []byte
			if testCase.readerFunc != nil {
				var err error
				expectBytes, err = io.ReadAll(testCase.readerFunc(t))
				require.NoError(t, err)
			}

			gameFile, err := gameFileService.SaveGameFile(ctx, testCase.readerFunc(t), testCase.gameID, testCase.fileType, testCase.entryPoint)

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
			assert.WithinDuration(t, time.Now(), gameFile.GetCreatedAt(), time.Second)

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
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameFileStorage := mockStorage.NewGameFile(ctrl, nil)

	gameFileService := NewGameFile(
		mockDB,
		mockGameRepository,
		mockGameFileRepository,
		mockGameFileStorage,
	)

	type test struct {
		description        string
		gameID             values.GameID
		gameFileID         values.GameFileID
		environment        values.LauncherEnvironment
		getGameErr         error
		executeGetGameFile bool
		file               *repository.GameFileInfo
		getGameFileErr     error
		executeGetTempURL  bool
		fileURL            values.GameFileTmpURL
		getTempURLErr      error
		isErr              bool
		err                error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode file: %v", err)
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()
	gameFileID4 := values.NewGameFileID()
	gameFileID5 := values.NewGameFileID()

	testCases := []test{
		{
			description:        "特に問題ないのでエラーなし",
			gameID:             gameID1,
			gameFileID:         gameFileID1,
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID1,
			},
			executeGetTempURL: true,
			fileURL:           values.NewGameFileTmpURL(urlLink),
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:        "ファイルがwindowsでもエラーなし",
			gameID:             gameID2,
			gameFileID:         gameFileID2,
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID2,
			},
			executeGetTempURL: true,
		},
		{
			description:        "ファイルがmacでもエラーなし",
			gameID:             gameID3,
			gameFileID:         gameFileID3,
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					gameFileID3,
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID3,
			},
			executeGetTempURL: true,
		},
		{
			description:        "GetGameFileがErrRecordNotFoundなのでErrInvalidGameFileID",
			gameID:             values.NewGameID(),
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			getGameFileErr:     repository.ErrRecordNotFound,
			isErr:              true,
			err:                service.ErrInvalidGameFileID,
		},
		{
			description:        "ゲームファイルに紐づくゲームIDが違うのでErrInvalidGameFileID",
			gameID:             values.NewGameID(),
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameFileID,
		},
		{
			description:        "GetGameFileがエラーなのでエラー",
			gameID:             values.NewGameID(),
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			getGameFileErr:     errors.New("error"),
			isErr:              true,
		},
		{
			description:        "GetTempURLがエラーなのでエラー",
			gameID:             gameID4,
			gameFileID:         gameFileID4,
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					gameFileID4,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID4,
			},
			executeGetTempURL: true,
			fileURL:           values.NewGameFileTmpURL(urlLink),
			getTempURLErr:     errors.New("error"),
			isErr:             true,
		},
		{
			description:        "ファイルが大きくてもエラーなし",
			gameID:             gameID5,
			gameFileID:         gameFileID5,
			environment:        *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					gameFileID5,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x70, 0x95, 0xba, 0xe0, 0x98, 0x25, 0x9e, 0xd, 0xda, 0x4b, 0x7a, 0xcc, 0x62, 0x4d, 0xe4, 0xe2}),
					time.Now(),
				),
				GameID: gameID5,
			},
			executeGetTempURL: true,
			fileURL:           values.NewGameFileTmpURL(urlLink),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameFile {
				mockGameFileRepository.
					EXPECT().
					GetGameFile(ctx, testCase.gameFileID, repository.LockTypeRecord).
					Return(testCase.file, testCase.getGameFileErr)
			}

			if testCase.executeGetTempURL {
				mockGameFileStorage.
					EXPECT().
					GetTempURL(ctx, testCase.file.GameFile, time.Minute).
					Return(testCase.fileURL, testCase.getTempURLErr)
			}

			tmpURL, err := gameFileService.GetGameFile(ctx, testCase.gameID, testCase.gameFileID)

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

			assert.Equal(t, testCase.fileURL, tmpURL)
		})
	}
}

func TestGetGameFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameFileStorage := mockStorage.NewGameFile(ctrl, nil)

	gameFileService := NewGameFile(
		mockDB,
		mockGameRepository,
		mockGameFileRepository,
		mockGameFileStorage,
	)

	type test struct {
		description         string
		gameID              values.GameID
		environment         values.LauncherEnvironment
		getGameErr          error
		executeGetGameFiles bool
		getGameFilesErr     error
		isErr               bool
		gameFiles           []*domain.GameFile
		err                 error
	}

	now := time.Now()
	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			gameID:              values.NewGameID(),
			environment:         *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFiles: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					now,
				),
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:         "ファイルが無くてもエラーなし",
			gameID:              values.NewGameID(),
			environment:         *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFiles: true,
			gameFiles:           []*domain.GameFile{},
		},
		{
			description:         "ファイルが複数でもエラーなし",
			gameID:              values.NewGameID(),
			executeGetGameFiles: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					now,
				),
				domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					now,
				),
			},
		},
		{
			description:         "ファイルがjarでもエラーなし",
			gameID:              values.NewGameID(),
			environment:         *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFiles: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					now,
				),
			},
		},
		{
			description:         "ファイルがmacでもエラーなし",
			gameID:              values.NewGameID(),
			environment:         *values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			executeGetGameFiles: true,
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					now,
				),
			},
		},
		{
			description:         "GetGameFilesがエラーなのでエラー",
			gameID:              values.NewGameID(),
			environment:         *values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFiles: true,
			getGameFilesErr:     errors.New("error"),
			isErr:               true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameFiles {
				mockGameFileRepository.
					EXPECT().
					GetGameFiles(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.gameFiles, testCase.getGameFilesErr)
			}

			gameFiles, err := gameFileService.GetGameFiles(ctx, testCase.gameID)

			if testCase.isErr {
				if testCase.err == nil {
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

			for i, gameFile := range gameFiles {
				assert.Equal(t, testCase.gameFiles[i].GetID(), gameFile.GetID())
				assert.Equal(t, testCase.gameFiles[i].GetFileType(), gameFile.GetFileType())
				assert.Equal(t, testCase.gameFiles[i].GetEntryPoint(), gameFile.GetEntryPoint())
				assert.Equal(t, testCase.gameFiles[i].GetHash(), gameFile.GetHash())
				assert.Equal(t, testCase.gameFiles[i].GetCreatedAt(), gameFile.GetCreatedAt())
			}
		})
	}
}

func TestGetGameFileMeta(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameFileStorage := mockStorage.NewGameFile(ctrl, nil)

	gameFileService := NewGameFile(
		mockDB,
		mockGameRepository,
		mockGameFileRepository,
		mockGameFileStorage,
	)

	type test struct {
		description        string
		gameID             values.GameID
		gameFileID         values.GameFileID
		environment        *values.LauncherEnvironment
		getGameErr         error
		executeGetGameFile bool
		file               *repository.GameFileInfo
		getGameFileErr     error
		isErr              bool
		err                error
	}

	gameID1 := values.NewGameID()

	testCases := []test{
		{
			description:        "特に問題ないのでエラーなし",
			gameID:             gameID1,
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID1,
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:        "windowdでもエラーなし",
			gameID:             gameID1,
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID1,
			},
		},
		{
			description:        "macでもエラーなし",
			gameID:             gameID1,
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: gameID1,
			},
		},
		{
			description:        "GetGameFileがErrRecordNotFoundなのでErrInvalidGameFileID",
			gameID:             values.NewGameID(),
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			getGameFileErr:     repository.ErrRecordNotFound,
			isErr:              true,
			err:                service.ErrInvalidGameFileID,
		},
		{
			description:        "ゲームファイルに紐づくゲームIDが違うのでErrInvalidGameFileID",
			gameID:             values.NewGameID(),
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			file: &repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					values.NewGameFileID(),
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("/path/to/file"),
					values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
					time.Now(),
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameFileID,
		},
		{
			description:        "GetGameFileがエラーなのでエラー",
			gameID:             values.NewGameID(),
			environment:        values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			executeGetGameFile: true,
			getGameFileErr:     errors.New("error"),
			isErr:              true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(ctx, testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetGameFile {
				mockGameFileRepository.
					EXPECT().
					GetGameFile(ctx, testCase.gameFileID, repository.LockTypeNone).
					Return(testCase.file, testCase.getGameFileErr)
			}

			gameFile, err := gameFileService.GetGameFileMeta(ctx, testCase.gameID, testCase.gameFileID)

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

			assert.Equal(t, testCase.file.GameFile.GetID(), gameFile.GetID())
			assert.Equal(t, testCase.file.GameFile.GetFileType(), gameFile.GetFileType())
			assert.Equal(t, testCase.file.GetEntryPoint(), gameFile.GetEntryPoint())
			assert.Equal(t, testCase.file.GetHash(), gameFile.GetHash())
			assert.WithinDuration(t, testCase.file.GameFile.GetCreatedAt(), gameFile.GetCreatedAt(), time.Second)
		})
	}
}
