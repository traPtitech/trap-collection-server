package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

func TestEditionResourceMembership(t *testing.T) {
	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	repo := NewEdition(&DB{db: tx})
	now := time.Now()
	gameID, imageID, videoID, fileID, versionID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	editionA, editionB := uuid.New(), uuid.New()
	require.NoError(t, tx.Create(&schema.GameTable2{ID: gameID, Name: "membership-game", Description: "test", VisibilityTypeID: 1, CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameImageTable2{ID: imageID, GameID: gameID, ImageTypeID: 1, CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameVideoTable2{ID: videoID, GameID: gameID, VideoTypeID: 1, CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameFileTable2{ID: fileID, GameID: gameID, FileTypeID: 1, Hash: "12345678901234567890123456789012", EntryPoint: "game", CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameVersionTable2{ID: versionID, GameID: gameID, GameImageID: imageID, GameVideoID: videoID, Name: "v1", Description: "test", CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.EditionTable{ID: editionA, Name: "membership-a", CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.EditionTable{ID: editionB, Name: "membership-b", CreatedAt: now}).Error)
	require.NoError(t, tx.Exec("INSERT INTO edition_game_version_relations (edition_id, game_version_id) VALUES (?, ?)", editionA, versionID).Error)
	require.NoError(t, tx.Exec("INSERT INTO game_version_game_file_relations (game_version_id, game_file_id) VALUES (?, ?)", versionID, fileID).Error)
	tests := []struct {
		name   string
		lookup func(values.EditionID) (*domain.GameVersion, error)
	}{
		{"game", func(id values.EditionID) (*domain.GameVersion, error) {
			return repo.GetEditionGameVersionByGameID(ctx, id, values.NewGameIDFromUUID(gameID), repository.LockTypeNone)
		}},
		{"image", func(id values.EditionID) (*domain.GameVersion, error) {
			return repo.GetEditionGameVersionByImageID(ctx, id, values.GameImageIDFromUUID(imageID), repository.LockTypeNone)
		}},
		{"video", func(id values.EditionID) (*domain.GameVersion, error) {
			return repo.GetEditionGameVersionByVideoID(ctx, id, values.NewGameVideoIDFromUUID(videoID), repository.LockTypeNone)
		}},
		{"file", func(id values.EditionID) (*domain.GameVersion, error) {
			return repo.GetEditionGameVersionByFileID(ctx, id, values.NewGameFileIDFromUUID(fileID), repository.LockTypeNone)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := tt.lookup(values.NewEditionIDFromUUID(editionA))
			require.NoError(t, err)
			require.Equal(t, versionID, uuid.UUID(version.GetID()))
			_, err = tt.lookup(values.NewEditionIDFromUUID(editionB))
			require.ErrorIs(t, err, repository.ErrRecordNotFound)
		})
	}
}
