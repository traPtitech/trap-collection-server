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
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestGetGameFilesWithoutTypesV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFileV2(testDB)

	type test struct {
		description string
		fileIDs     []values.GameFileID
		lockType    repository.LockType
		beforeGames []migrate.GameTable2
		gameFiles   []*repository.GameFileInfo
		isErr       bool
		err         error
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
	gameFileID6 := values.NewGameFileID()
	gameFileID7 := values.NewGameFileID()

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

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			fileIDs:     []values.GameFileID{gameFileID1},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(gameFileID1),
							GameID:     uuid.UUID(gameID1),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			gameFiles: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						[]byte("hash"),
						now,
					),
					GameID: gameID1,
				},
			},
		},
		{
			description: "windowsでもエラーなし",
			fileIDs:     []values.GameFileID{gameFileID2},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID2),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(gameFileID2),
							GameID:     uuid.UUID(gameID2),
							FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.exe",
							CreatedAt:  now,
						},
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			gameFiles: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						gameFileID2,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						[]byte("hash"),
						now,
					),
					GameID: gameID2,
				},
			},
		},
		{
			description: "macでもエラーなし",
			fileIDs:     []values.GameFileID{gameFileID3},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(gameFileID3),
							GameID:     uuid.UUID(gameID3),
							FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.app",
							CreatedAt:  now,
						},
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			gameFiles: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						gameFileID3,
						values.GameFileTypeMac,
						"/path/to/game.app",
						[]byte("hash"),
						now,
					),
					GameID: gameID3,
				},
			},
		},
		{
			description: "fileIDが空でもエラーなし",
			fileIDs:     []values.GameFileID{},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{},
			gameFiles:   []*repository.GameFileInfo{},
		},
		{
			description: "ファイルが複数でもエラーなし",
			fileIDs:     []values.GameFileID{gameFileID4, gameFileID5},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(gameFileID4),
							GameID:     uuid.UUID(gameID4),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
						{
							ID:         uuid.UUID(gameFileID5),
							GameID:     uuid.UUID(gameID4),
							FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.exe",
							CreatedAt:  now.Add(-time.Hour),
						},
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			gameFiles: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						gameFileID4,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						[]byte("hash"),
						now,
					),
					GameID: gameID4,
				},
				{
					GameFile: domain.NewGameFile(
						gameFileID5,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						[]byte("hash"),
						now.Add(-time.Hour),
					),
					GameID: gameID4,
				},
			},
		},
		{
			description: "対応するファイルが存在しない場合もエラーなし",
			fileIDs:     []values.GameFileID{gameFileID6},
			lockType:    repository.LockTypeNone,
			beforeGames: []migrate.GameTable2{},
			gameFiles:   []*repository.GameFileInfo{},
		},
		{
			description: "行ロックを取ってもエラーなし",
			fileIDs:     []values.GameFileID{gameFileID7},
			lockType:    repository.LockTypeRecord,
			beforeGames: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(gameFileID7),
							GameID:     uuid.UUID(gameID5),
							FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
							Hash:       "68617368",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
					VisibilityTypeID: gameVisibilityTypeIDPublic,
				},
			},
			gameFiles: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						gameFileID7,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						[]byte("hash"),
						now,
					),
					GameID: gameID5,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.beforeGames) != 0 {
				err := db.Create(&testCase.beforeGames).Error
				if err != nil {
					t.Fatalf("failed to create game: %v\n", err)
				}
			}

			gameFiles, err := gameFileRepository.GetGameFilesWithoutTypes(ctx, testCase.fileIDs, testCase.lockType)

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

			assert.Len(t, gameFiles, len(testCase.gameFiles))
			for i, expectGameFile := range testCase.gameFiles {
				actualGameFile := gameFiles[i]

				assert.Equal(t, expectGameFile.GetID(), actualGameFile.GetID())
				assert.Equal(t, expectGameFile.GetFileType(), actualGameFile.GetFileType())
				assert.Equal(t, expectGameFile.GetEntryPoint(), actualGameFile.GetEntryPoint())
				assert.Equal(t, expectGameFile.GetHash(), actualGameFile.GetHash())
				assert.WithinDuration(t, expectGameFile.GetCreatedAt(), actualGameFile.GetCreatedAt(), time.Second)
				assert.Equal(t, expectGameFile.GameID, actualGameFile.GameID)
			}
		})
	}
}

