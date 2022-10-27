package cache

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

type Seat interface {
	// SetActiveSeats
	// アクティブな座席情報一覧をキャッシュする
	SetActiveSeats(ctx context.Context, seats []*domain.Seat) error
	// GetActiveSeats
	// アクティブな座席情報一覧を取得する
	GetActiveSeats(ctx context.Context) ([]*domain.Seat, error)
}
