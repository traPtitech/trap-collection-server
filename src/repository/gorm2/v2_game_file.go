package gorm2

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

type GameFileV2 struct {
	db *DB
}

func NewGameFileV2(db *DB) *GameFileV2 {
	return &GameFileV2{
		db: db,
	}
}

func (gameFile *GameFileV2) SaveGameFile(ctx context.Context, gameID values.GameID, file *domain.GameFile) error {
	db, err := gameFile.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var fileTypeName string
	switch file.GetFileType() {
	case values.GameFileTypeJar:
		fileTypeName = migrate.GameFileTypeJar
	case values.GameFileTypeWindows:
		fileTypeName = migrate.GameFileTypeWindows
	case values.GameFileTypeMac:
		fileTypeName = migrate.GameFileTypeMac
	default:
		return fmt.Errorf("invalid file type: %d", file.GetFileType())
	}

	var fileType migrate.GameFileTypeTable
	err = db.
		Where("name = ?", fileTypeName).
		Select("id").
		Take(&fileType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	fileTypeID := fileType.ID

	err = db.
		Create(&migrate.GameFileTable2{
			ID:         uuid.UUID(file.GetID()),
			GameID:     uuid.UUID(gameID),
			EntryPoint: string(file.GetEntryPoint()),
			Hash:       file.GetHash().String(),
			FileTypeID: fileTypeID,
			CreatedAt:  file.GetCreatedAt(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to create game file: %w", err)
	}

	return nil
}

func (gameFile *GameFileV2) GetGameFile(ctx context.Context, gameFileID values.GameFileID, lockType repository.LockType, fileTypes []values.GameFileType) (*repository.GameFileInfo, error) {
	db, err := gameFile.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameFile.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var file migrate.GameFileTable2
	err = db.
		Joins("GameFileType").
		Where("v2_game_files.id = ?", uuid.UUID(gameFileID)).
		Take(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game file: %w", err)
	}

	var fileType values.GameFileType
	switch file.GameFileType.Name {
	case migrate.GameFileTypeJar:
		fileType = values.GameFileTypeJar
	case migrate.GameFileTypeWindows:
		fileType = values.GameFileTypeWindows
	case migrate.GameFileTypeMac:
		fileType = values.GameFileTypeMac
	default:
		return nil, fmt.Errorf("invalid file type: %s", file.GameFileType.Name)
	}

	return &repository.GameFileInfo{
		GameFile: domain.NewGameFile(
			values.NewGameFileIDFromUUID(file.ID),
			fileType,
			values.GameFileEntryPoint(file.EntryPoint),
			values.GameFileHash(file.Hash),
			file.CreatedAt,
		),
		GameID: values.NewGameIDFromUUID(file.GameID),
	}, nil
}

func (gameFile *GameFileV2) GetGameFiles(ctx context.Context, gameID values.GameID, lockType repository.LockType, fileTypes []values.GameFileType) ([]*domain.GameFile, error) {
	db, err := gameFile.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameFile.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var files []migrate.GameFileTable2
	err = db.
		Joins("GameFileType").
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game files: %w", err)
	}

	fileTypesMap := make(map[values.GameFileType]struct{})
	for _, fileType := range fileTypes {
		fileTypesMap[fileType] = struct{}{}
	}

	gameFiles := make([]*domain.GameFile, 0, len(files))
	for _, file := range files {
		var fileType values.GameFileType
		switch file.GameFileType.Name {
		case migrate.GameFileTypeJar:
			fileType = values.GameFileTypeJar
		case migrate.GameFileTypeWindows:
			fileType = values.GameFileTypeWindows
		case migrate.GameFileTypeMac:
			fileType = values.GameFileTypeMac
		default:
			// 1つ不正な値が格納されるだけで機能停止すると困るので、エラーを返さずにログを出力する
			log.Printf("error: unknown game file type: %s\n", file.GameFileType.Name)
			continue
		}
		if _, ok := fileTypesMap[fileType]; !ok {
			continue
		}

		gameFiles = append(gameFiles, domain.NewGameFile(
			values.NewGameFileIDFromUUID(file.ID),
			fileType,
			values.GameFileEntryPoint(file.EntryPoint),
			values.GameFileHash(file.Hash),
			file.CreatedAt,
		))
	}

	return gameFiles, nil
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
