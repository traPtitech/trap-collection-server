package v2

import (
	"context"
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

func TestGameFeedbackPutFeedbackQuestions(t *testing.T) {
	gameID := values.NewGameID()
	existingID := values.NewFeedbackQuestionID()
	archivedID := values.NewFeedbackQuestionID()
	createdAt := time.Now().Add(-time.Hour)
	existing := domain.NewFeedbackQuestion(existingID, gameID, values.NewFeedbackQuestionText("old"), values.FeedbackAnswerTypeYesNo, 0, createdAt, nil)
	archived := domain.NewFeedbackQuestion(archivedID, gameID, values.NewFeedbackQuestionText("archive"), values.FeedbackAnswerTypeFiveScale, 1, createdAt, nil)
	game := domain.NewGame(gameID, values.NewGameName("game"), values.NewGameDescription("description"), values.GameVisibilityTypePublic, time.Now())

	testCases := map[string]struct {
		inputs     []service.FeedbackQuestionInput
		getGameErr error
		existing   []*domain.FeedbackQuestion
		wantErr    error
		setExpects func(*mockRepository.MockGameFeedback)
	}{
		"updates creates and archives in one transaction": {
			inputs: []service.FeedbackQuestionInput{
				{ID: &existingID, QuestionText: values.NewFeedbackQuestionText("updated"), AnswerType: values.FeedbackAnswerTypeFiveScale},
				{QuestionText: values.NewFeedbackQuestionText("new"), AnswerType: values.FeedbackAnswerTypeYesNo},
			},
			existing: []*domain.FeedbackQuestion{existing, archived},
			setExpects: func(repo *mockRepository.MockGameFeedback) {
				repo.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).Return([]*domain.FeedbackQuestion{existing, archived}, nil)
				repo.EXPECT().UpdateFeedbackQuestions(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, questions []*domain.FeedbackQuestion) error {
					assert.Len(t, questions, 1)
					assert.Equal(t, existingID, questions[0].GetID())
					assert.Equal(t, createdAt, questions[0].GetCreatedAt())
					return nil
				})
				repo.EXPECT().CreateFeedbackQuestions(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, questions []*domain.FeedbackQuestion) error {
					assert.Len(t, questions, 1)
					assert.Equal(t, gameID, questions[0].GetGameID())
					return nil
				})
				repo.EXPECT().ArchiveFeedbackQuestions(gomock.Any(), []values.FeedbackQuestionID{archivedID}).Return(nil)
			},
		},
		"unknown question ID is rejected": {
			inputs:   []service.FeedbackQuestionInput{{ID: &archivedID, QuestionText: values.NewFeedbackQuestionText("unknown"), AnswerType: values.FeedbackAnswerTypeYesNo}},
			existing: []*domain.FeedbackQuestion{existing},
			wantErr:  service.ErrInvalidFeedbackQuestion,
			setExpects: func(repo *mockRepository.MockGameFeedback) {
				repo.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).Return([]*domain.FeedbackQuestion{existing}, nil)
			},
		},
		"duplicate question ID is rejected": {
			inputs: []service.FeedbackQuestionInput{
				{ID: &existingID, QuestionText: values.NewFeedbackQuestionText("one"), AnswerType: values.FeedbackAnswerTypeYesNo},
				{ID: &existingID, QuestionText: values.NewFeedbackQuestionText("two"), AnswerType: values.FeedbackAnswerTypeYesNo},
			},
			existing: []*domain.FeedbackQuestion{existing},
			wantErr:  service.ErrDuplicateFeedbackQuestion,
			setExpects: func(repo *mockRepository.MockGameFeedback) {
				repo.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).Return([]*domain.FeedbackQuestion{existing}, nil)
			},
		},
		"invalid text is rejected": {
			inputs:  []service.FeedbackQuestionInput{{QuestionText: values.NewFeedbackQuestionText(""), AnswerType: values.FeedbackAnswerTypeYesNo}},
			wantErr: service.ErrInvalidFeedbackQuestion,
			setExpects: func(repo *mockRepository.MockGameFeedback) {
				repo.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID, repository.LockTypeNone).Return(nil, nil)
			},
		},
		"missing game is rejected": {
			getGameErr: repository.ErrRecordNotFound,
			wantErr:    service.ErrInvalidGame,
			setExpects: func(*mockRepository.MockGameFeedback) {},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			feedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameRepository.EXPECT().GetGame(gomock.Any(), gameID, repository.LockTypeRecord).Return(game, testCase.getGameErr)
			testCase.setExpects(feedbackRepository)

			service := NewGameFeedback(mockRepository.NewMockDB(ctrl), gameRepository, feedbackRepository)
			questions, err := service.PutFeedbackQuestions(context.Background(), gameID, testCase.inputs)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
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
	ctrl := gomock.NewController(t)
	gameID := values.NewGameID()
	gameRepository := mockRepository.NewMockGameV2(ctrl)
	feedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
	gameRepository.EXPECT().GetGame(gomock.Any(), gameID, repository.LockTypeNone).Return(nil, repository.ErrRecordNotFound)
	feedbackService := NewGameFeedback(mockRepository.NewMockDB(ctrl), gameRepository, feedbackRepository)
	_, err := feedbackService.GetFeedbackQuestions(context.Background(), gameID)
	assert.ErrorIs(t, err, service.ErrInvalidGame)

}
