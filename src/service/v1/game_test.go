package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
)

func TestCreateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description     string
		name            values.GameName
		gameDescription values.GameDescription
		SaveGameErr     error
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "特に問題ないのでエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "nameが空でもエラーなし",
			name:            values.GameName(""),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "descriptionが空でもエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription(""),
		},
		{
			description:     "CreateGameがエラーなのでエラー",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			SaveGameErr:     errors.New("test"),
			isErr:           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				SaveGame(gomock.Any(), gomock.Any()).
				Return(testCase.SaveGameErr)

			game, err := gameVersionService.CreateGame(ctx, testCase.name, testCase.gameDescription)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.name, game.GetName())
			assert.Equal(t, testCase.gameDescription, game.GetDescription())
			assert.WithinDuration(t, time.Now(), game.GetCreatedAt(), time.Second)
		})
	}
}
