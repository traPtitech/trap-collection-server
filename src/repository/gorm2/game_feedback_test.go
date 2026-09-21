package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"github.com/traPtitech/trap-collection-server/src/service"
	servicev2 "github.com/traPtitech/trap-collection-server/src/service/v2"
	"gorm.io/gorm"
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

func TestFeedbackQuestionAnsweredTypeChange(t *testing.T) {
	type test struct {
		description       string
		initialAnswerType values.FeedbackAnswerType
		nextAnswerType    values.FeedbackAnswerType
		answer            int
		hasAnswer         bool
		expectedErr       error
	}
	testCases := []test{
		{
			description:       "回答済みのfiveScaleをyesNoへ変更するのでErrFeedbackQuestionAnswerTypeChange",
			initialAnswerType: values.FeedbackAnswerTypeFiveScale,
			nextAnswerType:    values.FeedbackAnswerTypeYesNo,
			answer:            5,
			hasAnswer:         true,
			expectedErr:       service.ErrFeedbackQuestionAnswerTypeChange,
		},
		{
			description:       "回答済みのyesNoをfiveScaleへ変更するのでErrFeedbackQuestionAnswerTypeChange",
			initialAnswerType: values.FeedbackAnswerTypeYesNo,
			nextAnswerType:    values.FeedbackAnswerTypeFiveScale,
			answer:            0,
			hasAnswer:         true,
			expectedErr:       service.ErrFeedbackQuestionAnswerTypeChange,
		},
		{
			description:       "未回答なのでanswerTypeを変更できる",
			initialAnswerType: values.FeedbackAnswerTypeYesNo,
			nextAnswerType:    values.FeedbackAnswerTypeFiveScale,
			answer:            0,
			hasAnswer:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			fixture := newQuestionTypeFixture(
				t,
				int(testCase.initialAnswerType),
				testCase.answer,
				testCase.hasAnswer,
			)
			gameFeedbackService := servicev2.NewGameFeedback(
				testDB,
				NewGameV2(testDB),
				NewGameFeedback(testDB),
			)
			questionID := fixture.questionID
			inputs := []service.FeedbackQuestionInput{
				{
					ID:           &questionID,
					QuestionText: values.NewFeedbackQuestionText("question"),
					AnswerType:   testCase.nextAnswerType,
				},
			}

			_, err := gameFeedbackService.PutFeedbackQuestions(t.Context(), fixture.gameID, inputs)
			if testCase.expectedErr != nil {
				require.ErrorIs(t, err, service.ErrFeedbackQuestionAnswerTypeChange)
			} else {
				require.NoError(t, err)
			}

			var question schema.GameFeedbackQuestionTable
			require.NoError(t, fixture.db.Where("id = ?", questionID.UUID()).Take(&question).Error)
			if testCase.expectedErr != nil {
				assert.Equal(t, int(testCase.initialAnswerType), question.AnswerType)
			} else {
				assert.Equal(t, int(testCase.nextAnswerType), question.AnswerType)
			}
		})
	}
}

type questionTypeFixture struct {
	db         *gorm.DB
	gameID     values.GameID
	questionID values.FeedbackQuestionID
}

type createFailingFeedbackRepository struct {
	repository.GameFeedback
	err error
}

func (r createFailingFeedbackRepository) CreateFeedbackQuestions(context.Context, []*domain.FeedbackQuestion) error {
	return r.err
}

