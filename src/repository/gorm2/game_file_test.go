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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestSaveGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFile(testDB)

	type test struct {
		description   string
		gameVersionID values.GameVersionID
		gameFile      *domain.GameFile
		beforeFiles   []migrate.GameFileTable
		expectFiles   []migrate.GameFileTable
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

	var fileTypes []*migrate.GameFileTypeTable
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

	now := time.Now()

	testCases := []test{
		{
			description:   "特に問題ないので問題なし",
			gameVersionID: gameVersionID1,
			gameFile: domain.NewGameFile(
				fileID1,
				values.GameFileTypeJar,
				"/path/to/game.jar",
				[]byte("hash"),
				now,
			),
			beforeFiles: []migrate.GameFileTable{},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID1),
					GameVersionID: uuid.UUID(gameVersionID1),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeJar],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.jar",
					CreatedAt:     now,
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID2),
					GameVersionID: uuid.UUID(gameVersionID2),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeWindows],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.exe",
					CreatedAt:     now,
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID3),
					GameVersionID: uuid.UUID(gameVersionID3),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeMac],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.app",
					CreatedAt:     now,
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{},
			expectFiles: []migrate.GameFileTable{},
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID6),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeWindows],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.exe",
					CreatedAt:     now,
				},
			},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID6),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeWindows],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.exe",
					CreatedAt:     now,
				},
				{
					ID:            uuid.UUID(fileID5),
					GameVersionID: uuid.UUID(gameVersionID5),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeJar],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.jar",
					CreatedAt:     now,
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID10),
					GameVersionID: uuid.UUID(gameVersionID7),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeJar],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game2.jar",
					CreatedAt:     now,
				},
			},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID10),
					GameVersionID: uuid.UUID(gameVersionID7),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeJar],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game2.jar",
					CreatedAt:     now,
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
				now,
			),
			beforeFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID8),
					GameVersionID: uuid.UUID(gameVersionID6),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeWindows],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.exe",
					CreatedAt:     now,
				},
			},
			expectFiles: []migrate.GameFileTable{
				{
					ID:            uuid.UUID(fileID8),
					GameVersionID: uuid.UUID(gameVersionID6),
					FileTypeID:    fileTypeMap[migrate.GameFileTypeWindows],
					Hash:          "68617368",
					EntryPoint:    "/path/to/game.exe",
					CreatedAt:     now,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable{
				ID:          uuid.UUID(values.NewGameID()),
				Name:        "test",
				Description: "test",
				CreatedAt:   time.Now(),
				GameVersions: []migrate.GameVersionTable{
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

			var files []migrate.GameFileTable
			err = db.
				Session(&gorm.Session{}).
				Where("game_version_id = ?", uuid.UUID(testCase.gameVersionID)).
				Find(&files).Error
			if err != nil {
				t.Fatalf("failed to get game files: %+v\n", err)
			}

			assert.Len(t, files, len(testCase.expectFiles))

			fileMap := make(map[uuid.UUID]migrate.GameFileTable)
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
				assert.WithinDuration(t, expectFile.CreatedAt, actualImage.CreatedAt, time.Second)
			}
		})
	}
}

func TestGetGameFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFile(testDB)

	type test struct {
		description        string
		gameVersionID      values.GameVersionID
		fileTypes          []values.GameFileType
		beforeGameVersions []migrate.GameVersionTable
		gameFiles          []*domain.GameFile
		isErr              bool
		err                error
	}

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()
	gameVersionID10 := values.NewGameVersionID()
	gameVersionID11 := values.NewGameVersionID()
	gameVersionID12 := values.NewGameVersionID()

	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameFileID3 := values.NewGameFileID()
	gameFileID4 := values.NewGameFileID()
	gameFileID5 := values.NewGameFileID()
	gameFileID6 := values.NewGameFileID()
	gameFileID7 := values.NewGameFileID()
	gameFileID8 := values.NewGameFileID()
	gameFileID9 := values.NewGameFileID()
	gameFileID10 := values.NewGameFileID()
	gameFileID11 := values.NewGameFileID()
	gameFileID12 := values.NewGameFileID()

	now := time.Now()

	var fileTypes []*migrate.GameFileTypeTable
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
			description:   "特に問題ないのでエラーなし",
			gameVersionID: gameVersionID1,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID1),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID1,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "fileTypeがwindowsでもエラーなし",
			gameVersionID: gameVersionID2,
			fileTypes: []values.GameFileType{
				values.GameFileTypeWindows,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID2),
							FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.exe",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID2,
					values.GameFileTypeWindows,
					"/path/to/game.exe",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "fileTypeがmacでもエラーなし",
			gameVersionID: gameVersionID3,
			fileTypes: []values.GameFileType{
				values.GameFileTypeMac,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID3),
							FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.app",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID3,
					values.GameFileTypeMac,
					"/path/to/game.app",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "指定のfileType以外のものは含まない",
			gameVersionID: gameVersionID4,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID4),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
						{
							ID:         uuid.UUID(gameFileID5),
							FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.exe",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID4,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "指定のfileTypeが複数でもエラーなし",
			gameVersionID: gameVersionID5,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
				values.GameFileTypeWindows,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID6),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
						{
							ID:         uuid.UUID(gameFileID7),
							FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.exe",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID6,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					[]byte("hash"),
					now,
				),
				domain.NewGameFile(
					gameFileID7,
					values.GameFileTypeWindows,
					"/path/to/game.exe",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "ファイルがなくてもエラーなし",
			gameVersionID: gameVersionID6,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles:   []migrate.GameFileTable{},
				},
			},
			gameFiles: []*domain.GameFile{},
		},
		{
			// 実際には発生しないが、念の為確認
			description:   "ゲームバージョンがなくてもエラーなし",
			gameVersionID: gameVersionID7,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
			},
			beforeGameVersions: []migrate.GameVersionTable{},
			gameFiles:          []*domain.GameFile{},
		},
		{
			// 実際には発生しないが、念の為確認
			description:   "fileTypesがなくてもエラーなし",
			gameVersionID: gameVersionID8,
			fileTypes:     []values.GameFileType{},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID8),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{},
		},
		{
			description:   "別のゲームバージョンにファイルがあっても含まない",
			gameVersionID: gameVersionID9,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID9),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID9),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
				{
					ID:          uuid.UUID(gameVersionID10),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID10),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID9,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					[]byte("hash"),
					now,
				),
			},
		},
		{
			description:   "fileTypesが不正なのでエラー",
			gameVersionID: gameVersionID11,
			fileTypes: []values.GameFileType{
				100,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID11),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID11),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
			isErr: true,
		},
		{
			description:   "fileTypeが存在しなければ含まない",
			gameVersionID: gameVersionID12,
			fileTypes: []values.GameFileType{
				values.GameFileTypeJar,
				values.GameFileTypeWindows,
			},
			beforeGameVersions: []migrate.GameVersionTable{
				{
					ID:          uuid.UUID(gameVersionID12),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameFiles: []migrate.GameFileTable{
						{
							ID:         uuid.UUID(gameFileID12),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
			gameFiles: []*domain.GameFile{
				domain.NewGameFile(
					gameFileID12,
					values.GameFileTypeJar,
					"/path/to/game.jar",
					[]byte("hash"),
					now,
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGameVersions) != 0 {
				err := db.Create(&migrate.GameTable{
					ID:           uuid.UUID(values.NewGameID()),
					Name:         "test",
					Description:  "test",
					CreatedAt:    time.Now(),
					GameVersions: testCase.beforeGameVersions,
				}).Error
				if err != nil {
					t.Fatalf("failed to create game: %v\n", err)
				}
			}

			gameFiles, err := gameFileRepository.GetGameFiles(ctx, testCase.gameVersionID, testCase.fileTypes)

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

			assert.Len(t, gameFiles, len(testCase.gameFiles))

			fileMap := make(map[values.GameFileID]*domain.GameFile, len(gameFiles))
			for _, gameFile := range gameFiles {
				fileMap[gameFile.GetID()] = gameFile
			}

			for _, expectGameFile := range testCase.gameFiles {
				actualGameFile, ok := fileMap[expectGameFile.GetID()]
				if !ok {
					t.Errorf("game file not found: %v", expectGameFile.GetID())
					continue
				}

				assert.Equal(t, expectGameFile.GetID(), actualGameFile.GetID())
				assert.Equal(t, expectGameFile.GetFileType(), actualGameFile.GetFileType())
				assert.Equal(t, expectGameFile.GetEntryPoint(), actualGameFile.GetEntryPoint())
				assert.Equal(t, expectGameFile.GetHash(), actualGameFile.GetHash())
				assert.WithinDuration(t, expectGameFile.GetCreatedAt(), actualGameFile.GetCreatedAt(), time.Second)
			}
		})
	}
}
