package storage

import (
	"context"
	"io"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile interface {
	SaveGameFile(ctx context.Context, reader io.Reader, fileID values.GameFileID) error
	GetTempURL(ctx context.Context, file *domain.GameFile, expires time.Duration) (values.GameFileTmpURL, error)
}
