package service

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, reader io.Reader, gameID values.GameID) error
	GetGameImage(ctx context.Context, gameID values.GameID) (io.ReadCloser, error)
}
