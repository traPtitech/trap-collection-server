package gorm2

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/file"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type GameFile struct {
	db *DB
}

func NewGameFile(db *DB) *GameFile {
	return &GameFile{
		db: db,
	}
}

func (gf *GameFile) SaveGameFile(ctx context.Context, gameVersionID values.GameVersionID, gameFile *domain.GameFile) error {
	gormDB, err := gf.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var fileTypeName string
	switch gameFile.GetFileType() {
	case values.GameFileTypeJar:
		fileTypeName = migrate.GameFileTypeJar
	case values.GameFileTypeWindows:
		fileTypeName = migrate.GameFileTypeWindows
	case values.GameFileTypeMac:
		fileTypeName = migrate.GameFileTypeMac
	default:
		return fmt.Errorf("invalid file type: %d", gameFile.GetFileType())
	}

	var fileType migrate.GameFileTypeTable
	err = gormDB.
		Where("name = ?", fileTypeName).
		Where("active").
		Select("id").
		Take(&fileType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	fileTypeID := fileType.ID

	err = gormDB.Create(&migrate.GameFileTable{
		ID:            uuid.UUID(gameFile.GetID()),
		GameVersionID: uuid.UUID(gameVersionID),
		FileTypeID:    fileTypeID,
		Hash:          hex.EncodeToString(gameFile.GetHash()),
		EntryPoint:    string(gameFile.GetEntryPoint()),
		CreatedAt:     gameFile.GetCreatedAt(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create game image: %w", err)
	}

	return nil
}

func (gf *GameFile) GetGameFiles(ctx context.Context, gameVersionID values.GameVersionID, fileTypes []values.GameFileType) ([]*domain.GameFile, error) {
	if len(fileTypes) == 0 {
		return []*domain.GameFile{}, nil
	}

	gormDB, err := gf.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	fileTypeNames := make([]string, 0, len(fileTypes))
	for _, fileType := range fileTypes {
		switch fileType {
		case values.GameFileTypeJar:
			fileTypeNames = append(fileTypeNames, migrate.GameFileTypeJar)
		case values.GameFileTypeWindows:
			fileTypeNames = append(fileTypeNames, migrate.GameFileTypeWindows)
		case values.GameFileTypeMac:
			fileTypeNames = append(fileTypeNames, migrate.GameFileTypeMac)
		default:
			return nil, fmt.Errorf("invalid file type: %d", fileType)
		}
	}

	var dbGameFiles []migrate.GameFileTable
	err = gormDB.
		Joins("GameFileType").
		Where("game_version_id = ?", uuid.UUID(gameVersionID)).
		Where("GameFileType.Name IN (?)", fileTypeNames).
		Where("GameFileType.Active"). // 無効化された種類のファイルは取得しない
		Select("game_files.id", "GameFileType.Name", "hash", "entry_point", "created_at").
		Find(&dbGameFiles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game files: %w", err)
	}

	gameFiles := make([]*domain.GameFile, 0, len(dbGameFiles))
	for _, gameFile := range dbGameFiles {
		var fileType values.GameFileType
		switch gameFile.GameFileType.Name {
		case migrate.GameFileTypeJar:
			fileType = values.GameFileTypeJar
		case migrate.GameFileTypeWindows:
			fileType = values.GameFileTypeWindows
		case migrate.GameFileTypeMac:
			fileType = values.GameFileTypeMac
		default:
			return nil, fmt.Errorf("invalid file type: %s", gameFile.GameFileType.Name)
		}

		bytesHash, err := hex.DecodeString(gameFile.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash: %w", err)
		}

		gameFiles = append(gameFiles, domain.NewGameFile(
			values.NewGameFileIDFromUUID(gameFile.ID),
			fileType,
			values.NewGameFileEntryPoint(gameFile.EntryPoint),
			values.NewGameFileHashFromBytes(bytesHash),
			gameFile.CreatedAt,
		))
	}

	return gameFiles, nil
}
