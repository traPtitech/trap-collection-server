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
	"gorm.io/gorm"
)

func TestCreateGameFeedback(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		comment             *values.FeedbackComment
		answerCount         int
		duplicateFeedbackID bool
		duplicateAnswerID   bool
		duplicateQuestion   bool
		mismatchedFeedback  bool
		wantAnyErr          bool
		wantErr             error
		wantFeedback        bool
		wantAnswers         int
	}{
		"creates feedback without answers and keeps nil comment null": {
			wantFeedback: true,
		},
		"creates feedback and answers with empty comment": {
			comment:      feedbackComment(""),
			answerCount:  2,
			wantFeedback: true,
			wantAnswers:  2,
		},
		"maps duplicate feedback ID": {
			duplicateFeedbackID: true,
			wantErr:             repository.ErrDuplicatedUniqueKey,
		},
		"rolls back parent when answer ID is duplicated": {
			answerCount:       1,
			duplicateAnswerID: true,
			wantErr:           repository.ErrDuplicatedUniqueKey,
		},
		"rolls back parent when question is answered twice": {
			answerCount:       2,
			duplicateQuestion: true,
			wantErr:           repository.ErrDuplicatedUniqueKey,
		},
		"does not save answers belonging to another feedback": {
			answerCount:        1,
			mismatchedFeedback: true,
			wantAnyErr:         true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := newGameFeedbackSaveFixture(t)
			feedbackID := values.NewGameFeedbackID()
			feedback := domain.NewGameFeedback(feedbackID, fixture.gameVersionID, testCase.comment, fixture.now)
			if testCase.duplicateFeedbackID {
				require.NoError(t, fixture.db.Create(&schema.GameFeedbackTable{
					ID:            feedbackID.UUID(),
					GameVersionID: uuid.UUID(fixture.gameVersionID),
					CreatedAt:     fixture.now,
				}).Error)
			}

			answers := make([]*domain.GameFeedbackAnswer, 0, testCase.answerCount)
			for i := 0; i < testCase.answerCount; i++ {
				answerID := values.NewGameFeedbackAnswerID()
				if testCase.duplicateAnswerID && i == 0 {
					answerID = fixture.existingAnswerID
				}
				questionID := fixture.questionIDs[i]
				if testCase.duplicateQuestion {
					questionID = fixture.questionIDs[0]
				}
				answerFeedbackID := feedbackID
				if testCase.mismatchedFeedback {
					answerFeedbackID = values.NewGameFeedbackID()
				}
				answers = append(answers, domain.NewGameFeedbackAnswer(answerID, answerFeedbackID, questionID, i+1))
			}

			err := NewGameFeedback(testDB).CreateGameFeedback(t.Context(), feedback, answers)
			if testCase.wantAnyErr {
				assert.Error(t, err)
			} else if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
			}

			var savedFeedback schema.GameFeedbackTable
			feedbackErr := fixture.db.Where("id = ?", feedbackID.UUID()).Take(&savedFeedback).Error
			if testCase.wantFeedback {
				require.NoError(t, feedbackErr)
				if testCase.comment == nil {
					assert.False(t, savedFeedback.Comment.Valid)
				} else {
					assert.Equal(t, string(*testCase.comment), savedFeedback.Comment.String)
					assert.True(t, savedFeedback.Comment.Valid)
				}
			} else if !testCase.duplicateFeedbackID {
				assert.Error(t, feedbackErr)
			}

			var savedAnswers []schema.GameFeedbackAnswerTable
			require.NoError(t, fixture.db.Where("feedback_id = ?", feedbackID.UUID()).Find(&savedAnswers).Error)
			assert.Len(t, savedAnswers, testCase.wantAnswers)
		})
	}
}

func TestCreateGameFeedbackRespectsParentTransaction(t *testing.T) {
	t.Parallel()

	fixture := newGameFeedbackSaveFixture(t)
	feedback := domain.NewGameFeedback(values.NewGameFeedbackID(), fixture.gameVersionID, nil, fixture.now)
	answer := domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedback.GetID(), fixture.questionIDs[0], 1)

	err := testDB.Transaction(t.Context(), nil, func(ctx context.Context) error {
		require.NoError(t, NewGameFeedback(testDB).CreateGameFeedback(ctx, feedback, []*domain.GameFeedbackAnswer{answer}))
		return assert.AnError
	})
	require.ErrorIs(t, err, assert.AnError)

	var feedbackCount, answerCount int64
	require.NoError(t, fixture.db.Model(&schema.GameFeedbackTable{}).Where("id = ?", feedback.GetID().UUID()).Count(&feedbackCount).Error)
	require.NoError(t, fixture.db.Model(&schema.GameFeedbackAnswerTable{}).Where("id = ?", answer.GetID().UUID()).Count(&answerCount).Error)
	assert.Zero(t, feedbackCount)
	assert.Zero(t, answerCount)
}

