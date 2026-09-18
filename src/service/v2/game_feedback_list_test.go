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

	testCases := map[string]struct {
		limit, offset int
		gameErr       error
		feedbackErr   error
		questionErr   error
		wantErr       error
	}{
		"取得できる":         {},
		"limitが不正":      {limit: 101, wantErr: service.ErrInvalidLimit},
		"gameが存在しない":    {gameErr: repository.ErrRecordNotFound, wantErr: service.ErrInvalidGame},
		"feedback取得に失敗": {feedbackErr: assert.AnError, wantErr: assert.AnError},
		"質問取得に失敗":       {questionErr: assert.AnError, wantErr: assert.AnError},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			games := mockRepository.NewMockGameV2(ctrl)
			feedbacks := mockRepository.NewMockGameFeedback(ctrl)
			versions := mockRepository.NewMockGameVersionV2(ctrl)
			sut := NewGameFeedback(nil, games, feedbacks, versions)

			limit := testCase.limit
			if limit == 0 {
				limit = 50
			}
			if testCase.wantErr == service.ErrInvalidLimit {
				require.ErrorIs(t, func() error {
					_, _, err := sut.GetGameFeedbacks(ctx, gameID, testCase.limit, testCase.offset)
					return err
				}(), service.ErrInvalidLimit)
				return
			}

			games.EXPECT().GetGame(ctx, gameID, repository.LockTypeNone).Return(nil, testCase.gameErr)
			if testCase.gameErr == nil {
				feedbacks.EXPECT().GetGameFeedbacksByGameID(ctx, gameID, limit, testCase.offset).
					Return([]*repository.GameFeedbackWithAnswers{{Feedback: feedback, Answers: []*domain.GameFeedbackAnswer{answer}}}, 3, testCase.feedbackErr)
			}
			if testCase.gameErr == nil && testCase.feedbackErr == nil {
				feedbacks.EXPECT().GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone).
					Return([]*domain.FeedbackQuestion{question}, testCase.questionErr)
			}

			got, total, err := sut.GetGameFeedbacks(ctx, gameID, limit, testCase.offset)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 3, total)
			require.Len(t, got, 1)
			require.Len(t, got[0].Answers, 1)
			assert.Equal(t, question.GetQuestionText(), got[0].Answers[0].QuestionText)
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
	sut := NewGameFeedback(nil, games, feedbacks, versions)

	games.EXPECT().GetGame(ctx, gameID, repository.LockTypeNone).Return(nil, nil)
	versions.EXPECT().GetGameVersionByID(ctx, gomock.Any(), repository.LockTypeNone).Return(&repository.GameVersionInfoWithGameID{GameID: values.NewGameID()}, nil)

	_, _, err := sut.GetGameVersionFeedbacks(ctx, gameID, values.NewGameVersionID(), 50, 0)
	assert.True(t, errors.Is(err, service.ErrInvalidGameVersion))
}
