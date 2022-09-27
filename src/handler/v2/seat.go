package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type Seat struct {
	seatUnimplemented
}

func NewSeat() *Seat {
	return &Seat{}
}

// seatUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type seatUnimplemented interface {
	// 座席一覧の取得
	// (GET /seats)
	GetSeats(ctx echo.Context) error
	// 席数の変更
	// (POST /seats)
	PostSeat(ctx echo.Context) error
	// 席の変更
	// (PATCH /seats/{seatID})
	PatchSeatStatus(ctx echo.Context, seatID openapi.SeatIDInPath) error
}
