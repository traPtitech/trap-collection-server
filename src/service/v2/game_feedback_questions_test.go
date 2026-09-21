package v2

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

type transactionContextKey struct{}

type transactionDB struct{ transactionContext context.Context }

func (*transactionDB) Close() error          { return nil }
func (*transactionDB) Get() (*sql.DB, error) { return nil, nil }
func (db *transactionDB) Transaction(ctx context.Context, _ *sql.TxOptions, fn func(context.Context) error) error {
	db.transactionContext = context.WithValue(ctx, transactionContextKey{}, true)
	return fn(db.transactionContext)
}

func TestGameFeedbackPutFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	existingID := values.NewFeedbackQuestionID()
	archivedID := values.NewFeedbackQuestionID()
	createdAt := time.Now().Add(-time.Hour)
	existingQuestion := domain.NewFeedbackQuestion(
		existingID,
		gameID,
		values.NewFeedbackQuestionText("old"),
		values.FeedbackAnswerTypeYesNo,
		values.NewFeedbackQuestionOrder(0),
		createdAt,
		nil,
	)
	archivedQuestion := domain.NewFeedbackQuestion(
		archivedID,
		gameID,
		values.NewFeedbackQuestionText("archive"),
		values.FeedbackAnswerTypeFiveScale,
		values.NewFeedbackQuestionOrder(1),
		createdAt,
		nil,
	)
	game := domain.NewGame(
		gameID,
		values.NewGameName("game"),
		values.NewGameDescription("description"),
		values.GameVisibilityTypePublic,
		time.Now(),
	)

	type test struct {
		description string
		inputs      []service.FeedbackQuestionInput

		getGameErr error
		existing   []*domain.FeedbackQuestion

		executeHasFeedbackAnswers bool
		hasFeedbackAnswersResult  bool
		hasFeedbackAnswersErr     error

		executeUpdateFeedbackQuestions  bool
		updateFeedbackQuestionsErr      error
		executeCreateFeedbackQuestions  bool
		createFeedbackQuestionsErr      error
		executeArchiveFeedbackQuestions bool
		archiveFeedbackQuestionsErr     error

		expectedErr       error
		expectedQuestions []*domain.FeedbackQuestion
	}
	testCases := []test{
		{
			description: "正常に更新，作成，アーカイブを1つのトランザクションで実行できる",
			inputs: []service.FeedbackQuestionInput{
				{
					ID:           &existingID,
					QuestionText: values.NewFeedbackQuestionText("updated"),
					AnswerType:   values.FeedbackAnswerTypeFiveScale,
				},
				{
					QuestionText: values.NewFeedbackQuestionText("new"),
					AnswerType:   values.FeedbackAnswerTypeYesNo,
				},
			},
			existing:                        []*domain.FeedbackQuestion{existingQuestion, archivedQuestion},
			executeHasFeedbackAnswers:       true,
			executeUpdateFeedbackQuestions:  true,
			executeCreateFeedbackQuestions:  true,
			executeArchiveFeedbackQuestions: true,
		},
		{
			description: "存在しない質問IDなのでErrInvalidFeedbackQuestion",
			inputs: []service.FeedbackQuestionInput{
				{ID: &archivedID, QuestionText: values.NewFeedbackQuestionText("unknown"), AnswerType: values.FeedbackAnswerTypeYesNo},
			},
			existing:    []*domain.FeedbackQuestion{existingQuestion},
			expectedErr: service.ErrInvalidFeedbackQuestion,
		},
		{
			description: "質問IDが重複しているのでErrDuplicateFeedbackQuestion",
			inputs: []service.FeedbackQuestionInput{
				{ID: &existingID, QuestionText: values.NewFeedbackQuestionText("one"), AnswerType: values.FeedbackAnswerTypeYesNo},
				{ID: &existingID, QuestionText: values.NewFeedbackQuestionText("two"), AnswerType: values.FeedbackAnswerTypeYesNo},
			},
			existing:    []*domain.FeedbackQuestion{existingQuestion},
			expectedErr: service.ErrDuplicateFeedbackQuestion,
		},
		{
			description: "質問文が不正なのでErrInvalidFeedbackQuestion",
			inputs: []service.FeedbackQuestionInput{
				{QuestionText: values.NewFeedbackQuestionText(""), AnswerType: values.FeedbackAnswerTypeYesNo},
			},
			expectedErr: service.ErrInvalidFeedbackQuestion,
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			getGameErr:  repository.ErrRecordNotFound,
			expectedErr: service.ErrInvalidGame,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameFeedbackService := NewGameFeedback(mockRepository.NewMockDB(ctrl), gameRepository, gameFeedbackRepository)

			gameRepository.
				EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeRecord).
				Return(game, testCase.getGameErr)

			if testCase.getGameErr == nil {
				gameFeedbackRepository.
					EXPECT().
					GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).
					Return(testCase.existing, nil)
			}
			if testCase.executeHasFeedbackAnswers {
				gameFeedbackRepository.
					EXPECT().
					HasFeedbackAnswers(gomock.Any(), []values.FeedbackQuestionID{existingID}, repository.LockTypeRecord).
					Return(testCase.hasFeedbackAnswersResult, testCase.hasFeedbackAnswersErr)
			}
			if testCase.executeUpdateFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					UpdateFeedbackQuestions(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, questions []*domain.FeedbackQuestion) error {
						assert.Len(t, questions, 1)
						assert.Equal(t, existingID, questions[0].GetID())
						assert.Equal(t, createdAt, questions[0].GetCreatedAt())
						return testCase.updateFeedbackQuestionsErr
					})
			}
			if testCase.executeCreateFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					CreateFeedbackQuestions(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, questions []*domain.FeedbackQuestion) error {
						assert.Len(t, questions, 1)
						assert.Equal(t, gameID, questions[0].GetGameID())
						return testCase.createFeedbackQuestionsErr
					})
			}
			if testCase.executeArchiveFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					ArchiveFeedbackQuestions(gomock.Any(), []values.FeedbackQuestionID{archivedID}).
					Return(testCase.archiveFeedbackQuestionsErr)
			}

			questions, err := gameFeedbackService.PutFeedbackQuestions(ctx, gameID, testCase.inputs)
			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				assert.Nil(t, questions)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, questions, 2)
			assert.Equal(t, existingID, questions[0].GetID())
			assert.Equal(t, values.FeedbackQuestionOrder(0), questions[0].GetQuestionOrder())
			assert.Equal(t, values.FeedbackQuestionOrder(1), questions[1].GetQuestionOrder())
		})
	}
}

func TestGameFeedbackGetFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now()
	gameID := values.NewGameID()
	game := domain.NewGame(gameID, values.NewGameName("game"), values.NewGameDescription("description"), values.GameVisibilityTypePublic, now)
	question := domain.NewFeedbackQuestion(values.NewFeedbackQuestionID(), gameID, values.NewFeedbackQuestionText("question"), values.FeedbackAnswerTypeYesNo, values.NewFeedbackQuestionOrder(0), now, nil)

	type test struct {
		description                 string
		gameID                      values.GameID
		getGameResult               *domain.Game
		getGameErr                  error
		executeGetFeedbackQuestions bool
		getFeedbackQuestionsResult  []*domain.FeedbackQuestion
		getFeedbackQuestionsErr     error
		expectedQuestions           []*domain.FeedbackQuestion
		expectedErr                 error
	}
	testCases := []test{
		{description: "GetGameがErrRecordNotFoundなのでErrInvalidGame", gameID: gameID, getGameErr: repository.ErrRecordNotFound, expectedErr: service.ErrInvalidGame},
		{description: "GetGameがエラーなのでエラー", gameID: gameID, getGameErr: assert.AnError, expectedErr: assert.AnError},
		{description: "GetFeedbackQuestionsがエラーなのでエラー", gameID: gameID, getGameResult: game, executeGetFeedbackQuestions: true, getFeedbackQuestionsErr: assert.AnError, expectedErr: assert.AnError},
		{description: "正常にフィードバック質問を取得できる", gameID: gameID, getGameResult: game, executeGetFeedbackQuestions: true, getFeedbackQuestionsResult: []*domain.FeedbackQuestion{question}, expectedQuestions: []*domain.FeedbackQuestion{question}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameFeedbackService := NewGameFeedback(mockRepository.NewMockDB(ctrl), gameRepository, gameFeedbackRepository)
			gameRepository.EXPECT().GetGame(ctx, testCase.gameID, repository.LockTypeNone).Return(testCase.getGameResult, testCase.getGameErr)
			if testCase.executeGetFeedbackQuestions {
				gameFeedbackRepository.EXPECT().GetFeedbackQuestions(ctx, testCase.gameID, repository.LockTypeNone).Return(testCase.getFeedbackQuestionsResult, testCase.getFeedbackQuestionsErr)
			}
			questions, err := gameFeedbackService.GetFeedbackQuestions(ctx, testCase.gameID)
			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				assert.Nil(t, questions)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedQuestions, questions)
		})
	}
}

