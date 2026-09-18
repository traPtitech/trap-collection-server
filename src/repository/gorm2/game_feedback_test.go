package gorm2

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
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

func TestGameFeedbackQuestions(t *testing.T) {
	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	var visibility schema.GameVisibilityTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameVisibilityTypePublic).Take(&visibility).Error)
	gameID := values.NewGameID()
	game := schema.GameTable2{ID: uuid.UUID(gameID), Name: "feedback questions", Description: "description", VisibilityTypeID: visibility.ID, CreatedAt: time.Now()}
	require.NoError(t, db.Create(&game).Error)

	questionID := values.NewFeedbackQuestionID()
	archivedID := values.NewFeedbackQuestionID()
	deletedID := values.NewFeedbackQuestionID()
	questions := []schema.GameFeedbackQuestionTable{
		{ID: questionID.UUID(), GameID: uuid.UUID(gameID), QuestionText: "visible", AnswerType: int(values.FeedbackAnswerTypeYesNo), QuestionOrder: 2, CreatedAt: time.Now()},
		{ID: archivedID.UUID(), GameID: uuid.UUID(gameID), QuestionText: "archived", AnswerType: int(values.FeedbackAnswerTypeYesNo), QuestionOrder: 0, CreatedAt: time.Now(), ArchivedAt: sql.NullTime{Time: time.Now(), Valid: true}},
		{ID: deletedID.UUID(), GameID: uuid.UUID(gameID), QuestionText: "deleted", AnswerType: int(values.FeedbackAnswerTypeYesNo), QuestionOrder: 1, CreatedAt: time.Now()},
	}
	require.NoError(t, db.Create(&questions).Error)
	require.NoError(t, db.Delete(&questions[2]).Error)
	t.Cleanup(func() {
		cleanupDB, cleanupErr := testDB.getDB(context.Background())
		require.NoError(t, cleanupErr)
		require.NoError(t, cleanupDB.Unscoped().Where("id IN ?", []uuid.UUID{questionID.UUID(), archivedID.UUID(), deletedID.UUID()}).Delete(&schema.GameFeedbackQuestionTable{}).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&game).Error)
	})

	repo := NewGameFeedback(testDB)
	actual, err := repo.GetFeedbackQuestions(ctx, gameID, repository.LockTypeNone)
	require.NoError(t, err)
	require.Len(t, actual, 1)
	assert.Equal(t, questionID, actual[0].GetID())

	newID := values.NewFeedbackQuestionID()
	newQuestion := domain.NewFeedbackQuestion(newID, gameID, values.NewFeedbackQuestionText("new"), values.FeedbackAnswerTypeFiveScale, 0, time.Now(), nil)
	require.NoError(t, repo.CreateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{newQuestion}))
	t.Cleanup(func() {
		require.NoError(t, db.Unscoped().Where("id = ?", newID.UUID()).Delete(&schema.GameFeedbackQuestionTable{}).Error)
	})

	updated := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("updated"), values.FeedbackAnswerTypeFiveScale, 3, questions[0].CreatedAt, nil)
	require.NoError(t, repo.UpdateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{updated}))
	require.NoError(t, repo.ArchiveFeedbackQuestions(ctx, []values.FeedbackQuestionID{questionID}))
	actual, err = repo.GetFeedbackQuestions(ctx, gameID, repository.LockTypeNone)
	require.NoError(t, err)
	assert.Len(t, actual, 1)
	assert.Equal(t, newID, actual[0].GetID())
}
