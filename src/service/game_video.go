package service

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideo interface {
	SaveGameVideo(ctx context.Context, reader io.Reader, gameID values.GameID) error
	GetGameVideo(ctx context.Context, gameID values.GameID) (io.Reader, error)
}
