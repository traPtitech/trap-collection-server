package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Game interface {
	GetGame(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.Game, error)
}
