package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre struct {
	id        values.GameGenreID
	name      values.GameGenreName
	createdAt time.Time
}

func NewGameGenre(
	id values.GameGenreID,
	name values.GameGenreName,
	createdAt time.Time,
) *GameGenre {
	return &GameGenre{
		id:        id,
		name:      name,
		createdAt: createdAt,
	}
}

func (gg *GameGenre) GetID() values.GameGenreID {
	return gg.id
}

func (gg *GameGenre) GetName() values.GameGenreName {
	return gg.name
}

func (gg *GameGenre) GetCreatedAt() time.Time {
	return gg.createdAt
}
