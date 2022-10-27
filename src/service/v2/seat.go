package v2

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.Seat = (*Seat)(nil)

type Seat struct {
	db             repository.DB
	seatRepository repository.Seat
	seatCache      cache.Seat
	seatMetrics    *SeatMetrics
}

func NewSeat(
	db repository.DB,
	seatRepository repository.Seat,
	seatCache cache.Seat,
	seatMetrics *SeatMetrics,
) *Seat {
	return &Seat{
		db:             db,
		seatRepository: seatRepository,
		seatCache:      seatCache,
		seatMetrics:    seatMetrics,
	}
}

func (s *Seat) GetSeats(ctx context.Context) ([]*domain.Seat, error) {
	seats, err := s.seatCache.GetActiveSeats(ctx)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// cacheからの取り出しに失敗しても、dbから取り出せば良いのでエラーは無視する
		log.Printf("error: failed to get seats from cache: %v\n", err)
	}
	if err == nil {
		return seats, nil
	}

	seats, err = s.seatRepository.GetActiveSeats(ctx, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	err = s.seatCache.SetActiveSeats(ctx, seats)
	if err != nil {
		// cacheの設定に失敗しても致命傷ではないのでエラーを返さない
		log.Printf("error: failed to set seats to cache: %v\n", err)
	}

	// 基本はSeatStatusのupdate時に更新しているが、定期的に正確な情報への同期を行いたいので
	s.seatMetrics.UpdateWithActiveSeats(seats)

	return seats, nil
}

func (s *Seat) UpdateSeatStatus(ctx context.Context, seatID values.SeatID, status values.SeatStatus) (*domain.Seat, error) {
	if status == values.SeatStatusNone {
		return nil, service.ErrInvalidSeatStatus
	}

	var seat *domain.Seat
	err := s.db.Transaction(ctx, nil, func(ctx context.Context) error {
		var err error
		seat, err = s.seatRepository.GetSeat(ctx, seatID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoSeat
		}
		if err != nil {
			return fmt.Errorf("failed to get seat: %w", err)
		}

		if seat.Status() == values.SeatStatusNone {
			return service.ErrNoSeat
		}

		if seat.Status() == status {
			return nil
		}

		seat.SetStatus(status)

		err = s.seatRepository.UpdateSeatsStatus(ctx, []values.SeatID{seatID}, status)
		if err != nil {
			return fmt.Errorf("failed to update seats status: %w", err)
		}

		s.seatMetrics.UpdateWithNewSeatStatus(status)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update seat status: %w", err)
	}

	return seat, nil
}

func (s *Seat) UpdateSeatNum(ctx context.Context, num uint) ([]*domain.Seat, error) {
	var activeSeats []*domain.Seat
	err := s.db.Transaction(ctx, nil, func(ctx context.Context) error {
		seats, err := s.seatRepository.GetSeats(ctx, repository.LockTypeNone)
		if err != nil {
			return fmt.Errorf("failed to get seats: %w", err)
		}
		seatNum := uint(len(seats))

		seatMap := make(map[values.SeatID]*domain.Seat)
		for _, seat := range seats {
			seatMap[seat.ID()] = seat
		}

		activeSeats = make([]*domain.Seat, 0, num)
		var (
			newSeats          []*domain.Seat
			deactivateSeatIDs []values.SeatID
			activateSeatIDs   []values.SeatID
		)
		for i := uint(1); i <= num; i++ {
			seatID := values.NewSeatID(i)

			seat, ok := seatMap[seatID]
			if ok {
				if seat.Status() == values.SeatStatusNone {
					activateSeatIDs = append(activateSeatIDs, seatID)
					seat.SetStatus(values.SeatStatusEmpty)
				}
			} else {
				newSeat := domain.NewSeat(seatID, values.SeatStatusEmpty)

				newSeats = append(newSeats, newSeat)
			}

			activeSeats = append(activeSeats, seat)
		}

		for i := num + 1; i <= seatNum; i++ {
			seatID := values.NewSeatID(i)

			seat, ok := seatMap[seatID]
			if ok {
				if seat.Status() != values.SeatStatusNone {
					deactivateSeatIDs = append(deactivateSeatIDs, seatID)
					seat.SetStatus(values.SeatStatusNone)
				}
			}
		}

		if len(newSeats) > 0 {
			err = s.seatRepository.CreateSeats(ctx, newSeats)
			if err != nil {
				return fmt.Errorf("failed to create seats: %w", err)
			}
		}

		if len(deactivateSeatIDs) > 0 {
			err = s.seatRepository.UpdateSeatsStatus(ctx, deactivateSeatIDs, values.SeatStatusNone)
			if err != nil {
				return fmt.Errorf("failed to deactivate seats: %w", err)
			}
		}

		if len(activateSeatIDs) > 0 {
			err = s.seatRepository.UpdateSeatsStatus(ctx, activateSeatIDs, values.SeatStatusEmpty)
			if err != nil {
				return fmt.Errorf("failed to activate seats: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update seat num: %w", err)
	}

	err = s.seatCache.SetActiveSeats(ctx, activeSeats)
	if err != nil {
		// cacheの設定に失敗しても致命傷ではないのでエラーを返さない
		log.Printf("error: failed to set seats to cache: %v", err)
	}

	s.seatMetrics.UpdateWithActiveSeats(activeSeats)

	return activeSeats, nil
}
