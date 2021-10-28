package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Game struct {
	id          values.GameID
	name        values.GameName
	description values.GameDescription
	createdAt   time.Time
}

func NewGame(
	id values.GameID,
	name values.GameName,
	description values.GameDescription,
	createdAt time.Time,
) *Game {
	return &Game{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
	}
}

func (g *Game) GetID() values.GameID {
	return g.id
}

func (g *Game) GetName() values.GameName {
	return g.name
}

func (g *Game) SetName(name values.GameName) {
	g.name = name
}

func (g *Game) GetDescription() values.GameDescription {
	return g.description
}

func (g *Game) SetDescription(description values.GameDescription) {
	g.description = description
}

func (g *Game) GetCreatedAt() time.Time {
	return g.createdAt
}
