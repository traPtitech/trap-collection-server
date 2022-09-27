package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile interface {
	SaveGameFile(ctx context.Context, gameVersionID values.GameVersionID, gameFile *domain.GameFile) error
	GetGameFiles(ctx context.Context, gameVersionID values.GameVersionID, fileTypes []values.GameFileType) ([]*domain.GameFile, error)
}
