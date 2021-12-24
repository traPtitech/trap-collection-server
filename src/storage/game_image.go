package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, reader io.Reader, imageID values.GameImageID) error
	GetGameImage(ctx context.Context, writer io.Writer, image *domain.GameImage) error
}
