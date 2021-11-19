package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameFile interface {
	SaveGameFile(ctx context.Context, reader io.Reader, file *domain.GameFile) error
	GetGameFile(ctx context.Context, writer io.Writer, file *domain.GameFile) error
}
