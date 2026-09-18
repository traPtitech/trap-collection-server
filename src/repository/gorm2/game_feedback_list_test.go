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

func TestGameFeedbackGetGameFeedbacksByGameVersionID(t *testing.T) {
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
	feedbackID := values.GameFeedbackID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	newerFeedbackID := values.GameFeedbackID(uuid.MustParse("00000000-0000-0000-0000-000000000002"))
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
	newerFeedback := schema.GameFeedbackTable{ID: uuid.UUID(newerFeedbackID), GameVersionID: uuid.UUID(versionID), Comment: sql.NullString{String: "comment only", Valid: true}, CreatedAt: createdAt.Add(time.Minute)}
	answers := []schema.GameFeedbackAnswerTable{{ID: uuid.MustParse("00000000-0000-0000-0000-000000000011"), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(activeID), Answer: 1}, {ID: uuid.MustParse("00000000-0000-0000-0000-000000000012"), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(archivedID), Answer: 5}, {ID: uuid.MustParse("00000000-0000-0000-0000-000000000013"), FeedbackID: uuid.UUID(feedbackID), QuestionID: uuid.UUID(deletedID), Answer: 0}}
	require.NoError(t, db.Create(&game).Error)
	require.NoError(t, db.Create(&image).Error)
	require.NoError(t, db.Create(&video).Error)
	require.NoError(t, db.Create(&version).Error)
	require.NoError(t, db.Create(&questions).Error)
	require.NoError(t, db.Create(&feedback).Error)
	require.NoError(t, db.Create(&newerFeedback).Error)
	require.NoError(t, db.Create(&answers).Error)
	t.Cleanup(func() {
		cleanupDB, cleanupErr := testDB.getDB(context.Background())
		require.NoError(t, cleanupErr)
		require.NoError(t, cleanupDB.Unscoped().Delete(&answers).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete([]schema.GameFeedbackTable{feedback, newerFeedback}).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&questions).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&version).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&image).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&video).Error)
		require.NoError(t, cleanupDB.Unscoped().Delete(&game).Error)
	})

	feedbacks, total, err := gameFeedbackRepository.GetGameFeedbacksByGameVersionID(ctx, versionID, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, feedbacks, 1)
	assert.Equal(t, newerFeedbackID, feedbacks[0].Feedback.GetID())
	assert.Empty(t, feedbacks[0].Answers)
	require.NotNil(t, feedbacks[0].Feedback.GetComment())

	feedbacks, total, err = gameFeedbackRepository.GetGameFeedbacksByGameVersionID(ctx, versionID, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, feedbacks, 1)
	assert.Equal(t, feedbackID, feedbacks[0].Feedback.GetID())
	require.Len(t, feedbacks[0].Answers, 2)
	assert.Equal(t, []values.FeedbackQuestionID{activeID, archivedID}, []values.FeedbackQuestionID{feedbacks[0].Answers[0].GetQuestionID(), feedbacks[0].Answers[1].GetQuestionID()})
	assert.Equal(t, []int{1, 5}, []int{feedbacks[0].Answers[0].GetAnswer(), feedbacks[0].Answers[1].GetAnswer()})

	feedbacks, total, err = gameFeedbackRepository.GetGameFeedbacksByGameVersionID(ctx, versionID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Empty(t, feedbacks)

	feedbacks, total, err = gameFeedbackRepository.GetGameFeedbacksByGameID(ctx, gameID, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, feedbacks, 1)
	assert.Len(t, feedbacks[0].Answers, 2)
	returnedQuestions, err := gameFeedbackRepository.GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone)
	require.NoError(t, err)
	require.Len(t, returnedQuestions, 2)
	assert.True(t, returnedQuestions[1].IsArchived())
}
