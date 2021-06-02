package router

import (
	"errors"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// SeatVersion 席のバージョン管理の構造体
type SeatVersion struct {
	db model.DBMeta
	openapi.SeatVersionApi
}

func newSeatVersion(db model.DBMeta) *SeatVersion {
	return &SeatVersion{
		db: db,
	}
}

func (sv *SeatVersion) PostSeatVersion(newSeatVersion *openapi.NewSeatVersion) (*openapi.SeatVersion, error) {
	if newSeatVersion.Hight <= 0 {
		return nil, errors.New("invalid height")
	}
	if newSeatVersion.Width <= 0 {
		return nil, errors.New("invalid width")
	}

	seatVersion, err := sv.db.InsertSeatVersion(uint(newSeatVersion.Hight), uint(newSeatVersion.Width))
	if err != nil {
		return nil, fmt.Errorf("failed to insert seat version: %w", err)
	}

	return seatVersion, nil
}

func (sv *SeatVersion) DeleteSeatVersion(seatVersionID string) error {
	err := sv.db.DeleteSeatVersion(seatVersionID)
	if errors.Is(err, model.ErrNotFound) {
		return errors.New("invalid seat version id")
	}
	if err != nil {
		return fmt.Errorf("failed to check seat version id: %w", err)
	}

	return nil
}

func (sv *SeatVersion) GetSeats(seatVersionID string) ([]*openapi.SeatDetail, error) {
	seatVersion, err := sv.db.GetSeatVersion(seatVersionID)
	if errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("invalid seat version id")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check seat version id: %w", err)
	}

	seats, err := sv.db.GetSeats(seatVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check seats: %w", err)
	}

	seatingTimes := make([][]*time.Time, 0, seatVersion.Height)
	var height uint
	for height = 0; height < seatVersion.Height; height++ {
		seatingTimes = append(seatingTimes, make([]*time.Time, seatVersion.Width))
	}

	for _, seat := range seats {
		seatingTimes[seat.Row][seat.Column] = &seat.StartedAt
	}

	var width uint
	seatDetails := make([]*openapi.SeatDetail, 0, seatVersion.Width*seatVersion.Height)
	for height = 0; height < seatVersion.Height; height++ {
		for width = 0; width < seatVersion.Width; width++ {
			var status int32
			var seatingTime time.Time
			if seatingTimes[height][width] != nil {
				status = 1
				seatingTime = *seatingTimes[height][width]
			} else {
				status = 0
			}

			seatDetails = append(seatDetails, &openapi.SeatDetail{
				Id:     int32(seatVersion.Width)*int32(height) + int32(width),
				Status: status,
				//TODO:誰も座っていない時に0001-01-01T00:00:00Zになってしまう
				SeatingTime: seatingTime,
			})
		}
	}

	return seatDetails, nil
}