func TestPutFeedbackQuestionsRollsBackUpdateWhenCreateFails(t *testing.T) {
	fixture := newQuestionTypeFixture(t, int(values.FeedbackAnswerTypeYesNo), 0, false)
	baseRepository := NewGameFeedback(testDB)
	createErr := errors.New("create failed")
	feedbackService := servicev2.NewGameFeedback(
		testDB,
		NewGameV2(testDB),
		createFailingFeedbackRepository{GameFeedback: baseRepository, err: createErr},
	)
	questionID := fixture.questionID
	_, err := feedbackService.PutFeedbackQuestions(t.Context(), fixture.gameID, []service.FeedbackQuestionInput{
		{ID: &questionID, QuestionText: values.NewFeedbackQuestionText("updated"), AnswerType: values.FeedbackAnswerTypeYesNo},
		{QuestionText: values.NewFeedbackQuestionText("new"), AnswerType: values.FeedbackAnswerTypeFiveScale},
	})
	require.ErrorIs(t, err, createErr)

	var persisted schema.GameFeedbackQuestionTable
	require.NoError(t, fixture.db.Where("id = ?", questionID.UUID()).Take(&persisted).Error)
	assert.Equal(t, "question", persisted.QuestionText)
	assert.Equal(t, 0, persisted.QuestionOrder)
	assert.False(t, persisted.ArchivedAt.Valid)
	var count int64
	require.NoError(t, fixture.db.Model(&schema.GameFeedbackQuestionTable{}).Where("game_id = ?", uuid.UUID(fixture.gameID)).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func newQuestionTypeFixture(t *testing.T, answerType, answer int, withAnswer bool) questionTypeFixture {
	t.Helper()
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)
	var visibility schema.GameVisibilityTypeTable
	require.NoError(t, db.Where("name = ?", schema.GameVisibilityTypePublic).Take(&visibility).Error)
	var imageType schema.GameImageTypeTable
	require.NoError(t, db.Where("name = ?", "jpeg").Take(&imageType).Error)
	var videoType schema.GameVideoTypeTable
	require.NoError(t, db.Where("name = ?", "mp4").Take(&videoType).Error)
	now := time.Now()
	gameID := values.NewGameID()
	imageID := values.NewGameImageID()
	videoID := values.NewGameVideoID()
	versionID := values.NewGameVersionID()
	questionID := values.NewFeedbackQuestionID()
	feedbackID := values.NewGameFeedbackID()
	game := schema.GameTable2{
		ID:               uuid.UUID(gameID),
		Name:             "question type",
		Description:      "test",
		VisibilityTypeID: visibility.ID,
		CreatedAt:        now,
	}
	require.NoError(t, db.Create(&game).Error)

	gameImage := schema.GameImageTable2{
		ID:          uuid.UUID(imageID),
		GameID:      uuid.UUID(gameID),
		ImageTypeID: imageType.ID,
		CreatedAt:   now,
	}
	require.NoError(t, db.Create(&gameImage).Error)

	gameVideo := schema.GameVideoTable2{
		ID:          uuid.UUID(videoID),
		GameID:      uuid.UUID(gameID),
		VideoTypeID: videoType.ID,
		CreatedAt:   now,
	}
	require.NoError(t, db.Create(&gameVideo).Error)

	gameVersion := schema.GameVersionTable2{
		ID:          uuid.UUID(versionID),
		GameID:      uuid.UUID(gameID),
		GameImageID: uuid.UUID(imageID),
		GameVideoID: uuid.UUID(videoID),
		Name:        "test",
		Description: "test",
		CreatedAt:   now,
	}
	require.NoError(t, db.Create(&gameVersion).Error)

	question := schema.GameFeedbackQuestionTable{
		ID:            questionID.UUID(),
		GameID:        uuid.UUID(gameID),
		QuestionText:  "question",
		AnswerType:    answerType,
		QuestionOrder: 0,
		CreatedAt:     now,
	}
	require.NoError(t, db.Create(&question).Error)

	if withAnswer {
		feedback := schema.GameFeedbackTable{
			ID:            feedbackID.UUID(),
			GameVersionID: uuid.UUID(versionID),
			CreatedAt:     now,
		}
		require.NoError(t, db.Create(&feedback).Error)

		feedbackAnswer := schema.GameFeedbackAnswerTable{
			ID:         values.NewGameFeedbackAnswerID().UUID(),
			FeedbackID: feedbackID.UUID(),
			QuestionID: questionID.UUID(),
			Answer:     answer,
		}
		require.NoError(t, db.Create(&feedbackAnswer).Error)
	}
	t.Cleanup(func() {
		c, _ := testDB.getDB(context.Background())
		c.Exec("DELETE game_feedback_answers FROM game_feedback_answers JOIN game_feedbacks ON game_feedback_answers.feedback_id = game_feedbacks.id WHERE game_feedbacks.game_version_id = ?", uuid.UUID(versionID))
		c.Where("game_version_id = ?", uuid.UUID(versionID)).Delete(&schema.GameFeedbackTable{})
		c.Unscoped().Delete(&schema.GameFeedbackQuestionTable{}, "id = ?", questionID.UUID())
		c.Delete(&schema.GameVersionTable2{}, "id = ?", uuid.UUID(versionID))
		c.Delete(&schema.GameImageTable2{}, "id = ?", uuid.UUID(imageID))
		c.Delete(&schema.GameVideoTable2{}, "id = ?", uuid.UUID(videoID))
		c.Delete(&schema.GameTable2{}, "id = ?", uuid.UUID(gameID))
	})
	return questionTypeFixture{
		db:         db,
		gameID:     gameID,
		questionID: questionID,
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
	assert.Equal(t, values.NewFeedbackQuestionText("visible"), actual[0].GetQuestionText())
	assert.Equal(t, values.NewFeedbackQuestionOrder(2), actual[0].GetQuestionOrder())

	newID := values.NewFeedbackQuestionID()
	newQuestion := domain.NewFeedbackQuestion(newID, gameID, values.NewFeedbackQuestionText("new"), values.FeedbackAnswerTypeFiveScale, 0, time.Now(), nil)
	require.NoError(t, repo.CreateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{newQuestion}))
	t.Cleanup(func() {
		require.NoError(t, db.Unscoped().Where("id = ?", newID.UUID()).Delete(&schema.GameFeedbackQuestionTable{}).Error)
	})

	updated := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("updated"), values.FeedbackAnswerTypeFiveScale, 3, questions[0].CreatedAt, nil)
	require.NoError(t, repo.UpdateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{updated}))
	require.NoError(t, repo.ArchiveFeedbackQuestions(ctx, []values.FeedbackQuestionID{questionID}))
	var persisted schema.GameFeedbackQuestionTable
	require.NoError(t, db.Unscoped().Where("id = ?", questionID.UUID()).Take(&persisted).Error)
	assert.Equal(t, "updated", persisted.QuestionText)
	assert.Equal(t, int(values.FeedbackAnswerTypeFiveScale), persisted.AnswerType)
	assert.Equal(t, 3, persisted.QuestionOrder)
	assert.True(t, persisted.ArchivedAt.Valid)
	assert.False(t, persisted.DeletedAt.Valid)

	missing := domain.NewFeedbackQuestion(values.NewFeedbackQuestionID(), gameID, values.NewFeedbackQuestionText("missing"), values.FeedbackAnswerTypeYesNo, 0, time.Now(), nil)
	err = repo.UpdateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{missing, updated})
	assert.ErrorIs(t, err, repository.ErrNoRecordUpdated)
	deleted := domain.NewFeedbackQuestion(deletedID, gameID, values.NewFeedbackQuestionText("deleted"), values.FeedbackAnswerTypeYesNo, 0, time.Now(), nil)
	err = repo.UpdateFeedbackQuestions(ctx, []*domain.FeedbackQuestion{deleted})
	assert.ErrorIs(t, err, repository.ErrNoRecordUpdated)
	actual, err = repo.GetFeedbackQuestions(ctx, gameID, repository.LockTypeNone)
	require.NoError(t, err)
	assert.Len(t, actual, 1)
	assert.Equal(t, newID, actual[0].GetID())

	rolledBackID := values.NewFeedbackQuestionID()
	err = testDB.Transaction(ctx, nil, func(txCtx context.Context) error {
		question := domain.NewFeedbackQuestion(rolledBackID, gameID, values.NewFeedbackQuestionText("rollback"), values.FeedbackAnswerTypeYesNo, 9, time.Now(), nil)
		require.NoError(t, repo.CreateFeedbackQuestions(txCtx, []*domain.FeedbackQuestion{question}))
		return errors.New("rollback")
	})
	require.Error(t, err)
	var rolledBack schema.GameFeedbackQuestionTable
	err = db.Unscoped().Where("id = ?", rolledBackID.UUID()).Take(&rolledBack).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
