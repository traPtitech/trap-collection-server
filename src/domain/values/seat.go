package values

type (
	SeatID     uint
	SeatStatus uint8
)

func NewSeatID(id uint) SeatID {
	return SeatID(id)
}

const (
	// SeatStatusNone 存在しない座席
	SeatStatusNone SeatStatus = iota
	// SeatStatusEmpty 空席
	SeatStatusEmpty
	// SeatStatusInUse 利用中
	SeatStatusInUse
)
