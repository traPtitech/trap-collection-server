package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGameFeedbackGetGameFeedbacks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	feedbackID := values.NewGameFeedbackID()
	versionID := values.NewGameVersionID()
	questionID := values.NewFeedbackQuestionID()
	answer := domain.NewGameFeedbackAnswer(
		values.NewGameFeedbackAnswerID(),
		feedbackID,
		questionID,
		1,
	)
	feedback := domain.NewGameFeedback(feedbackID, versionID, nil, time.Now())
	question := domain.NewFeedbackQuestion(
		questionID,
		gameID,
		values.NewFeedbackQuestionText("面白かったですか"),
		values.FeedbackAnswerTypeYesNo,
		values.NewFeedbackQuestionOrder(0),
		time.Now(),
		nil,
	)

	type test struct {
		description string
		limit       int
		offset      int

		executeGetGame bool
		getGameErr     error

		executeGetGameFeedbacksByGameID bool
		getGameFeedbacksByGameIDResult  []*repository.GameFeedbackWithAnswers
		getGameFeedbacksByGameIDTotal   int
		getGameFeedbacksByGameIDErr     error

		executeGetFeedbackQuestionsIncludingArchived bool
		getFeedbackQuestionsIncludingArchivedResult  []*domain.FeedbackQuestion
		getFeedbackQuestionsIncludingArchivedErr     error

		expectedDetails []*service.GameFeedbackDetail
		expectedTotal   int
		isErr           bool
		err             error
	}

	feedbackWithAnswer := &repository.GameFeedbackWithAnswers{
		Feedback: feedback,
		Answers:  []*domain.GameFeedbackAnswer{answer},
	}

	testCases := []test{
		{
			description:                     "正常にフィードバックを取得できる",
			limit:                           50,
			executeGetGame:                  true,
			executeGetGameFeedbacksByGameID: true,
			getGameFeedbacksByGameIDResult:  []*repository.GameFeedbackWithAnswers{feedbackWithAnswer},
			getGameFeedbacksByGameIDTotal:   3,
			executeGetFeedbackQuestionsIncludingArchived: true,
			getFeedbackQuestionsIncludingArchivedResult:  []*domain.FeedbackQuestion{question},
			expectedDetails: []*service.GameFeedbackDetail{
				{
					Feedback: feedback,
					Answers: []*service.GameFeedbackAnswerDetail{
						{
							Answer:       answer,
							QuestionText: question.GetQuestionText(),
							AnswerType:   question.GetAnswerType(),
						},
					},
				},
			},
			expectedTotal: 3,
		},
		{
			description: "limitが不正なのでErrInvalidLimit",
			limit:       101,
			isErr:       true,
			err:         service.ErrInvalidLimit,
		},
		{
			description:    "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			limit:          50,
			executeGetGame: true,
			getGameErr:     repository.ErrRecordNotFound,
			isErr:          true,
			err:            service.ErrInvalidGame,
		},
		{
			description:    "GetGameがエラーなのでエラー",
			limit:          50,
			executeGetGame: true,
			getGameErr:     assert.AnError,
			isErr:          true,
			err:            assert.AnError,
		},
		{
			description:                     "GetGameFeedbacksByGameIDがエラーなのでエラー",
			limit:                           50,
			executeGetGame:                  true,
			executeGetGameFeedbacksByGameID: true,
			getGameFeedbacksByGameIDErr:     assert.AnError,
			isErr:                           true,
			err:                             assert.AnError,
		},
		{
			description:                     "GetFeedbackQuestionsIncludingArchivedがエラーなのでエラー",
			limit:                           50,
			executeGetGame:                  true,
			executeGetGameFeedbacksByGameID: true,
			getGameFeedbacksByGameIDResult:  []*repository.GameFeedbackWithAnswers{feedbackWithAnswer},
			getGameFeedbacksByGameIDTotal:   3,
			executeGetFeedbackQuestionsIncludingArchived: true,
			getFeedbackQuestionsIncludingArchivedErr:     assert.AnError,
			isErr:                                        true,
			err:                                          assert.AnError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			gameFeedbackService := NewGameFeedback(gameRepository, gameFeedbackRepository, gameVersionRepository)

			if testCase.executeGetGame {
				gameRepository.
					EXPECT().
					GetGame(ctx, gameID, repository.LockTypeNone).
					Return(nil, testCase.getGameErr)
			}
			if testCase.executeGetGameFeedbacksByGameID {
				gameFeedbackRepository.
					EXPECT().
					GetGameFeedbacksByGameID(ctx, gameID, testCase.limit, testCase.offset).
					Return(
						testCase.getGameFeedbacksByGameIDResult,
						testCase.getGameFeedbacksByGameIDTotal,
						testCase.getGameFeedbacksByGameIDErr,
					)
			}
			if testCase.executeGetFeedbackQuestionsIncludingArchived {
				gameFeedbackRepository.
					EXPECT().
					GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone).
					Return(
						testCase.getFeedbackQuestionsIncludingArchivedResult,
						testCase.getFeedbackQuestionsIncludingArchivedErr,
					)
			}

			feedbackDetails, total, err := gameFeedbackService.GetGameFeedbacks(ctx, gameID, testCase.limit, testCase.offset)
			if testCase.isErr {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedTotal, total)
			assert.Equal(t, testCase.expectedDetails, feedbackDetails)
		})
	}
}

