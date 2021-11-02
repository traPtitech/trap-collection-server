package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, reader io.Reader, image *domain.GameImage) error
	GetGameImage(ctx context.Context, writer io.Writer, image *domain.GameImage) error
}
