package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Game struct {
	id          values.GameID
	name        values.GameName
	description values.GameDescription
	visibility  values.GameVisibility
	createdAt   time.Time
}

func NewGame(
	id values.GameID,
	name values.GameName,
	description values.GameDescription,
	visibility values.GameVisibility,
	createdAt time.Time,
) *Game {
	return &Game{
		id:          id,
		name:        name,
		description: description,
		visibility:  visibility,
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

func (g *Game) GetVisibility() values.GameVisibility {
	return g.visibility
}

func (g *Game) SetVisibility(visibility values.GameVisibility) {
	g.visibility = visibility
}

type GameWithGenres struct {
	game   *Game
	genres []*GameGenre
}

func NewGameWithGenres(game *Game, genres []*GameGenre) *GameWithGenres {
	return &GameWithGenres{
		game:   game,
		genres: genres,
	}
}

func (g *GameWithGenres) GetGame() *Game {
	return g.game
}

func (g *GameWithGenres) GetGenres() []*GameGenre {
	return g.genres
}
