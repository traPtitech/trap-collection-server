package v2

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetSeats(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		seats         []*domain.Seat
		GetSeatsError error
		resSeat       []*openapi.Seat
		isError       bool
		resStatus     int
	}{
		"正しく取得できる": {
			seats: []*domain.Seat{
				domain.NewSeat(1, values.SeatStatusEmpty),
				domain.NewSeat(2, values.SeatStatusInUse),
			},
			resSeat: []*openapi.Seat{
				{
					Id:     openapi.SeatID(1),
					Status: openapi.Empty,
				},
				{
					Id:     openapi.SeatID(2),
					Status: openapi.InUse,
				},
			},
			resStatus: http.StatusOK,
		},
		"GetSeatsがエラーなので500": {
			GetSeatsError: assert.AnError,
			isError:       true,
			resStatus:     http.StatusInternalServerError,
		},
		"無効な座席ステータスが含まれていたらそれを飛ばす": {
			seats: []*domain.Seat{
				domain.NewSeat(1, 100),
				domain.NewSeat(2, values.SeatStatusEmpty),
				domain.NewSeat(3, values.SeatStatusInUse),
			},
			resSeat: []*openapi.Seat{
				{
					Id:     openapi.SeatID(2),
					Status: openapi.Empty,
				},
				{
					Id:     openapi.SeatID(3),
					Status: openapi.InUse,
				},
			},
			resStatus: http.StatusOK,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seatMock := mock.NewMockSeat(ctrl)
			seatHandler := NewSeat(seatMock)

			req := httptest.NewRequest(http.MethodGet, "/seats", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			seatMock.EXPECT().GetSeats(gomock.Any()).
				Return(testCase.seats, testCase.GetSeatsError)

			err := seatHandler.GetSeats(c)

			if testCase.isError {
				assert.Error(t, err)
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, testCase.resStatus, httpErr.Code)
				}
			}

			if err != nil {
				return
			}

			assert.Equal(t, testCase.resStatus, rec.Code)

			var res []openapi.Seat
			err = json.NewDecoder(rec.Body).Decode(&res)
			assert.NoError(t, err)

			assert.Equal(t, len(testCase.resSeat), len(res))
			for i, seat := range res {
				assert.Equal(t, testCase.resSeat[i].Id, seat.Id)
				assert.Equal(t, testCase.resSeat[i].Status, seat.Status)
			}
		})
	}
}
