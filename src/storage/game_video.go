package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameVideo interface {
	SaveGameVideo(ctx context.Context, reader io.Reader, video *domain.GameVideo) error
	GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error
}
