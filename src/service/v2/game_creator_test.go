package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
	"go.uber.org/mock/gomock"
)

func TestGameCreatorService_GetGameCreators(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	testCases := map[string]struct {
		gameID                 values.GameID
		GetGameErr             error
		executeGetGameCreators bool
		creators               []*domain.GameCreatorWithJobs
		GetGameCreatorsErr     error
		err                    error
	}{
		// TODO: Add test cases.
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gameCreatorRepo := mock.NewMockGameCreator(ctrl)
			gameRepository := mock.NewMockGameV2(ctrl)
			gc := NewGameCreatorService(gameCreatorRepo, gameRepository)

			gameRepository.EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)
			if testCase.executeGetGameCreators {
				gameCreatorRepo.EXPECT().
					GetGameCreatorsByGameID(gomock.Any(), testCase.gameID).
					Return(testCase.creators, testCase.GetGameCreatorsErr)
			}

			creators, err := gc.GetGameCreators(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.creators, creators)
		})
	}
}
