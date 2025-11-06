package gorm2

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type Seat struct {
	db *DB
}

func NewSeat(db *DB) *Seat {
	return &Seat{
		db: db,
	}
}

func (s *Seat) CreateSeats(ctx context.Context, seats []*domain.Seat) error {
	db, err := s.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var status []schema.SeatStatusTable
	err = db.
		Where("active = true").
		Find(&status).Error
	if err != nil {
		return fmt.Errorf("failed to get seat status: %w", err)
	}

	statusMap := make(map[string]uint8, len(status))
	for _, s := range status {
		statusMap[s.Name] = s.ID
	}

	dbSeats := make([]schema.SeatTable, 0, len(seats))
	for _, seat := range seats {
		var (
			status uint8
			ok     bool
		)
		switch seat.Status() {
		case values.SeatStatusNone:
			status, ok = statusMap[schema.SeatStatusNone]
		case values.SeatStatusEmpty:
			status, ok = statusMap[schema.SeatStatusEmpty]
		case values.SeatStatusInUse:
			status, ok = statusMap[schema.SeatStatusInUse]
		default:
			return fmt.Errorf("invalid seat status: %d", seat.Status())
		}
		if !ok {
			return fmt.Errorf("invalid seat status: %d", seat.Status())
		}

		dbSeats = append(dbSeats, schema.SeatTable{
			ID:       uint(seat.ID()),
			StatusID: status,
		})
	}

	err = db.
		Create(&dbSeats).Error
	if err != nil {
		return fmt.Errorf("failed to create seats: %w", err)
	}

	return nil
}

func (s *Seat) UpdateSeatsStatus(ctx context.Context, seatIDs []values.SeatID, status values.SeatStatus) error {
	db, err := s.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var dbStatusName string
	switch status {
	case values.SeatStatusNone:
		dbStatusName = schema.SeatStatusNone
	case values.SeatStatusEmpty:
		dbStatusName = schema.SeatStatusEmpty
	case values.SeatStatusInUse:
		dbStatusName = schema.SeatStatusInUse
	default:
		return fmt.Errorf("invalid seat status: %d", status)
	}

	var dbStatus schema.SeatStatusTable
	err = db.
		Where("active = true").
		Where("name = ?", dbStatusName).
		Take(&dbStatus).Error
	if err != nil {
		return fmt.Errorf("failed to get seat status: %w", err)
	}

	result := db.
		Model(&schema.SeatTable{}).
		Where("id IN ?", seatIDs).
		Update("status_id", dbStatus.ID)
	if result.Error != nil {
		return fmt.Errorf("failed to update seats status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}
func (s *Seat) GetActiveSeats(ctx context.Context, lockType repository.LockType) ([]*domain.Seat, error) {
	db, err := s.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = s.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbSeats []schema.SeatTable
	err = db.
		Joins("SeatStatus").
		Order("seats.id").
		Where("SeatStatus.name != ?", schema.SeatStatusNone).
		Find(&dbSeats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	seats := make([]*domain.Seat, 0, len(dbSeats))
	for _, dbSeat := range dbSeats {
		var status values.SeatStatus
		switch dbSeat.SeatStatus.Name {
		case schema.SeatStatusEmpty:
			status = values.SeatStatusEmpty
		case schema.SeatStatusInUse:
			status = values.SeatStatusInUse
		default:
			log.Printf("invalid product key status: %s\n", dbSeat.SeatStatus.Name)
			continue
		}

		seats = append(seats, domain.NewSeat(
			values.NewSeatID(dbSeat.ID),
			status,
		))
	}

	return seats, nil
}

func (s *Seat) GetSeats(ctx context.Context, lockType repository.LockType) ([]*domain.Seat, error) {
	db, err := s.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = s.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbSeats []schema.SeatTable
	err = db.
		Joins("SeatStatus").
		Order("seats.id").
		Find(&dbSeats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	seats := make([]*domain.Seat, 0, len(dbSeats))
	for _, dbSeat := range dbSeats {
		var status values.SeatStatus
		switch dbSeat.SeatStatus.Name {
		case schema.SeatStatusNone:
			status = values.SeatStatusNone
		case schema.SeatStatusEmpty:
			status = values.SeatStatusEmpty
		case schema.SeatStatusInUse:
			status = values.SeatStatusInUse
		default:
			// 1つ不正な値が格納されるだけで機能停止すると困るので、エラーを返さずにログを出力する
			log.Printf("error: invalid seat status: %s\n", dbSeat.SeatStatus.Name)
			continue
		}

		seats = append(seats, domain.NewSeat(
			values.NewSeatID(dbSeat.ID),
			status,
		))
	}

	return seats, nil
}

func (s *Seat) GetSeat(ctx context.Context, seatID values.SeatID, lockType repository.LockType) (*domain.Seat, error) {
	db, err := s.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = s.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbSeat schema.SeatTable
	err = db.
		Joins("SeatStatus").
		Where("seats.id = ?", uint(seatID)).
		Take(&dbSeat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get seat: %w", err)
	}

	var status values.SeatStatus
	switch dbSeat.SeatStatus.Name {
	case schema.SeatStatusNone:
		status = values.SeatStatusNone
	case schema.SeatStatusEmpty:
		status = values.SeatStatusEmpty
	case schema.SeatStatusInUse:
		status = values.SeatStatusInUse
	default:
		return nil, fmt.Errorf("invalid product key status: %s", dbSeat.SeatStatus.Name)
	}

	return domain.NewSeat(
		values.NewSeatID(dbSeat.ID),
		status,
	), nil
}
