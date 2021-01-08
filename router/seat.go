package router

import (
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
/*func (s *Seat) PostSeat(c echo.Context) error {
	productKey, err := s.launcherAuth.GetProductKey(c)
	if err != nil {
		return fmt.Errorf("Failed In Getting ProductKey: %w", err)
	}

	err = s.db.PostPlayer(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Inserting Player: %w", err)
	}

	return nil
}*/

// DeleteSeat DELETE /seats の処理部分
/*func (s *Seat) DeleteSeat(c echo.Context) error {
	productKey, err := s.launcherAuth.GetProductKey(c)
	if err != nil {
		return fmt.Errorf("Failed In Getting ProductKey")
	}

	err = s.db.DeletePlayer(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Deleting Player: %w", err)
	}

	return nil
}*/
