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

			gameFiles, err := gameFileRepository.GetGameFiles(ctx, testCase.fileIDs, testCase.lockType)

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
