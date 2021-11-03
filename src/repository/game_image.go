package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, gameID values.GameID, image *domain.GameImage) error
	GetGameImage(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.GameImage, error)
}
