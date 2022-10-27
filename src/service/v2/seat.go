package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.Seat = (*Seat)(nil)

type Seat struct {
	db             repository.DB
	seatRepository repository.Seat
}

func NewSeat(db repository.DB, seatRepository repository.Seat) *Seat {
	return &Seat{
		db:             db,
		seatRepository: seatRepository,
	}
}

func (s *Seat) GetSeats(ctx context.Context) ([]*domain.Seat, error) {
	seats, err := s.seatRepository.GetActiveSeats(ctx, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

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

		err = s.seatRepository.UpdateSeatsStatus(ctx, []values.SeatID{seatID}, status)
		if err != nil {
			return fmt.Errorf("failed to update seats status: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update seat status: %w", err)
	}

	return seat, nil
}

func (s *Seat) UpdateSeatNum(ctx context.Context, num uint) ([]*domain.Seat, error) {
	seats, err := s.seatRepository.GetSeats(ctx, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}
	seatNum := uint(len(seats))

	switch {
	case seatNum < num:
		newSeats := make([]*domain.Seat, 0, num-seatNum)
		for i := seatNum + 1; i <= num; i++ {
			newSeats = append(newSeats, domain.NewSeat(
				values.NewSeatID(i),
				values.SeatStatusEmpty,
			))
		}

		err := s.seatRepository.CreateSeats(ctx, newSeats)
		if err != nil {
			return nil, fmt.Errorf("failed to create seats: %w", err)
		}

		seats = append(seats, newSeats...)
	case seatNum > num:
		seatIDs := make([]values.SeatID, 0, seatNum-num)
		for i := num + 1; i <= seatNum; i++ {
			seatIDs = append(seatIDs, values.NewSeatID(i))
		}

		err := s.seatRepository.UpdateSeatsStatus(ctx, seatIDs, values.SeatStatusNone)
		if err != nil {
			return nil, fmt.Errorf("failed to update seats: %w", err)
		}

		seats = seats[:num]
	default:
		return seats, nil
	}

	return seats, nil
}
