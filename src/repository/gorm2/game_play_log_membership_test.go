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

func TestCreateGamePlayLogRequiresEditionVersionAndGame(t *testing.T) {
	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	repo := NewGamePlayLogV2(&DB{db: tx})
	now := time.Now().UTC().Truncate(time.Second)
	game, otherGame, edition, otherEdition, version, image, video := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{game, otherGame} {
		require.NoError(t, tx.Create(&schema.GameTable2{ID: id, Name: id.String(), Description: "test", VisibilityTypeID: 1, CreatedAt: now}).Error)
	}
	require.NoError(t, tx.Create(&schema.GameImageTable2{ID: image, GameID: game, ImageTypeID: 1, CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameVideoTable2{ID: video, GameID: game, VideoTypeID: 1, CreatedAt: now}).Error)
	require.NoError(t, tx.Create(&schema.GameVersionTable2{ID: version, GameID: game, GameImageID: image, GameVideoID: video, Name: "v1", Description: "test", CreatedAt: now}).Error)
	for _, id := range []uuid.UUID{edition, otherEdition} {
		require.NoError(t, tx.Create(&schema.EditionTable{ID: id, Name: id.String()[:32], CreatedAt: now}).Error)
	}
	require.NoError(t, tx.Exec("INSERT INTO edition_game_version_relations (edition_id, game_version_id) VALUES (?, ?)", edition, version).Error)
	cases := []struct {
		name          string
		edition, game uuid.UUID
		wantErr       bool
	}{
		{"一致する所属", edition, game, false},
		{"異なるエディション", otherEdition, game, true},
		{"異なるゲーム", edition, otherGame, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			log := domain.NewGamePlayLog(values.GamePlayLogID(id), values.NewEditionIDFromUUID(tt.edition), values.NewGameIDFromUUID(tt.game), values.NewGameVersionIDFromUUID(version), now, nil, now, now)
			err := repo.CreateGamePlayLog(ctx, log)
			if tt.wantErr {
				require.ErrorIs(t, err, repository.ErrRecordNotFound)
			} else {
				require.NoError(t, err)
			}
			var count int64
			require.NoError(t, tx.Model(&schema.GamePlayLogTable{}).Where("id = ?", id).Count(&count).Error)
			if tt.wantErr {
				require.Zero(t, count)
			} else {
				require.EqualValues(t, 1, count)
			}
		})
	}
}
