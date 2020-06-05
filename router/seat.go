package router

import (
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Seat seatの構造体
type Seat struct {
	openapi.SeatApi
	base.LauncherAuthBase
}

// PostSeat POST /seats の処理部分
func (s *Seat)PostSeat(c echo.Context) error {
	productKey, err := s.GetProductKey(c)
	if err != nil {
		return fmt.Errorf("Failed In Getting ProductKey")
	}
	err = model.PostPlayer(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Inserting Player: %w", err)
	}
	return nil
}

// DeleteSeat DELETE /seats の処理部分
func (s *Seat)DeleteSeat(c echo.Context) error {
	productKey, err := s.GetProductKey(c)
	if err != nil {
		return fmt.Errorf("Failed In Getting ProductKey")
	}
	err = model.DeletePlayer(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Inserting Player: %w", err)
	}
	return nil
}
