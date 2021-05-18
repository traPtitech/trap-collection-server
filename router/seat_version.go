package router

import (
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
	seatVersion, err := sv.db.InsertSeatVersion(uint(newSeatVersion.Hight), uint(newSeatVersion.Width))
	if err != nil {
		return nil, fmt.Errorf("failed to insert seat version: %w", err)
	}

	return seatVersion, nil
}
