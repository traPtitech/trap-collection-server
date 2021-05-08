package domain

import (
	"github.com/traPtitech/trap-collection-server/domain/values"
)

type Event struct {
	id values.EventID
	name values.EventName
	createdAt values.EventCreatedAt
	deletedAt values.EventDeletedAt
}

func NewEvent(id values.EventID, name values.EventName, createdAt values.EventCreatedAt, deletedAt values.EventDeletedAt) *Event {
	return &Event{
		id: id,
		name: name,
		createdAt: createdAt,
		deletedAt: deletedAt,
	}
}

func (e *Event) GetID() values.EventID {
	return e.id
}

func (e *Event) GetName() values.EventName {
	return e.name
}

func (e *Event) GetCreatedAt() values.EventCreatedAt {
	return e.createdAt
}

func (e *Event) GetDeletedAt() values.EventDeletedAt {
	return e.deletedAt
}
