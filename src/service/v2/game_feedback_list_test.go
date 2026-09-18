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
			sut := NewGameFeedback(games, feedbacks, versions)

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
		"取得できる":         {},
		"gameが存在しない":    {gameErr: repository.ErrRecordNotFound, wantErr: service.ErrInvalidGame},
		"versionが存在しない": {versionErr: repository.ErrRecordNotFound, wantErr: service.ErrInvalidGameVersion},
		"version取得に失敗":  {versionErr: assert.AnError, wantErr: assert.AnError},
		"一覧取得に失敗":       {listErr: assert.AnError, wantErr: assert.AnError},
		"質問取得に失敗":       {questionErr: assert.AnError, wantErr: assert.AnError},
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
