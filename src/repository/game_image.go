package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, gameID values.GameID, image *domain.GameImage) error
	GetLatestGameImage(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.GameImage, error)
}
