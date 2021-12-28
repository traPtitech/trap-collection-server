package storage

import (
	"context"
	"io"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideo interface {
	SaveGameVideo(ctx context.Context, reader io.Reader, videoID values.GameVideoID) error
	GetTempURL(ctx context.Context, video *domain.GameVideo, expires time.Duration) (values.GameVideoTmpURL, error)
}
