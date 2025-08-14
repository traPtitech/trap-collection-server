package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Seat struct {
	seatService service.Seat
}

func NewSeat(seatService service.Seat) *Seat {
	return &Seat{
		seatService: seatService,
	}
}

// 座席一覧の取得
// (GET /seats)
func (seat *Seat) GetSeats(c echo.Context) error {
	seats, err := seat.seatService.GetSeats(c.Request().Context())
	if err != nil {
		log.Printf("error: failed to get seats: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get seats")
	}

	res := make([]*openapi.Seat, 0, len(seats))
	for _, seat := range seats {
		var status openapi.SeatStatus
		switch seat.Status() {
		case values.SeatStatusEmpty:
			status = openapi.Empty
		case values.SeatStatusInUse:
			status = openapi.InUse
		default:
			log.Printf("error: invalid seat status: %v\n", seat.Status())
			continue
		}

		res = append(res, &openapi.Seat{
			Id:     openapi.SeatID(seat.ID()),
			Status: status,
		})
	}

	return c.JSON(http.StatusOK, res)
}

// 席数の変更
// (POST /seats)
func (seat *Seat) PostSeat(c echo.Context) error {
	var req openapi.PostSeatRequest
	err := c.Bind(&req)
	if err != nil {
		log.Printf("error: failed to bind request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request")
	}

	if req.Num < 0 || req.Num > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid seat number")
	}

	seats, err := seat.seatService.UpdateSeatNum(c.Request().Context(), uint(req.Num))
	if err != nil {
		log.Printf("error: failed to post seat: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to post seat")
	}

	res := make([]*openapi.Seat, 0, len(seats))
	for _, seat := range seats {
		var status openapi.SeatStatus
		switch seat.Status() {
		case values.SeatStatusEmpty:
			status = openapi.Empty
		case values.SeatStatusInUse:
			status = openapi.InUse
		default:
			log.Printf("error: invalid seat status: %v\n", seat.Status())
			continue
		}

		res = append(res, &openapi.Seat{
			Id:     openapi.SeatID(seat.ID()),
			Status: status,
		})
	}

	return c.JSON(http.StatusOK, res)
}

// 席の変更
// (PATCH /seats/{seatID})
func (seat *Seat) PatchSeatStatus(c echo.Context, seatID openapi.SeatIDInPath) error {
	var req openapi.PatchSeatStatusRequest
	err := c.Bind(&req)
	if err != nil {
		log.Printf("error: failed to bind request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request")
	}

	var status values.SeatStatus
	switch req.Status {
	case openapi.Empty:
		status = values.SeatStatusEmpty
	case openapi.InUse:
		status = values.SeatStatusInUse
	default:
		log.Printf("error: invalid seat status: %v\n", req.Status)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid seat status")
	}

	domainSeat, err := seat.seatService.UpdateSeatStatus(c.Request().Context(), values.SeatID(seatID), status)
	if errors.Is(err, service.ErrNoSeat) || errors.Is(err, service.ErrInvalidSeatStatus) {
		return echo.NewHTTPError(http.StatusNotFound, "no seat")
	}
	if err != nil {
		log.Printf("error: failed to patch seat status: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to patch seat status")
	}

	var resStatus openapi.SeatStatus
	switch domainSeat.Status() {
	case values.SeatStatusEmpty:
		resStatus = openapi.Empty
	case values.SeatStatusInUse:
		resStatus = openapi.InUse
	default:
		log.Printf("error: invalid seat status: %v\n", domainSeat.Status())
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid seat status")
	}

	return c.JSON(http.StatusOK, openapi.Seat{
		Id:     openapi.SeatID(domainSeat.ID()),
		Status: resStatus,
	})
}
