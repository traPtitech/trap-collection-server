package router

import (
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
