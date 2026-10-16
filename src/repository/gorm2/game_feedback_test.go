package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

func TestGameFeedbackGetFeedbackConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	gameFeedbackRepository := NewGameFeedback(testDB)

	var visibility schema.GameVisibilityTypeTable
	err = db.
		Where("name = ?", schema.GameVisibilityTypePublic).
		Take(&visibility).Error
	require.NoError(t, err)

	gameIDEnabledTrue := values.NewGameID()
	gameIDEnabledFalse := values.NewGameID()
	gameIDNoConfig := values.NewGameID()

	games := []schema.GameTable2{
		{
			ID:               uuid.UUID(gameIDEnabledTrue),
			Name:             "feedback config enabled true",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDEnabledFalse),
			Name:             "feedback config enabled false",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDNoConfig),
			Name:             "feedback config no config",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
	}
	require.NoError(t, db.Create(&games).Error)

	configs := []schema.GameFeedbackConfigTable{
		{
			GameID:  uuid.UUID(gameIDEnabledTrue),
			Enabled: true,
		},
		{
			GameID:  uuid.UUID(gameIDEnabledFalse),
			Enabled: false,
		},
	}
	require.NoError(t, db.Create(&configs).Error)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupDB, err := testDB.getDB(cleanupCtx)
		require.NoError(t, err)

		require.NoError(t, cleanupDB.Unscoped().Delete(&configs).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&games).Error)
	})

	testCases := map[string]struct {
		gameID        values.GameID
		expectedValue bool
		expectedErr   error
	}{
		"enabledがtrueの設定を取得できる": {
			gameID:        gameIDEnabledTrue,
			expectedValue: true,
		},
		"enabledがfalseの設定を取得できる": {
			gameID:        gameIDEnabledFalse,
			expectedValue: false,
		},
		"設定レコードが存在しない場合ErrRecordNotFound": {
			gameID:      gameIDNoConfig,
			expectedErr: repository.ErrRecordNotFound,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			enabled, err := gameFeedbackRepository.GetFeedbackConfig(
				ctx,
				testCase.gameID,
				repository.LockTypeNone,
			)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedValue, enabled)
		})
	}
}

func TestGameFeedbackSaveFeedbackConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	gameFeedbackRepository := NewGameFeedback(testDB)

	var visibility schema.GameVisibilityTypeTable
	err = db.
		Where("name = ?", schema.GameVisibilityTypePublic).
		Take(&visibility).Error
	require.NoError(t, err)

	gameIDInsertTrue := values.NewGameID()
	gameIDInsertFalse := values.NewGameID()
	gameIDUpdateTrueToFalse := values.NewGameID()
	gameIDUpdateFalseToTrue := values.NewGameID()
	gameIDUpdateTrueToTrue := values.NewGameID()

	games := []schema.GameTable2{
		{
			ID:               uuid.UUID(gameIDInsertTrue),
			Name:             "save feedback config insert true",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDInsertFalse),
			Name:             "save feedback config insert false",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDUpdateTrueToFalse),
			Name:             "save feedback config update true to false",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDUpdateFalseToTrue),
			Name:             "save feedback config update false to true",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.UUID(gameIDUpdateTrueToTrue),
			Name:             "save feedback config update true to true",
			Description:      "description",
			VisibilityTypeID: visibility.ID,
			CreatedAt:        time.Now(),
		},
	}
	require.NoError(t, db.Create(&games).Error)

	initialConfigs := []schema.GameFeedbackConfigTable{
		{
			GameID:  uuid.UUID(gameIDUpdateTrueToFalse),
			Enabled: true,
		},
		{
			GameID:  uuid.UUID(gameIDUpdateFalseToTrue),
			Enabled: false,
		},
		{
			GameID:  uuid.UUID(gameIDUpdateTrueToTrue),
			Enabled: true,
		},
	}
	require.NoError(t, db.Create(&initialConfigs).Error)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupDB, err := testDB.getDB(cleanupCtx)
		require.NoError(t, err)

		require.NoError(t, cleanupDB.
			Unscoped().
			Where("game_id IN ?", []uuid.UUID{
				uuid.UUID(gameIDInsertTrue),
				uuid.UUID(gameIDInsertFalse),
				uuid.UUID(gameIDUpdateTrueToFalse),
				uuid.UUID(gameIDUpdateFalseToTrue),
				uuid.UUID(gameIDUpdateTrueToTrue),
			}).
			Delete(&schema.GameFeedbackConfigTable{}).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&games).Error)
	})

	testCases := map[string]struct {
		gameID      values.GameID
		enabled     bool
		wantEnabled bool
	}{
		"設定が存在しないゲームに対してenabled=trueで新規作成できる": {
			gameID:      gameIDInsertTrue,
			enabled:     true,
			wantEnabled: true,
		},
		"設定が存在しないゲームに対してenabled=falseで新規作成できる": {
			gameID:      gameIDInsertFalse,
			enabled:     false,
			wantEnabled: false,
		},
		"trueで既に存在する設定をfalseに更新できる": {
			gameID:      gameIDUpdateTrueToFalse,
			enabled:     false,
			wantEnabled: false,
		},
		"falseで既に存在する設定をtrueに更新できる": {
			gameID:      gameIDUpdateFalseToTrue,
			enabled:     true,
			wantEnabled: true,
		},
		"trueで既に存在する設定を同じくtrueで更新してもエラーにならない": {
			gameID:      gameIDUpdateTrueToTrue,
			enabled:     true,
			wantEnabled: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := gameFeedbackRepository.SaveFeedbackConfig(ctx, testCase.gameID, testCase.enabled)
			assert.NoError(t, err)

			var config schema.GameFeedbackConfigTable
			err = db.Where("game_id = ?", uuid.UUID(testCase.gameID)).Take(&config).Error
			require.NoError(t, err)
			assert.Equal(t, testCase.wantEnabled, config.Enabled)
		})
	}
}
