package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Seat interface {
	// CreateSeats
	// 座席を追加する
	CreateSeats(ctx context.Context, seats []*domain.Seat) error
	// UpdateSeatsStatus
	// 座席の状態を更新する
	UpdateSeatsStatus(ctx context.Context, seatIDs []values.SeatID, status values.SeatStatus) error
	// GetActiveSeats
	// 有効な座席の情報を取得する
	GetActiveSeats(ctx context.Context, lockType LockType) ([]*domain.Seat, error)
	// GetSeats
	// 無効な座席を含めた座席の情報を取得する
	GetSeats(ctx context.Context) ([]*domain.Seat, error)
	// GetSeat
	// 座席情報を取得する
	// 無効な座席情報も取得できる
	GetSeat(ctx context.Context, seatID values.SeatID, lockType LockType) (*domain.Seat, error)
}
