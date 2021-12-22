package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

type GameVersion struct {
	db *DB
}

func NewGameVersion(db *DB) *GameVersion {
	return &GameVersion{
		db: db,
	}
}

func (gv *GameVersion) CreateGameVersion(ctx context.Context, gameID values.GameID, version *domain.GameVersion) error {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gorm DB: %w", err)
	}

	gameVersion := GameVersionTable{
		ID:          uuid.UUID(version.GetID()),
		GameID:      uuid.UUID(gameID),
		Name:        string(version.GetName()),
		Description: string(version.GetDescription()),
		CreatedAt:   version.GetCreatedAt(),
	}

	err = gormDB.Create(&gameVersion).Error
	if err != nil {
		return fmt.Errorf("failed to create game version: %w", err)
	}

	return nil
}

func (gv *GameVersion) GetGameVersions(ctx context.Context, gameID values.GameID) ([]*domain.GameVersion, error) {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gorm DB: %w", err)
	}

	var gameVersions []GameVersionTable
	err = gormDB.
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at desc").
		Select("id", "name", "description", "created_at").
		Find(&gameVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	versions := make([]*domain.GameVersion, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		gameVersionID := values.NewGameVersionIDFromUUID(gameVersion.ID)
		gameVersionName := values.NewGameVersionName(gameVersion.Name)
		gameVersionDescription := values.NewGameVersionDescription(gameVersion.Description)
		versions = append(versions, domain.NewGameVersion(
			gameVersionID,
			gameVersionName,
			gameVersionDescription,
			gameVersion.CreatedAt,
		))
	}

	return versions, nil
}

func (gv *GameVersion) GetLatestGameVersion(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.GameVersion, error) {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gorm DB: %w", err)
	}

	gormDB, err = gv.db.setLock(gormDB, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersion GameVersionTable
	err = gormDB.
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at desc").
		Select("id", "name", "description", "created_at").
		First(&gameVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game version: %w", err)
	}

	return domain.NewGameVersion(
		values.NewGameVersionIDFromUUID(gameVersion.ID),
		values.NewGameVersionName(gameVersion.Name),
		values.NewGameVersionDescription(gameVersion.Description),
		gameVersion.CreatedAt,
	), nil
}

func (gv *GameVersion) GetLatestGameVersionsByGameIDs(ctx context.Context, gameIDs []values.GameID, lockType repository.LockType) (map[values.GameID]*domain.GameVersion, error) {
	db, err := gv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gorm DB: %w", err)
	}

	db, err = gv.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	uuidGameIDs := make([]uuid.UUID, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		uuidGameIDs = append(uuidGameIDs, uuid.UUID(gameID))
	}

	var gameVersionTables []GameVersionTable
	err = db.
		Where("game_versions.game_id in (?)", uuidGameIDs).
		Joins("JOIN (" +
			"SELECT game_id, MAX(created_at) AS created_at FROM game_versions GROUP BY game_id" +
			") as max_versions ON game_versions.game_id = max_versions.game_id AND game_versions.created_at = max_versions.created_at").
		Find(&gameVersionTables).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game versions: %w", err)
	}

	gameVersions := make(map[values.GameID]*domain.GameVersion, len(gameVersionTables))
	for _, gameVersion := range gameVersionTables {
		gameID := values.NewGameIDFromUUID(gameVersion.GameID)
		gameVersionID := values.NewGameVersionIDFromUUID(gameVersion.ID)
		gameVersionName := values.NewGameVersionName(gameVersion.Name)
		gameVersionDescription := values.NewGameVersionDescription(gameVersion.Description)

		gameVersions[gameID] = domain.NewGameVersion(
			gameVersionID,
			gameVersionName,
			gameVersionDescription,
			gameVersion.CreatedAt,
		)
	}

	return gameVersions, nil
}
