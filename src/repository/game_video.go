package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideo interface {
	SaveGameVideo(ctx context.Context, gameID values.GameID, video *domain.GameVideo) error
	GetLatestGameVideo(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.GameVideo, error)
}