func TestGameFeedbackGetGameVersionFeedbacksRejectsDifferentGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	ctrl := gomock.NewController(t)
	gameRepository := mockRepository.NewMockGameV2(ctrl)
	gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
	gameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
	gameFeedbackService := NewGameFeedback(
		gameRepository,
		gameFeedbackRepository,
		gameVersionRepository,
	)

	gameRepository.EXPECT().
		GetGame(ctx, gameID, repository.LockTypeNone).
		Return(nil, nil)
	gameVersionRepository.EXPECT().
		GetGameVersionByID(ctx, gomock.Any(), repository.LockTypeNone).
		Return(&repository.GameVersionInfoWithGameID{GameID: values.NewGameID()}, nil)

	_, _, err := gameFeedbackService.GetGameVersionFeedbacks(
		ctx,
		gameID,
		values.NewGameVersionID(),
		50,
		0,
	)
	assert.True(t, errors.Is(err, service.ErrInvalidGameVersion))
}

func TestGameFeedbackGetGameVersionFeedbacks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	versionID := values.NewGameVersionID()
	feedbackID := values.NewGameFeedbackID()
	questionID := values.NewFeedbackQuestionID()
	feedback := domain.NewGameFeedback(feedbackID, versionID, nil, time.Now())
	answer := domain.NewGameFeedbackAnswer(
		values.NewGameFeedbackAnswerID(),
		feedbackID,
		questionID,
		5,
	)
	question := domain.NewFeedbackQuestion(
		questionID,
		gameID,
		values.NewFeedbackQuestionText("評価"),
		values.FeedbackAnswerTypeFiveScale,
		0,
		time.Now(),
		nil,
	)
	feedbacksWithAnswers := []*repository.GameFeedbackWithAnswers{
		{
			Feedback: feedback,
			Answers: []*domain.GameFeedbackAnswer{
				answer,
			},
		},
	}
	questions := []*domain.FeedbackQuestion{
		question,
	}

	type test struct {
		description                                  string
		executeGetGame                               bool
		getGameErr                                   error
		executeGetGameVersionByID                    bool
		getGameVersionByIDResult                     *repository.GameVersionInfoWithGameID
		getGameVersionByIDErr                        error
		executeGetGameFeedbacksByGameVersionID       bool
		getGameFeedbacksByGameVersionIDResult        []*repository.GameFeedbackWithAnswers
		getGameFeedbacksByGameVersionIDTotal         int
		getGameFeedbacksByGameVersionIDErr           error
		executeGetFeedbackQuestionsIncludingArchived bool
		getFeedbackQuestionsResult                   []*domain.FeedbackQuestion
		getFeedbackQuestionsErr                      error
		expectedError                                error
		expectedTotal                                int
		expectedAnswerType                           values.FeedbackAnswerType
	}

	testCases := []test{
		{
			description:                                  "正常にゲームバージョンのフィードバックを取得できる",
			executeGetGame:                               true,
			executeGetGameVersionByID:                    true,
			getGameVersionByIDResult:                     &repository.GameVersionInfoWithGameID{GameID: gameID},
			executeGetGameFeedbacksByGameVersionID:       true,
			getGameFeedbacksByGameVersionIDResult:        feedbacksWithAnswers,
			getGameFeedbacksByGameVersionIDTotal:         1,
			executeGetFeedbackQuestionsIncludingArchived: true,
			getFeedbackQuestionsResult:                   questions,
			expectedTotal:                                1,
			expectedAnswerType:                           values.FeedbackAnswerTypeFiveScale,
		},
		{
			description:    "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			executeGetGame: true,
			getGameErr:     repository.ErrRecordNotFound,
			expectedError:  service.ErrInvalidGame,
		},
		{
			description:               "GetGameVersionByIDがErrRecordNotFoundなのでErrInvalidGameVersion",
			executeGetGame:            true,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     repository.ErrRecordNotFound,
			expectedError:             service.ErrInvalidGameVersion,
		},
		{
			description:               "GetGameVersionByIDがエラーなのでエラーが返される",
			executeGetGame:            true,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     assert.AnError,
			expectedError:             assert.AnError,
		},
		{
			description:                            "GetGameFeedbacksByGameVersionIDがエラーなのでエラーが返される",
			executeGetGame:                         true,
			executeGetGameVersionByID:              true,
			getGameVersionByIDResult:               &repository.GameVersionInfoWithGameID{GameID: gameID},
			executeGetGameFeedbacksByGameVersionID: true,
			getGameFeedbacksByGameVersionIDErr:     assert.AnError,
			expectedError:                          assert.AnError,
		},
		{
			description:                                  "GetFeedbackQuestionsIncludingArchivedがエラーなのでエラーが返される",
			executeGetGame:                               true,
			executeGetGameVersionByID:                    true,
			getGameVersionByIDResult:                     &repository.GameVersionInfoWithGameID{GameID: gameID},
			executeGetGameFeedbacksByGameVersionID:       true,
			getGameFeedbacksByGameVersionIDResult:        feedbacksWithAnswers,
			getGameFeedbacksByGameVersionIDTotal:         1,
			executeGetFeedbackQuestionsIncludingArchived: true,
			getFeedbackQuestionsErr:                      assert.AnError,
			expectedError:                                assert.AnError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)
			gameFeedbackService := NewGameFeedback(
				gameRepository,
				gameFeedbackRepository,
				gameVersionRepository,
			)

			if testCase.executeGetGame {
				gameRepository.EXPECT().
					GetGame(ctx, gameID, repository.LockTypeNone).
					Return(nil, testCase.getGameErr)
			}
			if testCase.executeGetGameVersionByID {
				gameVersionRepository.EXPECT().
					GetGameVersionByID(ctx, versionID, repository.LockTypeNone).
					Return(testCase.getGameVersionByIDResult, testCase.getGameVersionByIDErr)
			}
			if testCase.executeGetGameFeedbacksByGameVersionID {
				gameFeedbackRepository.EXPECT().
					GetGameFeedbacksByGameVersionID(ctx, versionID, 7, 3).
					Return(
						testCase.getGameFeedbacksByGameVersionIDResult,
						testCase.getGameFeedbacksByGameVersionIDTotal,
						testCase.getGameFeedbacksByGameVersionIDErr,
					)
			}
			if testCase.executeGetFeedbackQuestionsIncludingArchived {
				gameFeedbackRepository.EXPECT().
					GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone).
					Return(testCase.getFeedbackQuestionsResult, testCase.getFeedbackQuestionsErr)
			}

			feedbackDetails, total, err := gameFeedbackService.GetGameVersionFeedbacks(ctx, gameID, versionID, 7, 3)
			if testCase.expectedError != nil {
				assert.ErrorIs(t, err, testCase.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedTotal, total)
			require.Len(t, feedbackDetails, 1)
			require.Len(t, feedbackDetails[0].Answers, 1)
			assert.Equal(t, testCase.expectedAnswerType, feedbackDetails[0].Answers[0].AnswerType)
		})
	}
}
