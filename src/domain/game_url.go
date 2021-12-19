package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameURL struct {
	id        values.GameURLID
	link      values.GameURLLink
	createdAt time.Time
}

func NewGameURL(
	id values.GameURLID,
	link values.GameURLLink,
	createdAt time.Time,
) *GameURL {
	return &GameURL{
		id:        id,
		link:      link,
		createdAt: createdAt,
	}
}

func (a *GameURL) GetID() values.GameURLID {
	return a.id
}

func (a *GameURL) GetLink() values.GameURLLink {
	return a.link
}

func (a *GameURL) GetCreatedAt() time.Time {
	return a.createdAt
}
