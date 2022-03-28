package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameURL interface {
	SaveGameURL(ctx context.Context, gameVersionID values.GameVersionID, url *domain.GameURL) error
	GetGameURL(ctx context.Context, gameVersionID values.GameVersionID) (*domain.GameURL, error)
}
