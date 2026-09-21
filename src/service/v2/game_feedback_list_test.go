package v2

import (
	"context"
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
	answer := domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedbackID, questionID, 1)
	feedback := domain.NewGameFeedback(feedbackID, versionID, nil, time.Now())
	question := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("面白かったですか"), values.FeedbackAnswerTypeYesNo, values.NewFeedbackQuestionOrder(0), time.Now(), nil)

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
			gameFeedbackService := NewGameFeedback(gameRepository, gameFeedbackRepository)

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
