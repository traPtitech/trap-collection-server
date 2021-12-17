package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameURL interface {
	SaveGameURL(ctx context.Context, gameID values.GameID, link values.GameURLLink) (*domain.GameURL, error)
	GetGameURL(ctx context.Context, gameID values.GameID) (*domain.GameURL, error)
}
