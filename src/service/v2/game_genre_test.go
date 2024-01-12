package v2

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestDeleteGameGenre(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	gameGenreService := NewGameGenre(mockGameGenreRepository)

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
