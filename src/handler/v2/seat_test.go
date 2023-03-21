package v2

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSeat_GetSeats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSeatService := mock.NewMockSeat(ctrl)
	seatHandler := NewSeat(mockSeatService)

	type test struct {
		description     string
		seats           []*domain.Seat
		res             []*openapi.Seat
		executeGetSeats bool
		getSeatsErr     error
		isErr           bool
		err             error
		statusCode      int
	}

	testCases := []test{
		{
			description:     "正常に席の情報を取得できる",
			executeGetSeats: true,
			statusCode:      http.StatusOK,
			seats:           []*domain.Seat{domain.NewSeat(1, 2)},
			res:             []*openapi.Seat{{Id: 1, Status: "in-use"}},
		},
		{
			description:     "getSeatsがエラーなので500",
			executeGetSeats: true,
			getSeatsErr:     errors.New("error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/seats", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetSeats {
				mockSeatService.
					EXPECT().
					GetSeats(gomock.Any()).
					Return(testCase.seats, testCase.getSeatsErr)
			}

			err := seatHandler.GetSeats(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			var resSeats []*openapi.Seat
			err = json.NewDecoder(rec.Body).Decode(&resSeats)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.res, resSeats)
			for i, seat := range resSeats {
				assert.Equal(t, seat.Id, resSeats[i].Id)
				assert.Equal(t, seat.Status, resSeats[i].Status)
			}
		})
	}
}

func TestSeat_PostSeat(t *testing.T) {
	type fields struct {
		seatService service.Seat
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seat := &Seat{
				seatService: tt.fields.seatService,
			}
			if err := seat.PostSeat(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Seat.PostSeat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSeat_PatchSeatStatus(t *testing.T) {
	type fields struct {
		seatService service.Seat
	}
	type args struct {
		c      echo.Context
		seatID openapi.SeatIDInPath
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seat := &Seat{
				seatService: tt.fields.seatService,
			}
			if err := seat.PatchSeatStatus(tt.args.c, tt.args.seatID); (err != nil) != tt.wantErr {
				t.Errorf("Seat.PatchSeatStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
