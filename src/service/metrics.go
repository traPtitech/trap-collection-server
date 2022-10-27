package service

import (
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type SeatMetrics interface {
	// UpdateWithActiveSeats
	// 全メトリクスを更新する
	UpdateWithActiveSeats(activeSeats []*domain.Seat)
	// UpdateWithNewSeatStatus
	// SeatStatusの変化に基づいてメトリクスを増減させる
	UpdateWithNewSeatStatus(newStatus values.SeatStatus)
}
