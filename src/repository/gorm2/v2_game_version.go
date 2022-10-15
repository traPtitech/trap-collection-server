package gorm2

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
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
