package values

import (
	"time"

	"github.com/google/uuid"
)

type (
	EventID string
	EventName string
	EventCreatedAt time.Time
	EventDeletedAt nullableTime
)

var (
	NullEventDeletedAt EventDeletedAt = EventDeletedAt(nullTime)
)

func NewEventID() EventID {
	return EventID(uuid.New().String())
}

func NewEventIDFromString(id string) (EventID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return EventID(id), nil
}

func NewEventName(name string) (EventName, error) {
	if len(name) > 32 {
		return "", ErrTooLong
	}

	return EventName(name), nil
}

func NewEventCreatedAt(createdAt time.Time) (EventCreatedAt, error) {
	return EventCreatedAt(createdAt), nil
}

func NewEventDeletedAt(deletedAt time.Time) (EventDeletedAt, error) {
	return EventDeletedAt(nullableTime{
		time: deletedAt,
	}), nil
}
