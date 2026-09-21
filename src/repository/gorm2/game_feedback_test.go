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

func TestGameFeedbackGetFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	gameFeedbackRepository := NewGameFeedback(testDB)

	var visibility schema.GameVisibilityTypeTable
	require.NoError(t, db.
		Where("name = ?", schema.GameVisibilityTypePublic).
		Take(&visibility).Error)

	gameID := values.NewGameID()
	game := schema.GameTable2{
		ID:               uuid.UUID(gameID),
		Name:             "feedback questions",
		Description:      "description",
		VisibilityTypeID: visibility.ID,
		CreatedAt:        time.Now(),
	}
	require.NoError(t, db.Create(&game).Error)

	visibleQuestionID := values.NewFeedbackQuestionID()
	archivedQuestionID := values.NewFeedbackQuestionID()
	deletedQuestionID := values.NewFeedbackQuestionID()
	now := time.Now()
	questions := []schema.GameFeedbackQuestionTable{
		{
			ID:            visibleQuestionID.UUID(),
			GameID:        uuid.UUID(gameID),
			QuestionText:  "visible",
			AnswerType:    int(values.FeedbackAnswerTypeYesNo),
			QuestionOrder: 2,
			CreatedAt:     now,
		},
		{
			ID:            archivedQuestionID.UUID(),
			GameID:        uuid.UUID(gameID),
			QuestionText:  "archived",
			AnswerType:    int(values.FeedbackAnswerTypeYesNo),
			QuestionOrder: 0,
			CreatedAt:     now,
			ArchivedAt: sql.NullTime{
				Time:  now,
				Valid: true,
			},
		},
		{
			ID:            deletedQuestionID.UUID(),
			GameID:        uuid.UUID(gameID),
			QuestionText:  "deleted",
			AnswerType:    int(values.FeedbackAnswerTypeYesNo),
			QuestionOrder: 1,
			CreatedAt:     now,
		},
	}
	require.NoError(t, db.Create(&questions).Error)
	require.NoError(t, db.Delete(&questions[2]).Error)
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupDB, err := testDB.getDB(cleanupCtx)
		require.NoError(t, err)
		require.NoError(t, cleanupDB.
			Unscoped().
			Where("id IN ?", []uuid.UUID{
				visibleQuestionID.UUID(),
				archivedQuestionID.UUID(),
				deletedQuestionID.UUID(),
			}).
			Delete(&schema.GameFeedbackQuestionTable{}).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&game).Error)
	})

	type test struct {
		description           string
		gameID                values.GameID
		expectedQuestionID    values.FeedbackQuestionID
		expectedQuestionText  values.FeedbackQuestionText
		expectedAnswerType    values.FeedbackAnswerType
		expectedQuestionOrder values.FeedbackQuestionOrder
		expectedErr           error
	}

	testCases := []test{
		{
			description:           "正常にアーカイブ済みと削除済みの質問を除いて取得できる",
			gameID:                gameID,
			expectedQuestionID:    visibleQuestionID,
			expectedQuestionText:  values.NewFeedbackQuestionText("visible"),
			expectedAnswerType:    values.FeedbackAnswerTypeYesNo,
			expectedQuestionOrder: values.NewFeedbackQuestionOrder(2),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			questions, err := gameFeedbackRepository.GetFeedbackQuestions(
				ctx,
				testCase.gameID,
				repository.LockTypeNone,
			)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				assert.Nil(t, questions)
				return
			}

			assert.NoError(t, err)
			require.Len(t, questions, 1)
			assert.Equal(t, testCase.expectedQuestionID, questions[0].GetID())
			assert.Equal(t, testCase.expectedQuestionText, questions[0].GetQuestionText())
			assert.Equal(t, testCase.expectedAnswerType, questions[0].GetAnswerType())
			assert.Equal(t, testCase.expectedQuestionOrder, questions[0].GetQuestionOrder())
		})
	}
}
