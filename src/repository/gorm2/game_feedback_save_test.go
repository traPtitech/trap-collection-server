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
		answerValues        []int
		duplicateFeedbackID bool
		duplicateAnswerID   bool
		duplicateQuestion   bool
		mismatchedFeedback  bool
		invalidVersion      bool
		invalidQuestion     bool
		wantAnyErr          bool
		wantErr             error
		wantFeedback        bool
		wantAnswers         int
	}{
		"creates feedback without answers and keeps nil comment null": {
			wantFeedback: true,
		},
		"creates feedback with empty comment": {
			comment:      feedbackComment(""),
			wantFeedback: true,
		},
		"creates feedback and answers with nonempty comment": {
			comment:      feedbackComment("excellent game"),
			answerCount:  2,
			answerValues: []int{0, 5},
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
			answerValues:       []int{5},
			mismatchedFeedback: true,
			wantAnyErr:         true,
		},
		"maps nonexistent game version to foreign key error": {
			invalidVersion: true,
			wantErr:        repository.ErrForeignKeyViolated,
		},
		"maps nonexistent question to foreign key error and rolls back feedback": {
			answerCount:     1,
			invalidQuestion: true,
			wantErr:         repository.ErrForeignKeyViolated,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := newGameFeedbackSaveFixture(t)
			feedbackID := values.NewGameFeedbackID()
			gameVersionID := fixture.gameVersionID
			if testCase.invalidVersion {
				gameVersionID = values.NewGameVersionID()
			}
			feedback := domain.NewGameFeedback(feedbackID, gameVersionID, testCase.comment, fixture.now)
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
				if testCase.invalidQuestion {
					questionID = values.NewFeedbackQuestionID()
				}
				answerFeedbackID := feedbackID
				if testCase.mismatchedFeedback {
					answerFeedbackID = fixture.existingFeedbackID
					questionID = fixture.questionIDs[1]
				}
				answerValue := i
				if len(testCase.answerValues) > i {
					answerValue = testCase.answerValues[i]
				}
				answers = append(answers, domain.NewGameFeedbackAnswer(answerID, answerFeedbackID, questionID, answerValue))
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
				assert.Equal(t, feedbackID.UUID(), savedFeedback.ID)
				assert.Equal(t, uuid.UUID(gameVersionID), savedFeedback.GameVersionID)
				assert.True(t, fixture.now.Equal(savedFeedback.CreatedAt))
				if testCase.comment == nil {
					assert.False(t, savedFeedback.Comment.Valid)
				} else {
					assert.Equal(t, string(*testCase.comment), savedFeedback.Comment.String)
					assert.True(t, savedFeedback.Comment.Valid)
				}
			} else if !testCase.duplicateFeedbackID {
				assert.Error(t, feedbackErr)
			} else {
				require.NoError(t, feedbackErr)
				assert.Equal(t, uuid.UUID(fixture.gameVersionID), savedFeedback.GameVersionID)
				assert.True(t, fixture.now.Equal(savedFeedback.CreatedAt))
			}

			var savedAnswers []schema.GameFeedbackAnswerTable
			require.NoError(t, fixture.db.Where("feedback_id = ?", feedbackID.UUID()).Find(&savedAnswers).Error)
			assert.Len(t, savedAnswers, testCase.wantAnswers)
			if testCase.wantFeedback {
				for _, answer := range answers {
					assert.Contains(t, savedAnswers, schema.GameFeedbackAnswerTable{
						ID: answer.GetID().UUID(), FeedbackID: feedbackID.UUID(), QuestionID: answer.GetQuestionID().UUID(), Answer: answer.GetAnswer(),
					})
				}
			}
			if testCase.mismatchedFeedback || testCase.duplicateAnswerID || testCase.duplicateQuestion {
				assertExistingGameFeedbackUnchanged(t, fixture)
			}
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

func TestCreateGameFeedbackFailedNestedSaveKeepsOuterTransaction(t *testing.T) {
	t.Parallel()

	fixture := newGameFeedbackSaveFixture(t)
	sentinel := schema.GameFeedbackTable{
		ID: uuid.New(), GameVersionID: uuid.UUID(fixture.gameVersionID), Comment: sql.NullString{String: "sentinel", Valid: true}, CreatedAt: fixture.now,
	}
	attempted := domain.NewGameFeedback(values.NewGameFeedbackID(), fixture.gameVersionID, nil, fixture.now)
	duplicateAnswer := domain.NewGameFeedbackAnswer(fixture.existingAnswerID, attempted.GetID(), fixture.questionIDs[1], 0)

	require.NoError(t, testDB.Transaction(t.Context(), nil, func(ctx context.Context) error {
		outerDB, err := testDB.getDB(ctx)
		if err != nil {
			return err
		}
		if err := outerDB.Create(&sentinel).Error; err != nil {
			return err
		}
		err = NewGameFeedback(testDB).CreateGameFeedback(ctx, attempted, []*domain.GameFeedbackAnswer{duplicateAnswer})
		require.ErrorIs(t, err, repository.ErrDuplicatedUniqueKey)
		return nil
	}))

	var savedSentinel schema.GameFeedbackTable
	require.NoError(t, fixture.db.Where("id = ?", sentinel.ID).Take(&savedSentinel).Error)
	assert.Equal(t, sentinel.ID, savedSentinel.ID)
	assert.Equal(t, sentinel.GameVersionID, savedSentinel.GameVersionID)
	assert.Equal(t, sentinel.Comment, savedSentinel.Comment)
	assert.True(t, sentinel.CreatedAt.Equal(savedSentinel.CreatedAt))
	assertGameFeedbackAbsent(t, fixture.db, attempted.GetID())
	assertExistingGameFeedbackUnchanged(t, fixture)
}

func feedbackComment(comment string) *values.FeedbackComment {
	value := values.NewFeedbackComment(comment)
	return &value
}

type gameFeedbackSaveFixture struct {
	db                 *gorm.DB
	gameVersionID      values.GameVersionID
	questionIDs        []values.FeedbackQuestionID
	existingFeedbackID values.GameFeedbackID
	existingAnswerID   values.GameFeedbackAnswerID
	now                time.Time
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

	now := time.Now().Truncate(time.Second)
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
		{ID: questionIDs[1].UUID(), GameID: uuid.UUID(gameID), QuestionText: "question 2", AnswerType: 1, QuestionOrder: 1, CreatedAt: now},
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

	return gameFeedbackSaveFixture{db: db, gameVersionID: gameVersionID, questionIDs: questionIDs, existingFeedbackID: existingFeedbackID, existingAnswerID: existingAnswerID, now: now}
}

func assertGameFeedbackAbsent(t *testing.T, db *gorm.DB, feedbackID values.GameFeedbackID) {
	t.Helper()
	var feedback schema.GameFeedbackTable
	assert.ErrorIs(t, db.Where("id = ?", feedbackID.UUID()).Take(&feedback).Error, gorm.ErrRecordNotFound)
	var answers []schema.GameFeedbackAnswerTable
	require.NoError(t, db.Where("feedback_id = ?", feedbackID.UUID()).Find(&answers).Error)
	assert.Empty(t, answers)
}

func assertExistingGameFeedbackUnchanged(t *testing.T, fixture gameFeedbackSaveFixture) {
	t.Helper()
	var feedback schema.GameFeedbackTable
	require.NoError(t, fixture.db.Where("id = ?", fixture.existingFeedbackID.UUID()).Take(&feedback).Error)
	assert.Equal(t, fixture.existingFeedbackID.UUID(), feedback.ID)
	assert.Equal(t, uuid.UUID(fixture.gameVersionID), feedback.GameVersionID)
	assert.Equal(t, sql.NullString{}, feedback.Comment)
	assert.True(t, fixture.now.Equal(feedback.CreatedAt))
	var answers []schema.GameFeedbackAnswerTable
	require.NoError(t, fixture.db.Where("feedback_id = ?", fixture.existingFeedbackID.UUID()).Find(&answers).Error)
	assert.Equal(t, []schema.GameFeedbackAnswerTable{{
		ID: fixture.existingAnswerID.UUID(), FeedbackID: fixture.existingFeedbackID.UUID(), QuestionID: fixture.questionIDs[0].UUID(), Answer: 1,
	}}, answers)
}

func questionIDsToUUIDs(ids []values.FeedbackQuestionID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.UUID())
	}
	return result
}
