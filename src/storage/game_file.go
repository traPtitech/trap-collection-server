package storage

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile interface {
	SaveGameFile(ctx context.Context, reader io.Reader, fileID values.GameFileID) error
	GetGameFile(ctx context.Context, writer io.Writer, file *domain.GameFile) error
}
