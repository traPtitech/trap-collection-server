package gorm2

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

func TestGameFeedbackGetGameFeedbacksByGameID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)
	gameFeedbackRepository := NewGameFeedback(testDB)

	var visibility schema.GameVisibilityTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameVisibilityTypePublic).Take(&visibility).Error)
	var imageType schema.GameImageTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameImageTypeJpeg).Take(&imageType).Error)
	var videoType schema.GameVideoTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameVideoTypeMp4).Take(&videoType).Error)
	gameID, versionID := values.NewGameID(), values.NewGameVersionID()
	imageID, videoID := values.NewGameImageID(), values.NewGameVideoID()
	activeID, archivedID, deletedID := values.NewFeedbackQuestionID(), values.NewFeedbackQuestionID(), values.NewFeedbackQuestionID()
	feedbackID := values.NewGameFeedbackID()
	createdAt, archivedAt, deletedAt := time.Now().Add(-time.Hour), time.Now().Add(-30*time.Minute), time.Now().Add(-15*time.Minute)

	game := schema.GameTable2{ID: uuid.UUID(gameID), Name: "feedback-list", Description: "description", VisibilityTypeID: visibility.ID, CreatedAt: createdAt}
	image := schema.GameImageTable2{ID: uuid.UUID(imageID), GameID: uuid.UUID(gameID), ImageTypeID: imageType.ID, CreatedAt: createdAt}
	video := schema.GameVideoTable2{ID: uuid.UUID(videoID), GameID: uuid.UUID(gameID), VideoTypeID: videoType.ID, CreatedAt: createdAt}
	version := schema.GameVersionTable2{ID: uuid.UUID(versionID), GameID: uuid.UUID(gameID), GameImageID: uuid.UUID(imageID), GameVideoID: uuid.UUID(videoID), Name: "v1", Description: "description", CreatedAt: createdAt}
	questions := []schema.GameFeedbackQuestionTable{
		{ID: uuid.UUID(activeID), GameID: uuid.UUID(gameID), QuestionText: "active", AnswerType: int(values.FeedbackAnswerTypeYesNo), QuestionOrder: 0, CreatedAt: createdAt},
		{ID: uuid.UUID(archivedID), GameID: uuid.UUID(gameID), QuestionText: "archived", AnswerType: int(values.FeedbackAnswerTypeFiveScale), QuestionOrder: 1, CreatedAt: createdAt, ArchivedAt: sql.NullTime{Time: archivedAt, Valid: true}},
		{ID: uuid.UUID(deletedID), GameID: uuid.UUID(gameID), QuestionText: "deleted", AnswerType: int(values.FeedbackAnswerTypeYesNo), QuestionOrder: 2, CreatedAt: createdAt, DeletedAt: gorm.DeletedAt{Time: deletedAt, Valid: true}},
	}
	feedback := schema.GameFeedbackTable{ID: uuid.UUID(feedbackID), GameVersionID: uuid.UUID(versionID), CreatedAt: createdAt}
	answers := []schema.GameFeedbackAnswerTable{{ID: uuid.New(), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(activeID), Answer: 1}, {ID: uuid.New(), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(archivedID), Answer: 5}, {ID: uuid.New(), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(deletedID), Answer: 0}}
	require.NoError(t, db.Create(&game).Error)
	require.NoError(t, db.Create(&image).Error)
	require.NoError(t, db.Create(&video).Error)
	require.NoError(t, db.Create(&version).Error)
	require.NoError(t, db.Create(&questions).Error)
	require.NoError(t, db.Create(&feedback).Error)
	require.NoError(t, db.Create(&answers).Error)
	t.Cleanup(func() {
		cleanupDB, cleanupErr := testDB.getDB(context.Background())
		require.NoError(t, cleanupErr)
		require.NoError(t, cleanupDB.Unscoped().Delete(&answers).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&feedback).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&questions).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&version).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&image).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&video).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&game).Error)
	})

	feedbacks, total, err := gameFeedbackRepository.GetGameFeedbacksByGameID(ctx, gameID, 50, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, feedbacks, 1)
	assert.Len(t, feedbacks[0].Answers, 2)
	returnedQuestions, err := gameFeedbackRepository.GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone)
	require.NoError(t, err)
	require.Len(t, returnedQuestions, 2)
	assert.True(t, returnedQuestions[1].IsArchived())
}
