package gorm2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameVersionV2 struct {
	db *DB
}

func NewGameVersionV2(db *DB) *GameVersionV2 {
	return &GameVersionV2{
		db: db,
	}
}

func (gameVersion *GameVersionV2) CreateGameVersion(
	ctx context.Context,
	gameID values.GameID,
	imageID values.GameImageID,
	videoID values.GameVideoID,
	optionURL types.Option[values.GameURLLink],
	fileIDs []values.GameFileID,
	version *domain.GameVersion,
) error {
	db, err := gameVersion.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var dbURL string
	if urlValue, ok := optionURL.Value(); ok {
		dbURL = (*url.URL)(urlValue).String()
	}

	err = db.
		Session(&gorm.Session{}).
		Create(&schema.GameVersionTable2{
			ID:          uuid.UUID(version.GetID()),
			Name:        string(version.GetName()),
			Description: string(version.GetDescription()),
			URL:         dbURL,
			CreatedAt:   version.GetCreatedAt(),
			GameID:      uuid.UUID(gameID),
			GameImageID: uuid.UUID(imageID),
			GameVideoID: uuid.UUID(videoID),
		}).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return repository.ErrDuplicatedUniqueKey
		}
	}
	if err != nil {
		return fmt.Errorf("failed to create game version: %w", err)
	}

	files := make([]*schema.GameFileTable2, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		files = append(files, &schema.GameFileTable2{
			ID: uuid.UUID(fileID),
		})
	}

	err = db.
		Model(&schema.GameVersionTable2{
			ID: uuid.UUID(version.GetID()),
		}).
		Association("GameFiles").
		Append(files)
	if err != nil {
		return fmt.Errorf("failed to append game files: %w", err)
	}

	err = db.
		Model(&schema.GameTable2{ID: uuid.UUID(gameID)}).
		Update("latest_version_updated_at", version.GetCreatedAt()).
		Error
	if err != nil {
		return fmt.Errorf("failed to update latest version updated at: %w", err)
	}

	return nil
}