func feedbackComment(comment string) *values.FeedbackComment {
	value := values.NewFeedbackComment(comment)
	return &value
}

type gameFeedbackSaveFixture struct {
	db               *gorm.DB
	gameVersionID    values.GameVersionID
	questionIDs      []values.FeedbackQuestionID
	existingAnswerID values.GameFeedbackAnswerID
	now              time.Time
}

func newGameFeedbackSaveFixture(t *testing.T) gameFeedbackSaveFixture {
	t.Helper()
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	var visibility schema.GameVisibilityTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameVisibilityTypePublic).Take(&visibility).Error)
	var imageType schema.GameImageTypeTable
	require.NoError(t, db.Where("name = ?", "jpeg").Take(&imageType).Error)
	var videoType schema.GameVideoTypeTable
	require.NoError(t, db.Where("name = ?", "mp4").Take(&videoType).Error)

	now := time.Now().Truncate(time.Microsecond)
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	questionIDs := []values.FeedbackQuestionID{values.NewFeedbackQuestionID(), values.NewFeedbackQuestionID()}
	existingFeedbackID := values.NewGameFeedbackID()
	existingAnswerID := values.NewGameFeedbackAnswerID()

	require.NoError(t, db.Create(&schema.GameTable2{
		ID: uuid.UUID(gameID), Name: "feedback save test", Description: "test", VisibilityTypeID: visibility.ID, CreatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&schema.GameImageTable2{
		ID: uuid.UUID(imageID), GameID: uuid.UUID(gameID), ImageTypeID: imageType.ID, CreatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&schema.GameVideoTable2{
		ID: uuid.UUID(videoID), GameID: uuid.UUID(gameID), VideoTypeID: videoType.ID, CreatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&schema.GameVersionTable2{
		ID: uuid.UUID(gameVersionID), GameID: uuid.UUID(gameID), GameImageID: uuid.UUID(imageID), GameVideoID: uuid.UUID(videoID), Name: "test", Description: "test", CreatedAt: now,
	}).Error)
	questions := []schema.GameFeedbackQuestionTable{
		{ID: questionIDs[0].UUID(), GameID: uuid.UUID(gameID), QuestionText: "question 1", AnswerType: 0, QuestionOrder: 0, CreatedAt: now},
		{ID: questionIDs[1].UUID(), GameID: uuid.UUID(gameID), QuestionText: "question 2", AnswerType: 0, QuestionOrder: 1, CreatedAt: now},
	}
	require.NoError(t, db.Create(&questions).Error)
	require.NoError(t, db.Create(&schema.GameFeedbackTable{
		ID: existingFeedbackID.UUID(), GameVersionID: uuid.UUID(gameVersionID), Comment: sql.NullString{}, CreatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&schema.GameFeedbackAnswerTable{
		ID: existingAnswerID.UUID(), FeedbackID: existingFeedbackID.UUID(), QuestionID: questionIDs[0].UUID(), Answer: 1,
	}).Error)

	t.Cleanup(func() {
		cleanupDB, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		require.NoError(t, cleanupDB.Exec(
			"DELETE game_feedback_answers FROM game_feedback_answers JOIN game_feedbacks ON game_feedback_answers.feedback_id = game_feedbacks.id WHERE game_feedbacks.game_version_id = ?",
			uuid.UUID(gameVersionID),
		).Error)
		require.NoError(t, cleanupDB.Where("game_version_id = ?", uuid.UUID(gameVersionID)).Delete(&schema.GameFeedbackTable{}).Error)
		require.NoError(t, cleanupDB.Where("id IN ?", questionIDsToUUIDs(questionIDs)).Delete(&schema.GameFeedbackQuestionTable{}).Error)
		require.NoError(t, cleanupDB.Where("id = ?", uuid.UUID(gameVersionID)).Delete(&schema.GameVersionTable2{}).Error)
		require.NoError(t, cleanupDB.Where("id = ?", uuid.UUID(imageID)).Delete(&schema.GameImageTable2{}).Error)
		require.NoError(t, cleanupDB.Where("id = ?", uuid.UUID(videoID)).Delete(&schema.GameVideoTable2{}).Error)
		require.NoError(t, cleanupDB.Where("id = ?", uuid.UUID(gameID)).Delete(&schema.GameTable2{}).Error)
	})

	return gameFeedbackSaveFixture{db: db, gameVersionID: gameVersionID, questionIDs: questionIDs, existingAnswerID: existingAnswerID, now: now}
}

func questionIDsToUUIDs(ids []values.FeedbackQuestionID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.UUID())
	}
	return result
}
