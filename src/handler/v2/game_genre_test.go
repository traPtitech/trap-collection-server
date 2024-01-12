package v2

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestDeleteGameGenre(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreService := mock.NewMockGameGenre(ctrl)

	gameGenre := NewGameGenre(mockGameGenreService)

	type test struct {
		genreID            openapi.GameGenreIDInPath
		DeleteGameGenreErr error
		isErr              bool
		expectedErr        error
		statusCode         int
	}

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			genreID: uuid.New(),
		},
		"存在しないジャンルIDなので404": {
			genreID:            uuid.New(),
			DeleteGameGenreErr: service.ErrNoGameGenre,
			isErr:              true,
			statusCode:         http.StatusNotFound,
		},
		"DeleteGameGenreがエラーなのでエラー": {
			genreID:            uuid.New(),
			DeleteGameGenreErr: errors.New("error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockGameGenreService.
				EXPECT().
				DeleteGameGenre(gomock.Any(), values.GameGenreIDFromUUID(testCase.genreID)).
				Return(testCase.DeleteGameGenreErr)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/genres/%s", testCase.genreID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := gameGenre.DeleteGameGenre(c, testCase.genreID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echoHTTPError: %v", err)
					}
				} else if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
