package service

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile interface {
	SaveGameFile(ctx context.Context, reader io.Reader, gameID values.GameID, fileType values.GameFileType, entryPoint values.GameFileEntryPoint) (*domain.GameFile, error)
	GetGameFile(ctx context.Context, gameID values.GameID, environment *values.LauncherEnvironment) (values.GameFileTmpURL, *domain.GameFile, error)
}
