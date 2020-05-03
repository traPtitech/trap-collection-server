package router

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Seat seatの構造体
type Seat struct {
	openapi.SeatApi
	LauncherAuthBase
}

// PostSeat POST /seats の処理部分
func (s *Seat)PostSeat(sess sessionMap) (sessionMap, error) {
	productKey, err := s.getProductKey(sess)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Getting ProductKey")
	}
	err = model.PostPlayer(productKey)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Inserting Player: %w", err)
	}
	return sessionMap{}, nil
}

// DeleteSeat DELETE /seats の処理部分
func (s *Seat)DeleteSeat(sess sessionMap) (sessionMap, error) {
	productKey, err := s.getProductKey(sess)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Getting ProductKey")
	}
	err = model.DeletePlayer(productKey)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Inserting Player: %w", err)
	}
	return sessionMap{}, nil
}