func TestGameFeedbackPutFeedbackQuestionsStopsAfterStageFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	gameID := values.NewGameID()
	questionID := values.NewFeedbackQuestionID()
	gameRepository := mockRepository.NewMockGameV2(ctrl)
	feedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
	gameRepository.EXPECT().GetGame(gomock.Any(), gameID, repository.LockTypeRecord).Return(
		domain.NewGame(gameID, values.NewGameName("game"), values.NewGameDescription("description"), values.GameVisibilityTypePublic, time.Now()), nil,
	)
	existing := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("old"), values.FeedbackAnswerTypeYesNo, 0, time.Now(), nil)
	feedbackRepository.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).Return([]*domain.FeedbackQuestion{existing}, nil)
	feedbackRepository.EXPECT().HasFeedbackAnswers(gomock.Any(), []values.FeedbackQuestionID{questionID}, repository.LockTypeRecord).Return(false, nil)
	stageErr := errors.New("update failed")
	feedbackRepository.EXPECT().UpdateFeedbackQuestions(gomock.Any(), gomock.Any()).Return(stageErr)

	feedbackService := NewGameFeedback(mockRepository.NewMockDB(ctrl), gameRepository, feedbackRepository)
	_, err := feedbackService.PutFeedbackQuestions(context.Background(), gameID, []service.FeedbackQuestionInput{{
		ID: questionIDPtr(questionID), QuestionText: values.NewFeedbackQuestionText("updated"), AnswerType: values.FeedbackAnswerTypeFiveScale,
	}})
	assert.ErrorIs(t, err, stageErr)
}

