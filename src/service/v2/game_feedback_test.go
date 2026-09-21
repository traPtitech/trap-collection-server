package v2

import (
	"context"
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

func TestGameFeedbackGetFeedbackConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type test struct {
		description                        string
		gameID                             values.GameID
		getGameErr                         error
		executeRepositoryGetFeedbackConfig bool
		repositoryGetFeedbackConfigResult  bool
		repositoryGetFeedbackConfigErr     error
		expectedEnabled                    bool
		expectedErr                        error
	}

	errUnexpected := errors.New("unexpected error")

	testCases := []test{
		{
			description:                        "enabledがtrueの設定を取得できる",
			gameID:                             values.NewGameID(),
			executeRepositoryGetFeedbackConfig: true,
			repositoryGetFeedbackConfigResult:  true,
			expectedEnabled:                    true,
		},
		{
			description:                        "enabledがfalseの設定を取得できる",
			gameID:                             values.NewGameID(),
			executeRepositoryGetFeedbackConfig: true,
			repositoryGetFeedbackConfigResult:  false,
			expectedEnabled:                    false,
		},
		{
			description:                        "設定レコードが存在しない場合enabled falseとして返す",
			gameID:                             values.NewGameID(),
			executeRepositoryGetFeedbackConfig: true,
			repositoryGetFeedbackConfigErr:     repository.ErrRecordNotFound,
			expectedEnabled:                    false,
		},
		{
			description: "ゲームが存在しない場合ErrInvalidGame",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			expectedErr: service.ErrInvalidGame,
		},
		{
			description: "ゲーム取得で予期しないエラーが起きた場合エラー",
			gameID:      values.NewGameID(),
			getGameErr:  errUnexpected,
			expectedErr: errUnexpected,
		},
		{
			description:                        "設定取得で予期しないエラーが起きた場合エラー",
			gameID:                             values.NewGameID(),
			executeRepositoryGetFeedbackConfig: true,
			repositoryGetFeedbackConfigErr:     errUnexpected,
			expectedErr:                        errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)

			gameFeedbackService := NewGameFeedback(
				mockGameRepository,
				mockGameFeedbackRepository,
				nil,
			)

			game := domain.NewGame(
				testCase.gameID,
				values.NewGameName("game"),
				values.NewGameDescription("description"),
				values.GameVisibilityTypePublic,
				time.Now(),
			)

			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(game, testCase.getGameErr)

			if testCase.executeRepositoryGetFeedbackConfig {
				mockGameFeedbackRepository.
					EXPECT().
					GetFeedbackConfig(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.repositoryGetFeedbackConfigResult, testCase.repositoryGetFeedbackConfigErr)
			}

			enabled, err := gameFeedbackService.GetFeedbackConfig(ctx, testCase.gameID)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedEnabled, enabled)
		})
	}
}

func TestGameFeedbackGetFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now()
	gameID := values.NewGameID()
	game := domain.NewGame(
		gameID,
		values.NewGameName("game"),
		values.NewGameDescription("description"),
		values.GameVisibilityTypePublic,
		now,
	)
	question := domain.NewFeedbackQuestion(
		values.NewFeedbackQuestionID(),
		gameID,
		values.NewFeedbackQuestionText("question"),
		values.FeedbackAnswerTypeYesNo,
		values.NewFeedbackQuestionOrder(0),
		now,
		nil,
	)

	type test struct {
		description                 string
		gameID                      values.GameID
		executeGetGame              bool
		getGameResult               *domain.Game
		getGameErr                  error
		executeGetFeedbackQuestions bool
		getFeedbackQuestionsResult  []*domain.FeedbackQuestion
		getFeedbackQuestionsErr     error
		expectedQuestions           []*domain.FeedbackQuestion
		expectedErr                 error
	}

	testCases := []test{
		{
			description:    "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			gameID:         gameID,
			executeGetGame: true,
			getGameErr:     repository.ErrRecordNotFound,
			expectedErr:    service.ErrInvalidGame,
		},
		{
			description:    "GetGameがエラーなのでエラー",
			gameID:         gameID,
			executeGetGame: true,
			getGameErr:     assert.AnError,
			expectedErr:    assert.AnError,
		},
		{
			description:                 "GetFeedbackQuestionsがエラーなのでエラー",
			gameID:                      gameID,
			executeGetGame:              true,
			getGameResult:               game,
			executeGetFeedbackQuestions: true,
			getFeedbackQuestionsErr:     assert.AnError,
			expectedErr:                 assert.AnError,
		},
		{
			description:                 "正常にフィードバック質問を取得できる",
			gameID:                      gameID,
			executeGetGame:              true,
			getGameResult:               game,
			executeGetFeedbackQuestions: true,
			getFeedbackQuestionsResult: []*domain.FeedbackQuestion{
				question,
			},
			expectedQuestions: []*domain.FeedbackQuestion{
				question,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameRepository := mockRepository.NewMockGameV2(ctrl)
			gameFeedbackRepository := mockRepository.NewMockGameFeedback(ctrl)
			gameFeedbackService := NewGameFeedback(gameRepository, gameFeedbackRepository, nil)

			if testCase.executeGetGame {
				gameRepository.
					EXPECT().
					GetGame(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.getGameResult, testCase.getGameErr)
			}

			if testCase.executeGetFeedbackQuestions {
				gameFeedbackRepository.
					EXPECT().
					GetFeedbackQuestions(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.getFeedbackQuestionsResult, testCase.getFeedbackQuestionsErr)
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
