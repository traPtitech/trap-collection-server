package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
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

			c, _, rec := setupTestRequest(t, http.MethodGet, "/seats", nil)

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

func TestPostSeat(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req                  openapi.PostSeatRequest
		executeUpdateSeatNum bool
		seats                []*domain.Seat
		UpdateSeatNumError   error
		resSeat              []*openapi.Seat
		isError              bool
		resStatus            int
	}{
		"正しく変更できる": {
			req: openapi.PostSeatRequest{
				Num: 3,
			},
			executeUpdateSeatNum: true,
			seats: []*domain.Seat{
				domain.NewSeat(1, values.SeatStatusEmpty),
				domain.NewSeat(2, values.SeatStatusInUse),
				domain.NewSeat(3, values.SeatStatusEmpty),
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
				{
					Id:     openapi.SeatID(3),
					Status: openapi.Empty,
				},
			},
			resStatus: http.StatusOK,
		},
		"UpdateSeatNumがエラーなので500": {
			req: openapi.PostSeatRequest{
				Num: 3,
			},
			executeUpdateSeatNum: true,
			UpdateSeatNumError:   assert.AnError,
			isError:              true,
			resStatus:            http.StatusInternalServerError,
		},
		"UpdateSeatNumで0を指定した場合": {
			req: openapi.PostSeatRequest{
				Num: 0,
			},
			executeUpdateSeatNum: true,
			seats:                []*domain.Seat{},
			resSeat:              []*openapi.Seat{},
			resStatus:            http.StatusOK,
		},
		"UpdateSeatNumが負のとき400": {
			req: openapi.PostSeatRequest{
				Num: -1,
			},
			isError:   true,
			resStatus: http.StatusBadRequest,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seatMock := mock.NewMockSeat(ctrl)
			seatHandler := NewSeat(seatMock)

			c, _, rec := setupTestRequest(t, http.MethodPost, "/seats", withJSONBody(t, testCase.req))

			c.SetPath("/seats")

			if testCase.executeUpdateSeatNum {
				seatMock.EXPECT().UpdateSeatNum(gomock.Any(), uint(testCase.req.Num)).
					Return(testCase.seats, testCase.UpdateSeatNumError)
			}

			err := seatHandler.PostSeat(c)
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

func TestPatchSeatStatus(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req                     openapi.PatchSeatStatusRequest
		seatID                  values.SeatID
		executeUpdateSeatStatus bool
		seat                    *domain.Seat
		UpdateSeatStatusError   error
		resSeat                 *openapi.Seat
		isError                 bool
		resStatus               int
	}{
		"正しく変更できる": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.InUse,
			},
			seatID:                  values.SeatID(1),
			executeUpdateSeatStatus: true,
			seat:                    domain.NewSeat(1, values.SeatStatusInUse),
			resSeat: &openapi.Seat{
				Id:     openapi.SeatID(1),
				Status: openapi.InUse,
			},
			resStatus: http.StatusOK,
		},
		"無効な座席ステータスが指定されたら400": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.SeatStatus("invalid"),
			},
			seatID:    values.SeatID(1),
			isError:   true,
			resStatus: http.StatusBadRequest,
		},
		"UpdateSeatStatusがエラーなので500": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.InUse,
			},
			seatID:                  values.SeatID(1),
			executeUpdateSeatStatus: true,
			UpdateSeatStatusError:   assert.AnError,
			isError:                 true,
			resStatus:               http.StatusInternalServerError,
		},
		"UpdateSeatStatusで存在しない座席IDが指定されたら404": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.InUse,
			},
			seatID:                  values.SeatID(999),
			executeUpdateSeatStatus: true,
			UpdateSeatStatusError:   service.ErrNoSeat,
			isError:                 true,
			resStatus:               http.StatusNotFound,
		},
		"UpdateSeatStatusで無効な座席ステータスが指定されたら404": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.InUse,
			},
			seatID:                  values.SeatID(1),
			executeUpdateSeatStatus: true,
			UpdateSeatStatusError:   service.ErrInvalidSeatStatus,
			isError:                 true,
			resStatus:               http.StatusNotFound,
		},
		"UpdateSeatStatusで座席が空席に変更された場合": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.Empty,
			},
			seatID:                  values.SeatID(1),
			executeUpdateSeatStatus: true,
			seat:                    domain.NewSeat(1, values.SeatStatusEmpty),
			resSeat: &openapi.Seat{
				Id:     openapi.SeatID(1),
				Status: openapi.Empty,
			},
			resStatus: http.StatusOK,
		},
		"UpdateSeatStatusの返り値に無効な座席ステータスが含まれていたら500": {
			req: openapi.PatchSeatStatusRequest{
				Status: openapi.InUse,
			},
			seatID:                  values.SeatID(1),
			executeUpdateSeatStatus: true,
			seat:                    domain.NewSeat(1, 100), // 無効な座席ステータス
			isError:                 true,
			resStatus:               http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seatMock := mock.NewMockSeat(ctrl)
			seatHandler := NewSeat(seatMock)

			c, _, rec := setupTestRequest(t, http.MethodPatch, fmt.Sprintf("/seats/%d", testCase.seatID), withJSONBody(t, testCase.req))

			c.SetPath("/seats/:seatID")
			c.SetParamNames("seatID")
			c.SetParamValues(strconv.Itoa(int(testCase.seatID)))

			if testCase.executeUpdateSeatStatus {
				var status values.SeatStatus
				switch testCase.req.Status {
				case openapi.Empty:
					status = values.SeatStatusEmpty
				case openapi.InUse:
					status = values.SeatStatusInUse
				default:
					t.Fatalf("invalid seat status: %v", testCase.req.Status)
				}
				seatMock.EXPECT().UpdateSeatStatus(gomock.Any(), testCase.seatID, status).
					Return(testCase.seat, testCase.UpdateSeatStatusError)
			}

			err := seatHandler.PatchSeatStatus(c, openapi.SeatIDInPath(testCase.seatID))
			if testCase.isError {
				assert.Error(t, err)
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, testCase.resStatus, httpErr.Code)
				}
				return
			}

			assert.Equal(t, testCase.resStatus, rec.Code)
			var res openapi.Seat
			err = json.NewDecoder(rec.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, testCase.resSeat.Id, res.Id)
			assert.Equal(t, testCase.resSeat.Status, res.Status)
		})
	}
}
