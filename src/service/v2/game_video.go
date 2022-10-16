package v2

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

var _ service.GameVideoV2 = &GameVideo{}

type GameVideo struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameVideoRepository repository.GameVideoV2
	gameVideoStorage    storage.GameVideo
}

func NewGameVideo(
	db repository.DB,
	gameRepository repository.Game,
	gameVideoRepository repository.GameVideoV2,
	gameVideoStorage storage.GameVideo,
) *GameVideo {
	return &GameVideo{
		db:                  db,
		gameRepository:      gameRepository,
		gameVideoRepository: gameVideoRepository,
		gameVideoStorage:    gameVideoStorage,
	}
}

func (gameVideo *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, gameID values.GameID) error

func (gameVideo *GameVideo) GetGameVideos(ctx context.Context, gameID values.GameID) ([]*domain.GameVideo, error)

func (gameVideo *GameVideo) GetGameVideo(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (values.GameVideoTmpURL, error)

func (gameVideo *GameVideo) GetGameVideoMeta(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (*domain.GameVideo, error)
