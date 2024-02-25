package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestGetGameGenres(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	type test struct {
		isLoginUser      bool
		gameInfosRepo    []*repository.GameGenreInfo
		GetGameGenresErr error
		gameInfos        []*service.GameGenreInfo
		isErr            bool
		expectedErr      error
	}

	gameGenre1 := domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now())
	gameGenre2 := domain.NewGameGenre(values.NewGameGenreID(), "2D", time.Now())

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameInfosRepo: []*repository.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
			gameInfos:     []*service.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
		},
		"ログインしていてもエラー無し": {
			isLoginUser:   true,
			gameInfosRepo: []*repository.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
			gameInfos:     []*service.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
		},
		"複数でもエラー無し": {
			gameInfosRepo: []*repository.GameGenreInfo{
				{GameGenre: *gameGenre1, Num: 1},
				{GameGenre: *gameGenre2, Num: 3},
			},
			gameInfos: []*service.GameGenreInfo{
				{GameGenre: *gameGenre1, Num: 1},
				{GameGenre: *gameGenre2, Num: 3},
			},
		},
		"GetGameGenresがエラーなのでエラー": {
			GetGameGenresErr: errors.New("test"),
			isErr:            true,
		},
	}

	visibilitiesAll := []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate}
	visibilitiesNotLogin := []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			var argVisibilities []values.GameVisibility
			if testCase.isLoginUser {
				argVisibilities = visibilitiesAll
			} else {
				argVisibilities = visibilitiesNotLogin
			}
			mockGameGenreRepository.
				EXPECT().
				GetGameGenres(gomock.Any(), gomock.InAnyOrder(argVisibilities)).
				Return(testCase.gameInfosRepo, testCase.GetGameGenresErr)

			gameInfos, err := gameGenreService.GetGameGenres(ctx, testCase.isLoginUser)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			assert.Len(t, gameInfos, len(testCase.gameInfos))
			for i := range gameInfos {
				assert.Equal(t, testCase.gameInfos[i].GetID(), gameInfos[i].GetID())
				assert.Equal(t, testCase.gameInfos[i].GetName(), gameInfos[i].GetName())
				assert.Equal(t, testCase.gameInfos[i].Num, gameInfos[i].Num)
				assert.WithinDuration(t, testCase.gameInfos[i].GetCreatedAt(), gameInfos[i].GetCreatedAt(), time.Second)
			}

		})
	}
}

func TestDeleteGameGenre(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	type test struct {
		ID                 values.GameGenreID
		RemoveGameGenreErr error
		isErr              bool
		expectedErr        error
	}

	genreID := values.NewGameGenreID()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			ID: genreID,
		},
		"RemoveGameGenreがErrNoRecordDeletedなのでエラー": {
			ID:                 genreID,
			RemoveGameGenreErr: repository.ErrNoRecordDeleted,
			isErr:              true,
			expectedErr:        service.ErrNoGameGenre,
		},
		"RemoveGameGenreが他のエラーなのでエラー": {
			ID:                 genreID,
			RemoveGameGenreErr: errors.New("error"),
			isErr:              true,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockGameGenreRepository.
				EXPECT().
				RemoveGameGenre(ctx, testCase.ID).
				Return(testCase.RemoveGameGenreErr)

			err := gameGenreService.DeleteGameGenre(ctx, genreID)

			if !testCase.isErr {
				assert.NoError(t, err)
				return
			}

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
