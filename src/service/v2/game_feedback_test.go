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
