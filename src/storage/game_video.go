package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideo interface {
	SaveGameVideo(ctx context.Context, reader io.Reader, videoID values.GameVideoID) error
	GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error
}
