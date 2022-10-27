package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

type Seat struct {
	id     values.SeatID
	status values.SeatStatus
}

func NewSeat(id values.SeatID, status values.SeatStatus) *Seat {
	return &Seat{
		id:     id,
		status: status,
	}
}

func (s *Seat) ID() values.SeatID {
	return s.id
}

func (s *Seat) Status() values.SeatStatus {
	return s.status
}

func (s *Seat) SetStatus(status values.SeatStatus) {
	s.status = status
}
