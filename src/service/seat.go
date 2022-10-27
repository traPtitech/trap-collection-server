package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Seat interface {
	// UpdateSeatNum
	// 座席数を変更する。
	// 既に存在する座席の状態は保持する。
	UpdateSeatNum(ctx context.Context, num uint) ([]*domain.Seat, error)
	// UpdateSeatStatus
	// 座席の状態を変更する
	// 座席が存在しない場合はErrNoSeatを返す。
	// 無効な状態を指定した場合はErrInvalidSeatStatusを返す。
	UpdateSeatStatus(ctx context.Context, seatID values.SeatID, status values.SeatStatus) (*domain.Seat, error)
	// GetSeats
	// 座席情報を取得する
	GetSeats(ctx context.Context) ([]*domain.Seat, error)
}
