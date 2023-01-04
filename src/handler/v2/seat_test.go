package v2

import (
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSeat(t *testing.T) {
}

func TestSeat_GetSeats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	mockSeatService := mock.NewMockSeat(ctrl)
	seatHandler := NewSeat(mockSeatService)

	type test struct {
		description string
		isErr       bool
		statusCode  int
		seats       []*openapi.Seat
	}

	testCases := []test{
		{
			description: "正常に席の情報を取得できる",
			isErr:       false,
			statusCode:  http.StatusOK,
			seats:       []*openapi.Seat{{Id: 1, Status: "in-use"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/seats", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetAllActiveUser {
				mockUserService.
					EXPECT().
					GetAllActiveUser(gomock.Any(), gomock.Any()).
					Return(testCase.userInfos, testCase.GetAllActiveUserErr)
			}

			err := seatHandler.GetSeats(c)
			if err != nil {
				t.Errorf("failed to get seats: %v", err)
			}

			//if testCase.isErr {
			//	var httpError *echo.HTTPError
			//	if errors.As(err, &httpError) {
			//		assert.Equal(t, testCase.statusCode, httpError.Code)
			//	} else {
			//		t.Errorf("error is not *echo.HTTPError")
			//	}
			//	return
			//}

			var resSeats []*openapi.User
			err = json.NewDecoder(rec.Body).Decode(&resSeats)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.seats, resSeats)
			for i, seat := range resSeats {
				assert.Equal(t, seat.Id, resSeats[i].Id)
				assert.Equal(t, seat.Name, resSeats[i].Name)
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
