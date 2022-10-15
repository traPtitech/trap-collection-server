package gorm2

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
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
		Create(&migrate.GameVersionTable2{
			ID:          uuid.UUID(version.GetID()),
			Name:        string(version.GetName()),
			Description: string(version.GetDescription()),
			URL:         dbURL,
			CreatedAt:   version.GetCreatedAt(),
			GameID:      uuid.UUID(gameID),
			GameImageID: uuid.UUID(imageID),
			GameVideoID: uuid.UUID(videoID),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to create game version: %w", err)
	}

	files := make([]*migrate.GameFileTable2, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		files = append(files, &migrate.GameFileTable2{
			ID: uuid.UUID(fileID),
		})
	}

	err = db.
		Model(&migrate.GameVersionTable2{
			ID: uuid.UUID(version.GetID()),
		}).
		Association("GameFiles").
		Append(files)
	if err != nil {
		return fmt.Errorf("failed to append game files: %w", err)
	}

	return nil
}

func (gameVersion *GameVersionV2) GetGameVersions(
	ctx context.Context,
	gameID values.GameID,
	limit uint,
	offset uint,
	lockType repository.LockType,
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
		Model(&migrate.GameVersionTable2{}).
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

	var gameVersions []*migrate.GameVersionTable2
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