func TestSaveGameFileV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFileV2(testDB)

	type test struct {
		description string
		gameID      values.GameID
		file        *domain.GameFile
		beforeFiles []migrate.GameFileTable2
		expectFiles []migrate.GameFileTable2
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()
	fileID8 := values.NewGameFileID()

	var fileTypes []*migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&fileTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	fileTypeMap := make(map[string]int, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypeMap[fileType.Name] = fileType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			file: domain.NewGameFile(
				fileID1,
				values.GameFileTypeJar,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{},
			expectFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID1),
					GameID:     uuid.UUID(gameID1),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
		},
		{
			description: "windowsでも問題なし",
			gameID:      gameID2,
			file: domain.NewGameFile(
				fileID2,
				values.GameFileTypeWindows,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{},
			expectFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID2),
					GameID:     uuid.UUID(gameID2),
					FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
		},
		{
			description: "macでも問題なし",
			gameID:      gameID3,
			file: domain.NewGameFile(
				fileID3,
				values.GameFileTypeMac,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{},
			expectFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID3),
					GameID:     uuid.UUID(gameID3),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
		},
		{
			description: "想定外の画像の種類なのでエラー",
			gameID:      gameID4,
			file: domain.NewGameFile(
				fileID4,
				100,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{},
			expectFiles: []migrate.GameFileTable2{},
			isErr:       true,
		},
		{
			description: "既にファイルが存在しても問題なし",
			gameID:      gameID5,
			file: domain.NewGameFile(
				fileID5,
				values.GameFileTypeMac,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID6),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID6),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
				{
					ID:         uuid.UUID(fileID5),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
		},
		{
			description: "エラーの場合変更なし",
			gameID:      gameID6,
			file: domain.NewGameFile(
				fileID7,
				100,
				values.NewGameFileEntryPoint("path/to/file"),
				md5Hash,
				now,
			),
			beforeFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID8),
					GameID:     uuid.UUID(gameID6),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID8),
					GameID:     uuid.UUID(gameID6),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := db.Create(&migrate.GameTable2{
				ID:               uuid.UUID(testCase.gameID),
				Name:             "test",
				Description:      "test",
				CreatedAt:        time.Now(),
				GameFiles:        testCase.beforeFiles,
				VisibilityTypeID: gameVisibilityTypeIDPublic,
			}).Error
			if err != nil {
				t.Fatalf("failed to create game table: %+v\n", err)
			}

			err = gameFileRepository.SaveGameFile(ctx, testCase.gameID, testCase.file)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			var files []migrate.GameFileTable2
			err = db.
				Session(&gorm.Session{}).
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Find(&files).Error
			if err != nil {
				t.Fatalf("failed to get role table: %+v\n", err)
			}

			assert.Len(t, files, len(testCase.expectFiles))

			fileMap := make(map[uuid.UUID]migrate.GameFileTable2)
			for _, file := range files {
				fileMap[file.ID] = file
			}

			for _, expectFile := range testCase.expectFiles {
				actualFile, ok := fileMap[expectFile.ID]
				if !ok {
					t.Errorf("not found file: %+v", expectFile)
				}

				assert.Equal(t, expectFile.GameID, actualFile.GameID)
				assert.Equal(t, expectFile.FileTypeID, actualFile.FileTypeID)
				assert.Equal(t, expectFile.EntryPoint, actualFile.EntryPoint)
				assert.Equal(t, expectFile.Hash, actualFile.Hash)
				assert.WithinDuration(t, expectFile.CreatedAt, actualFile.CreatedAt, 2*time.Second)
			}
		})
	}
}

func TestGetGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFileV2(testDB)

	type test struct {
		description string
		fileID      values.GameFileID
		lockType    repository.LockType
		files       []migrate.GameFileTable2
		expectFile  repository.GameFileInfo
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()

	var fileTypes []*migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&fileTypes).Error
	if err != nil {
		t.Fatalf("failed to get file type table: %+v\n", err)
	}

	fileTypeMap := make(map[string]int, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypeMap[fileType.Name] = fileType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			fileID:      fileID1,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID1),
					GameID:     uuid.UUID(gameID1),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFile: repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					fileID1,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				GameID: gameID1,
			},
		},
		{
			description: "windowsでも問題なし",
			fileID:      fileID2,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID2),
					GameID:     uuid.UUID(gameID2),
					FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFile: repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					fileID2,
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				GameID: gameID2,
			},
		},
		{
			description: "macでも問題なし",
			fileID:      fileID3,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID3),
					GameID:     uuid.UUID(gameID3),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFile: repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					fileID3,
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				GameID: gameID3,
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			fileID:      fileID4,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID4),
					GameID:     uuid.UUID(gameID4),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFile: repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					fileID4,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				GameID: gameID4,
			},
		},
		{
			description: "複数の画像があっても問題なし",
			fileID:      fileID5,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID5),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
				{
					ID:         uuid.UUID(fileID6),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFile: repository.GameFileInfo{
				GameFile: domain.NewGameFile(
					fileID5,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				GameID: gameID5,
			},
		},
		{
			description: "ファイルが存在しないのでRecordNotFound",
			fileID:      fileID7,
			lockType:    repository.LockTypeNone,
			files:       []migrate.GameFileTable2{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, file := range testCase.files {
				if game, ok := gameIDMap[file.GameID]; ok {
					game.GameFiles = append(game.GameFiles, file)
				} else {
					gameIDMap[file.GameID] = &migrate.GameTable2{
						ID:               file.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameFiles:        []migrate.GameFileTable2{file},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			file, err := gameFileRepository.GetGameFile(ctx, testCase.fileID, testCase.lockType)

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

			assert.Equal(t, testCase.expectFile.GetID(), file.GetID())
			assert.Equal(t, testCase.expectFile.GetFileType(), file.GetFileType())
			assert.Equal(t, testCase.expectFile.GetEntryPoint(), file.GetEntryPoint())
			assert.Equal(t, testCase.expectFile.GetHash(), file.GetHash())
			assert.WithinDuration(t, testCase.expectFile.GetCreatedAt(), file.GetCreatedAt(), time.Second)
			assert.Equal(t, testCase.expectFile.GameID, file.GameID)
		})
	}
}

func TestGetGameFilesV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %+v\n", err)
	}

	gameFileRepository := NewGameFileV2(testDB)

	type test struct {
		description string
		gameID      values.GameID
		lockType    repository.LockType
		files       []migrate.GameFileTable2
		expectFiles []*domain.GameFile
		isErr       bool
		err         error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()

	var fileTypes []*migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Find(&fileTypes).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	fileTypeMap := make(map[string]int, len(fileTypes))
	for _, fileType := range fileTypes {
		fileTypeMap[fileType.Name] = fileType.ID
	}

	var gameVisibilityPublic migrate.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&migrate.GameVisibilityTypeTable{Name: migrate.GameVisibilityTypePublic}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("failed to get game visibility: %v\n", err)
	}
	gameVisibilityTypeIDPublic := gameVisibilityPublic.ID

	md5Hash := values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6})

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			gameID:      gameID1,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID1),
					GameID:     uuid.UUID(gameID1),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []*domain.GameFile{
				domain.NewGameFile(
					fileID1,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
		},
		{
			description: "windowsでも問題なし",
			gameID:      gameID2,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID2),
					GameID:     uuid.UUID(gameID2),
					FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []*domain.GameFile{
				domain.NewGameFile(
					fileID2,
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
		},
		{
			description: "macでも問題なし",
			gameID:      gameID3,
			lockType:    repository.LockTypeNone,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID3),
					GameID:     uuid.UUID(gameID3),
					FileTypeID: fileTypeMap[migrate.GameFileTypeMac],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []*domain.GameFile{
				domain.NewGameFile(
					fileID3,
					values.GameFileTypeMac,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
		},
		{
			description: "lockTypeがRecordでも問題なし",
			gameID:      gameID4,
			lockType:    repository.LockTypeRecord,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID4),
					GameID:     uuid.UUID(gameID4),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
			},
			expectFiles: []*domain.GameFile{
				domain.NewGameFile(
					fileID4,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
			},
		},
		{
			description: "複数のファイルがあっても問題なし",
			gameID:      gameID5,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID5),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeJar],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now,
				},
				{
					ID:         uuid.UUID(fileID6),
					GameID:     uuid.UUID(gameID5),
					FileTypeID: fileTypeMap[migrate.GameFileTypeWindows],
					EntryPoint: "path/to/file",
					Hash:       md5Hash.String(),
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			expectFiles: []*domain.GameFile{
				domain.NewGameFile(
					fileID5,
					values.GameFileTypeJar,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now,
				),
				domain.NewGameFile(
					fileID6,
					values.GameFileTypeWindows,
					values.NewGameFileEntryPoint("path/to/file"),
					md5Hash,
					now.Add(-time.Hour),
				),
			},
		},
		{
			description: "ファイルが存在しなくても問題なし",
			gameID:      gameID6,
			lockType:    repository.LockTypeNone,
			files:       []migrate.GameFileTable2{},
			expectFiles: []*domain.GameFile{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameIDMap := map[uuid.UUID]*migrate.GameTable2{}
			for _, file := range testCase.files {
				if game, ok := gameIDMap[file.GameID]; ok {
					game.GameFiles = append(game.GameFiles, file)
				} else {
					gameIDMap[file.GameID] = &migrate.GameTable2{
						ID:               file.GameID,
						Name:             "test",
						Description:      "test",
						CreatedAt:        now,
						GameFiles:        []migrate.GameFileTable2{file},
						VisibilityTypeID: gameVisibilityTypeIDPublic,
					}
				}
			}

			games := make([]migrate.GameTable2, 0, len(gameIDMap))
			for _, game := range gameIDMap {
				games = append(games, *game)
			}

			if len(games) > 0 {
				err := db.Create(games).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}
			}

			files, err := gameFileRepository.GetGameFiles(ctx, testCase.gameID, testCase.lockType)

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

			for i, expectFile := range testCase.expectFiles {
				assert.Equal(t, expectFile.GetID(), files[i].GetID())
				assert.Equal(t, expectFile.GetFileType(), files[i].GetFileType())
				assert.Equal(t, expectFile.GetEntryPoint(), files[i].GetEntryPoint())
				assert.Equal(t, expectFile.GetHash(), files[i].GetHash())
				assert.WithinDuration(t, expectFile.GetCreatedAt(), files[i].GetCreatedAt(), time.Second)
			}
		})
	}
}
