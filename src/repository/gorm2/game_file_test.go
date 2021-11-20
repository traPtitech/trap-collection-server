package gorm2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"gorm.io/gorm"
)

func TestSetupFileTypeTable(t *testing.T) {
	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		description     string
		beforeFileTypes []string
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "何も存在しない場合問題なし",
			beforeFileTypes: []string{},
		},
		{
			description: "1つのみ存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
			},
		},
		{
			description: "2つ存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
				gameFileTypeWindows,
			},
		},
		{
			description: "全て存在する場合問題なし",
			beforeFileTypes: []string{
				gameFileTypeJar,
				gameFileTypeWindows,
				gameFileTypeMac,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				err := db.
					Session(&gorm.Session{
						AllowGlobalUpdate: true,
					}).
					Delete(&GameFileTypeTable{}).Error
				if err != nil {
					t.Fatalf("failed to delete role type table: %+v\n", err)
				}
			}()

			if len(testCase.beforeFileTypes) != 0 {
				fileTypes := make([]*GameFileTypeTable, 0, len(testCase.beforeFileTypes))
				for _, fileType := range testCase.beforeFileTypes {
					fileTypes = append(fileTypes, &GameFileTypeTable{
						Name: fileType,
					})
				}

				err := db.Create(fileTypes).Error
				if err != nil {
					t.Fatalf("failed to setup role type table: %+v\n", err)
				}
			}

			err := setupFileTypeTable(db)

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

			var fileTypes []*GameFileTypeTable
			err = db.
				Select("name").
				Find(&fileTypes).Error
			if err != nil {
				t.Fatalf("failed to get role type table: %+v\n", err)
			}

			fileTypeNames := make([]string, 0, len(fileTypes))
			for _, fileType := range fileTypes {
				fileTypeNames = append(fileTypeNames, fileType.Name)
			}

			assert.ElementsMatch(t, []string{
				gameFileTypeJar,
				gameFileTypeWindows,
				gameFileTypeMac,
			}, fileTypeNames)
		})
	}
}

func TestSaveGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository, err := NewGameFile(testDB)
	if err != nil {
		t.Fatalf("failed to create game management role repository: %+v\n", err)
	}

	type test struct {
		description   string
		gameVersionID values.GameVersionID
		gameFile      *domain.GameFile
		beforeFiles   []GameFileTable
		expectFiles   []GameFileTable
		isErr         bool
		err           error
	}

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()
	fileID8 := values.NewGameFileID()
	fileID9 := values.NewGameFileID()
	fileID10 := values.NewGameFileID()

	var fileTypes []*GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&fileTypes).Error
	if err != nil {
		t.Fatalf("failed to get file types: %v\n", err)
	}

	fileTypeMap := make(map[string]int, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypeMap[fileType.Name] = fileType.ID
	}

	testCases := []test{
		{
			description:   "特に問題ないので問題なし",
			gameVersionID: gameVersionID1,
			gameFile: domain.NewGameFile(
				fileID1,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID1),
					GameVersionID: uuid.UUID(gameVersionID1),
					FileTypeID:    fileTypeMap[gameFileTypeJar],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.jar",
				},
			},
		},
		{
			description:   "windowsでも問題なし",
			gameVersionID: gameVersionID2,
			gameFile: domain.NewGameFile(
				fileID2,
				values.GameFileTypeWindows,
				"/path/to/game.exe",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID2),
					GameVersionID: uuid.UUID(gameVersionID2),
					FileTypeID:    fileTypeMap[gameFileTypeWindows],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.exe",
				},
			},
		},
		{
			description:   "macでも問題なし",
			gameVersionID: gameVersionID3,
			gameFile: domain.NewGameFile(
				fileID3,
				values.GameFileTypeMac,
				"/path/to/game.app",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID3),
					GameVersionID: uuid.UUID(gameVersionID3),
					FileTypeID:    fileTypeMap[gameFileTypeMac],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.app",
				},
			},
		},
		{
			description:   "想定外のファイルの種類なのでエラー",
			gameVersionID: gameVersionID4,
			gameFile: domain.NewGameFile(
				fileID4,
				100,
				"/path/to/game.jar",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{},
			expectFiles: []GameFileTable{},
			isErr:       true,
		},
		{
			description:   "既に別の種類のファイルが存在しても問題なし",
			gameVersionID: gameVersionID5,
			gameFile: domain.NewGameFile(
				fileID5,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID6),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[gameFileTypeWindows],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.exe",
				},
			},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID6),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[gameFileTypeWindows],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.exe",
				},
				{
					ID:            uuid.UUID(fileID5),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[gameFileTypeJar],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.jar",
				},
			},
		},
		{
			// 実際には発生しないはずだが、念のため確認
			description:   "既に同じ種類のファイルが存在するのでエラー",
			gameVersionID: gameVersionID7,
			gameFile: domain.NewGameFile(
				fileID9,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID10),
					GameVersionID: uuid.UUID(gameVersionID7),
					FileTypeID:    fileTypeMap[gameFileTypeJar],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game2.jar",
				},
			},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID10),
					GameVersionID: uuid.UUID(gameVersionID7),
					FileTypeID:    fileTypeMap[gameFileTypeJar],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game2.jar",
				},
			},
			isErr: true,
		},
		{
			description:   "エラーの場合変更なし",
			gameVersionID: gameVersionID6,
			gameFile: domain.NewGameFile(
				fileID7,
				100,
				"/path/to/game.jar",
				[]byte("hash"),
			),
			beforeFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID8),
					GameVersionID: uuid.UUID(gameVersionID6),
					FileTypeID:    fileTypeMap[gameFileTypeWindows],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.exe",
				},
			},
			expectFiles: []GameFileTable{
				{
					ID:            uuid.UUID(fileID8),
					GameVersionID: uuid.UUID(gameVersionID6),
					FileTypeID:    fileTypeMap[gameFileTypeWindows],
					Hash:          []byte("hash"),
					EntryPoint:    "/path/to/game.exe",
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&GameTable{
				ID:          uuid.UUID(values.NewGameID()),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameVersions: []GameVersionTable{
					{
						ID:          uuid.UUID(testCase.gameVersionID),
						GameID:      uuid.UUID(values.NewGameID()),
						Name:        "test",
						Description: "test",
						CreatedAt:   time.Now(),
						GameFiles:   testCase.beforeFiles,
					},
				},
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameFileRepository.SaveGameFile(ctx, testCase.gameVersionID, testCase.gameFile)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var files []GameFileTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_version_id = ?", uuid.UUID(testCase.gameVersionID)).
				Find(&files).Error
			if err != nil {
				t.Fatalf("failed to get game files: %+v\n", err)
			}

			assert.Len(t, files, len(testCase.expectFiles))

			fileMap := make(map[uuid.UUID]GameFileTable)
			for _, file := range files {
				fileMap[file.ID] = file
			}

			for _, expectFile := range testCase.expectFiles {
				actualImage, ok := fileMap[expectFile.ID]
				if !ok {
					t.Errorf("not found image: %+v", expectFile)
				}

				assert.Equal(t, expectFile.GameVersionID, actualImage.GameVersionID)
				assert.Equal(t, expectFile.FileTypeID, actualImage.FileTypeID)
				assert.Equal(t, expectFile.EntryPoint, actualImage.EntryPoint)
				assert.Equal(t, expectFile.Hash, actualImage.Hash)
			}
		})
	}
}
