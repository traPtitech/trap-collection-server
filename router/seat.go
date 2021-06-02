package router

import (
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Seat seatの構造体
type Seat struct {
	db           model.DBMeta
	launcherAuth base.LauncherAuth
	openapi.SeatApi
}

func newSeat(db model.DBMeta, launcherAuth base.LauncherAuth) openapi.SeatApi {
	seat := new(Seat)

	seat.db = db
	seat.launcherAuth = launcherAuth

	return seat
}

// PostSeat POST /seats の処理部分
func (s *Seat) PostSeat(seatReq *openapi.Seat) (*openapi.SeatDetail, error) {
	seatVersion, err := s.db.GetSeatVersion(seatReq.SeatVersionId)
	if errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("invalid seat version id")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check seat version id: %w", err)
	}

	row := seatReq.SeatId / int32(seatVersion.Height)
	column := seatReq.SeatId % int32(seatVersion.Width)

	seat, err := s.db.InsertSeat(seatReq.SeatVersionId, int(row), int(column))
	if errors.Is(err, model.ErrAlreadyExists) {
		return nil, errors.New("already seated")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert seat: %w", err)
	}

	return &openapi.SeatDetail{
		Id: seatReq.SeatId,
		Status: 1,
		SeatingTime: seat.StartedAt,
	}, nil
}

// DeleteSeat DELETE /seats の処理部分
func (s *Seat) DeleteSeat(seatReq *openapi.Seat) (*openapi.SeatDetail, error) {
	seatVersion, err := s.db.GetSeatVersion(seatReq.SeatVersionId)
	if errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("invalid seat version id")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check seat version id: %w", err)
	}

	row := seatReq.SeatId / int32(seatVersion.Height)
	column := seatReq.SeatId % int32(seatVersion.Width)

	err = s.db.DeleteSeat(seatReq.SeatVersionId, int(row), int(column))
	if errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("not seated")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete seat: %w", err)
	}

	return &openapi.SeatDetail{
		Id: seatReq.SeatId,
		Status: 0,
	}, nil
}
