package gorm2

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type GameFileV2 struct {
	db *DB
}

func NewGameFileV2(db *DB) *GameFileV2 {
	return &GameFileV2{
		db: db,
	}
}

func (gameFile *GameFileV2) GetGameFilesWithoutTypes(ctx context.Context, fileIDs []values.GameFileID, lockType repository.LockType) ([]*repository.GameFileInfo, error) {
	if len(fileIDs) == 0 {
		return []*repository.GameFileInfo{}, nil
	}

	db, err := gameFile.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameFile.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	uuidFileIDs := make([]uuid.UUID, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		uuidFileIDs = append(uuidFileIDs, uuid.UUID(fileID))
	}

	var gameFiles []*migrate.GameFileTable2
	err = db.
		Joins("GameFileType").
		Where("v2_game_files.id IN ?", uuidFileIDs).
		Order("created_at DESC").
		Find(&gameFiles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find game files: %w", err)
	}

	gameFileInfos := make([]*repository.GameFileInfo, 0, len(gameFiles))
	for _, gameFile := range gameFiles {
		var fileType values.GameFileType
		switch gameFile.GameFileType.Name {
		case migrate.GameFileTypeWindows:
			fileType = values.GameFileTypeWindows
		case migrate.GameFileTypeMac:
			fileType = values.GameFileTypeMac
		case migrate.GameFileTypeJar:
			fileType = values.GameFileTypeJar
		default:
			// 1つ不正な値が格納されるだけで機能停止すると困るので、エラーを返さずにログを出力する
			log.Printf("error: unknown game file type: %s\n", gameFile.GameFileType.Name)
			continue
		}

		bytesHash, err := hex.DecodeString(gameFile.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash: %w", err)
		}

		gameFileInfos = append(gameFileInfos, &repository.GameFileInfo{
			GameFile: domain.NewGameFile(
				values.NewGameFileIDFromUUID(gameFile.ID),
				fileType,
				values.NewGameFileEntryPoint(gameFile.EntryPoint),
				values.NewGameFileHashFromBytes(bytesHash),
				gameFile.CreatedAt,
			),
			GameID: values.NewGameIDFromUUID(gameFile.GameID),
		})
	}

	return gameFileInfos, nil
}