func (gameVersion *GameVersionV2) GetGameVersions(
	ctx context.Context,
	gameID values.GameID,
	limit uint,
	offset uint,
	_ repository.LockType,
) (uint, []*repository.GameVersionInfo, error) {
	db, err := gameVersion.db.getDB(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get db: %w", err)
	}

	query := db.
		Where("game_id = ?", uuid.UUID(gameID))

	var count int64
	err = query.
		Session(&gorm.Session{}).
		Model(&schema.GameVersionTable2{}).
		Count(&count).Error
	if err != nil {
		return 0, nil, fmt.Errorf("failed to count game versions: %w", err)
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	query = query.Order("created_at DESC")

	var gameVersions []*schema.GameVersionTable2
	err = query.
		Session(&gorm.Session{}).
		Preload("GameFiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
		Find(&gameVersions).Error
	if err != nil {
		return 0, nil, fmt.Errorf("failed to find game versions: %w", err)
	}

	gameVersionInfos := make([]*repository.GameVersionInfo, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		var optionURL types.Option[values.GameURLLink]
		if len(gameVersion.URL) != 0 {
			url, err := url.Parse(gameVersion.URL)
			if err != nil {
				// 1つのurlが不正なだけでエラーになると困るのでログを出して続行
				log.Printf("failed to parse game version url: %v\n", err)
				continue
			}

			optionURL = types.NewOption(values.NewGameURLLink(url))
		}

		fileIDs := make([]values.GameFileID, 0, len(gameVersion.GameFiles))
		for _, file := range gameVersion.GameFiles {
			fileIDs = append(fileIDs, values.NewGameFileIDFromUUID(file.ID))
		}

		gameVersionInfos = append(gameVersionInfos, &repository.GameVersionInfo{
			GameVersion: domain.NewGameVersion(
				values.NewGameVersionIDFromUUID(gameVersion.ID),
				values.NewGameVersionName(gameVersion.Name),
				values.NewGameVersionDescription(gameVersion.Description),
				gameVersion.CreatedAt,
			),
			ImageID: values.GameImageIDFromUUID(gameVersion.GameImageID),
			VideoID: values.NewGameVideoIDFromUUID(gameVersion.GameVideoID),
			URL:     optionURL,
			FileIDs: fileIDs,
		})
	}

	return uint(count), gameVersionInfos, nil
}

func (gameVersion *GameVersionV2) GetLatestGameVersion(
	ctx context.Context,
	gameID values.GameID,
	_ repository.LockType,
) (*repository.GameVersionInfo, error) {
	db, err := gameVersion.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var gameVersionTable schema.GameVersionTable2
	err = db.
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		Preload("GameFiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
		First(&gameVersionTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find game version: %w", err)
	}

	var optionURL types.Option[values.GameURLLink]
	if len(gameVersionTable.URL) != 0 {
		url, err := url.Parse(gameVersionTable.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse game version url: %w", err)
		}

		optionURL = types.NewOption(values.NewGameURLLink(url))
	}

	fileIDs := make([]values.GameFileID, 0, len(gameVersionTable.GameFiles))
	for _, file := range gameVersionTable.GameFiles {
		fileIDs = append(fileIDs, values.NewGameFileIDFromUUID(file.ID))
	}

	return &repository.GameVersionInfo{
		GameVersion: domain.NewGameVersion(
			values.NewGameVersionIDFromUUID(gameVersionTable.ID),
			values.NewGameVersionName(gameVersionTable.Name),
			values.NewGameVersionDescription(gameVersionTable.Description),
			gameVersionTable.CreatedAt,
		),
		ImageID: values.GameImageIDFromUUID(gameVersionTable.GameImageID),
		VideoID: values.NewGameVideoIDFromUUID(gameVersionTable.GameVideoID),
		URL:     optionURL,
		FileIDs: fileIDs,
	}, nil
}

func (gameVersion *GameVersionV2) GetGameVersionsByIDs(
	ctx context.Context,
	gameVersionIDs []values.GameVersionID,
	lockType repository.LockType,
) ([]*repository.GameVersionInfoWithGameID, error) {
	db, err := gameVersion.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameVersion.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	uuidGameVersionIDs := make([]uuid.UUID, 0, len(gameVersionIDs))
	for _, gameVersionID := range gameVersionIDs {
		uuidGameVersionIDs = append(uuidGameVersionIDs, uuid.UUID(gameVersionID))
	}

	var gameVersionTables []*schema.GameVersionTable2
	err = db.
		Where("id IN ?", uuidGameVersionIDs).
		Preload("GameFiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
		Find(&gameVersionTables).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find game versions: %w", err)
	}

	gameVersionInfos := make([]*repository.GameVersionInfoWithGameID, 0, len(gameVersionTables))
	for _, gameVersionTable := range gameVersionTables {
		var optionURL types.Option[values.GameURLLink]
		if len(gameVersionTable.URL) != 0 {
			url, err := url.Parse(gameVersionTable.URL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse game version url: %w", err)
			}

			optionURL = types.NewOption(values.NewGameURLLink(url))
		}

		fileIDs := make([]values.GameFileID, 0, len(gameVersionTable.GameFiles))
		for _, file := range gameVersionTable.GameFiles {
			fileIDs = append(fileIDs, values.NewGameFileIDFromUUID(file.ID))
		}

		gameVersionInfos = append(gameVersionInfos, &repository.GameVersionInfoWithGameID{
			GameVersion: domain.NewGameVersion(
				values.NewGameVersionIDFromUUID(gameVersionTable.ID),
				values.NewGameVersionName(gameVersionTable.Name),
				values.NewGameVersionDescription(gameVersionTable.Description),
				gameVersionTable.CreatedAt,
			),
			GameID:  values.NewGameIDFromUUID(gameVersionTable.GameID),
			ImageID: values.GameImageIDFromUUID(gameVersionTable.GameImageID),
			VideoID: values.NewGameVideoIDFromUUID(gameVersionTable.GameVideoID),
			URL:     optionURL,
			FileIDs: fileIDs,
		})
	}

	return gameVersionInfos, nil
}

func (gameVersion *GameVersionV2) GetGameVersionByID(
	ctx context.Context,
	gameVersionID values.GameVersionID,
	lockType repository.LockType,
) (*repository.GameVersionInfoWithGameID, error) {
	// TODO: 正しい実装を行う
	// 仮実装: GetGameVersionsByIDsを使って1件取得
	gameVersionInfos, err := gameVersion.GetGameVersionsByIDs(ctx, []values.GameVersionID{gameVersionID}, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to get game version by id: %w", err)
	}
	if len(gameVersionInfos) == 0 {
		return nil, repository.ErrRecordNotFound
	}
	return gameVersionInfos[0], nil
}
