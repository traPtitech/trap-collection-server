package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Game interface {
	CreateGame(ctx context.Context, name values.GameName, description values.GameDescription) (*domain.Game, error)
	UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription) (*domain.Game, error)
	GetGame(ctx context.Context, id values.GameID) (*GameInfo, error)
	GetGames(ctx context.Context) ([]*GameInfo, error)
	GetMyGames(ctx context.Context, session *domain.OIDCSession) ([]*GameInfo, error)
	DeleteGame(ctx context.Context, id values.GameID) error
}

type GameInfo struct {
	*domain.Game
	// nullableなことに注意!
	LatestVersion *domain.GameVersion
}