func TestGameFeedbackPutFeedbackQuestionsStopsAfterCreateAndArchiveFailures(t *testing.T) {
	stageErr := errors.New("stage failure")

	type test struct {
		description string
		inputs      []service.FeedbackQuestionInput

		executeCreateFeedbackQuestions  bool
		createFeedbackQuestionsErr      error
		executeArchiveFeedbackQuestions bool
		archiveFeedbackQuestionsErr     error

		expectedErr error
	}
	testCases := []test{
		{
			description: "CreateFeedbackQuestionsがエラーなのでArchiveFeedbackQuestionsを実行しない",
			inputs: []service.FeedbackQuestionInput{
				{
					QuestionText: values.NewFeedbackQuestionText("new"),
					AnswerType:   values.FeedbackAnswerTypeYesNo,
				},
			},
			executeCreateFeedbackQuestions: true,
			createFeedbackQuestionsErr:     stageErr,
			expectedErr:                    stageErr,
		},
		{
			description:                     "ArchiveFeedbackQuestionsがエラーなのでエラーを返す",
			inputs:                          []service.FeedbackQuestionInput{},
			executeCreateFeedbackQuestions:  true,
			executeArchiveFeedbackQuestions: true,
			archiveFeedbackQuestionsErr:     stageErr,
			expectedErr:                     stageErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gameID := values.NewGameID()
			questionID := values.NewFeedbackQuestionID()
			existingQuestion := domain.NewFeedbackQuestion(
				questionID,
				gameID,
				values.NewFeedbackQuestionText("old"),
				values.FeedbackAnswerTypeYesNo,
				values.NewFeedbackQuestionOrder(0),
				time.Now(),
				nil,
			)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)

			gameRepository.
				EXPECT().
				GetGame(gomock.Any(), gameID, repository.LockTypeRecord).
				Return(domain.NewGame(gameID, values.NewGameName("game"), values.NewGameDescription("description"), values.GameVisibilityTypePublic, time.Now()), nil)
			gameFeedbackRepository.
				EXPECT().
				GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).
				Return([]*domain.FeedbackQuestion{existingQuestion}, nil)
			gameFeedbackRepository.
				EXPECT().
				UpdateFeedbackQuestions(gomock.Any(), []*domain.FeedbackQuestion{}).
				Return(nil)

			if testCase.executeCreateFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					CreateFeedbackQuestions(gomock.Any(), gomock.Any()).
					Return(testCase.createFeedbackQuestionsErr)
			}
			if testCase.executeArchiveFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					ArchiveFeedbackQuestions(gomock.Any(), []values.FeedbackQuestionID{questionID}).
					Return(testCase.archiveFeedbackQuestionsErr)
			}

			gameFeedbackService := NewGameFeedback(&transactionDB{}, gameRepository, gameFeedbackRepository)
			_, err := gameFeedbackService.PutFeedbackQuestions(context.Background(), gameID, testCase.inputs)

			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func TestGameFeedbackPutFeedbackQuestionsUsesTransactionContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	gameID := values.NewGameID()
	db := &transactionDB{}
	gameRepo := mockRepository.NewMockGameV2(ctrl)
	repo := mockRepository.NewMockGameFeedback(ctrl)
	gameRepo.EXPECT().GetGame(gomock.Any(), gameID, repository.LockTypeRecord).DoAndReturn(func(ctx context.Context, _ values.GameID, _ repository.LockType) (*domain.Game, error) {
		assert.Equal(t, db.transactionContext, ctx)
		return domain.NewGame(gameID, values.NewGameName("game"), values.NewGameDescription("description"), values.GameVisibilityTypePublic, time.Now()), nil
	})
	repo.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).DoAndReturn(func(ctx context.Context, _ values.GameID, _ repository.LockType) ([]*domain.FeedbackQuestion, error) {
		assert.Equal(t, db.transactionContext, ctx)
		return nil, nil
	})
	repo.EXPECT().UpdateFeedbackQuestions(gomock.Any(), []*domain.FeedbackQuestion{}).DoAndReturn(func(ctx context.Context, _ []*domain.FeedbackQuestion) error {
		assert.Equal(t, db.transactionContext, ctx)
		return nil
	})
	repo.EXPECT().CreateFeedbackQuestions(gomock.Any(), []*domain.FeedbackQuestion{}).DoAndReturn(func(ctx context.Context, _ []*domain.FeedbackQuestion) error {
		assert.Equal(t, db.transactionContext, ctx)
		return nil
	})
	repo.EXPECT().ArchiveFeedbackQuestions(gomock.Any(), []values.FeedbackQuestionID{}).DoAndReturn(func(ctx context.Context, _ []values.FeedbackQuestionID) error {
		assert.Equal(t, db.transactionContext, ctx)
		return nil
	})
	_, err := NewGameFeedback(db, gameRepo, repo).PutFeedbackQuestions(context.Background(), gameID, []service.FeedbackQuestionInput{})
	assert.NoError(t, err)
}

func questionIDPtr(id values.FeedbackQuestionID) *values.FeedbackQuestionID {
	return &id
}
