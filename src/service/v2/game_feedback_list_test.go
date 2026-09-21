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
	games := mockRepository.NewMockGameV2(ctrl)
	feedbacks := mockRepository.NewMockGameFeedback(ctrl)
	versions := mockRepository.NewMockGameVersionV2(ctrl)
	sut := NewGameFeedback(games, feedbacks, versions)

	games.EXPECT().GetGame(ctx, gameID, repository.LockTypeNone).Return(nil, nil)
	versions.EXPECT().GetGameVersionByID(ctx, gomock.Any(), repository.LockTypeNone).Return(&repository.GameVersionInfoWithGameID{GameID: values.NewGameID()}, nil)

	_, _, err := sut.GetGameVersionFeedbacks(ctx, gameID, values.NewGameVersionID(), 50, 0)
	assert.True(t, errors.Is(err, service.ErrInvalidGameVersion))
}

func TestGameFeedbackGetGameVersionFeedbacks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	gameID, versionID := values.NewGameID(), values.NewGameVersionID()
	feedbackID, questionID := values.NewGameFeedbackID(), values.NewFeedbackQuestionID()
	feedback := domain.NewGameFeedback(feedbackID, versionID, nil, time.Now())
	answer := domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedbackID, questionID, 5)
	question := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("評価"), values.FeedbackAnswerTypeFiveScale, 0, time.Now(), nil)

	testCases := map[string]struct {
		gameErr, versionErr, listErr, questionErr error
		wantErr                                   error
	}{
		"正常にフィードバックを取得できる": {},
		"GetGameがErrRecordNotFoundなのでErrInvalidGame": {
			gameErr: repository.ErrRecordNotFound,
			wantErr: service.ErrInvalidGame,
		},
		"GetGameVersionByIDがErrRecordNotFoundなのでErrInvalidGameVersion": {
			versionErr: repository.ErrRecordNotFound,
			wantErr:    service.ErrInvalidGameVersion,
		},
		"GetGameVersionByIDがエラーなのでエラー":                    {versionErr: assert.AnError, wantErr: assert.AnError},
		"GetGameFeedbacksByGameVersionIDがエラーなのでエラー":       {listErr: assert.AnError, wantErr: assert.AnError},
		"GetFeedbackQuestionsIncludingArchivedがエラーなのでエラー": {questionErr: assert.AnError, wantErr: assert.AnError},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			games, feedbacks, versions := mockRepository.NewMockGameV2(ctrl), mockRepository.NewMockGameFeedback(ctrl), mockRepository.NewMockGameVersionV2(ctrl)
			sut := NewGameFeedback(games, feedbacks, versions)
			games.EXPECT().GetGame(ctx, gameID, repository.LockTypeNone).Return(nil, testCase.gameErr)
			if testCase.gameErr != nil {
				_, _, err := sut.GetGameVersionFeedbacks(ctx, gameID, versionID, 7, 3)
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}
			versions.EXPECT().GetGameVersionByID(ctx, versionID, repository.LockTypeNone).Return(&repository.GameVersionInfoWithGameID{GameID: gameID}, testCase.versionErr)
			if testCase.versionErr == nil {
				feedbacks.EXPECT().GetGameFeedbacksByGameVersionID(ctx, versionID, 7, 3).Return([]*repository.GameFeedbackWithAnswers{{Feedback: feedback, Answers: []*domain.GameFeedbackAnswer{answer}}}, 1, testCase.listErr)
			}
			if testCase.versionErr == nil && testCase.listErr == nil {
				feedbacks.EXPECT().GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone).Return([]*domain.FeedbackQuestion{question}, testCase.questionErr)
			}
			got, total, err := sut.GetGameVersionFeedbacks(ctx, gameID, versionID, 7, 3)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, total)
			require.Len(t, got, 1)
			require.Len(t, got[0].Answers, 1)
			assert.Equal(t, values.FeedbackAnswerTypeFiveScale, got[0].Answers[0].AnswerType)
		})
	}
}
