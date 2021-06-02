package router

import (
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// SeatVersion 席のバージョン管理の構造体
type SeatVersion struct {
	db           model.DBMeta
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

func (sv *SeatVersion) GetSeats(seatVersionID string) ([]*openapi.SeatDetail, error) {
	isValidSeatVersionID, err := sv.db.CheckSeatVersion(seatVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check seat version id: %w", err)
	}
	if !isValidSeatVersionID {
		return nil, errors.New("invalid seat version id")
	}

	seatDetails, err := sv.db.GetSeatDetails(seatVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check seats: %w", err)
	}

	return seatDetails, nil
}
