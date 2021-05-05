package domain

import (
	"github.com/traPtitech/trap-collection-server/domain/values"
)

type Game struct {
	id values.GameID
	name values.GameName
	description values.GameDescription
	createdAt values.GameCreatedAt
}

func NewGame(id values.GameID, name values.GameName, description values.GameDescription, createdAt values.GameCreatedAt) *Game {
	return &Game{
		id: id,
		name: name,
		description: description,
		createdAt: createdAt,
	}
}

func (g *Game) GetGame() values.GameID {
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

func (g *Game) GetCreatedAt() values.GameCreatedAt {
	return g.createdAt
}
