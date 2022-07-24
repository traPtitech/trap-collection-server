package storage

import (
	"context"
	"io"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImage interface {
	SaveGameImage(ctx context.Context, reader io.Reader, imageID values.GameImageID) error
	GetTempURL(ctx context.Context, image *domain.GameImage, expires time.Duration) (values.GameImageTmpURL, error)
}
