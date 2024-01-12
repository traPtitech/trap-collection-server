package v2

import (
	"context"
	"errors"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameGenre struct {
	gameGenreRepository repository.GameGenre
}

func NewGameGenre(gameGenreRepository repository.GameGenre) *GameGenre {
	return &GameGenre{
		gameGenreRepository: gameGenreRepository,
	}
}

var _ service.GameGenre = &GameGenre{}

func (gameGenre *GameGenre) GetGameGenres(ctx context.Context) ([]*service.GameGenreInfo, error) {
	gameInfosRepo, err := gameGenre.gameGenreRepository.GetGameGenres(ctx)
	if err != nil {
		return nil, err
	}

	gameInfos := make([]*service.GameGenreInfo, 0, len(gameInfosRepo))
	for i := range gameInfosRepo {
		gameInfos = append(gameInfos, &service.GameGenreInfo{
			GameGenre: gameInfosRepo[i].GameGenre,
			Num:       gameInfosRepo[i].Num,
		})
	}
	return gameInfos, nil
}

func (gameGenre *GameGenre) DeleteGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error {
	err := gameGenre.gameGenreRepository.RemoveGameGenre(ctx, gameGenreID)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return service.ErrNoGameGenre
	}
	if err != nil {
		return err
	}

	return nil
}
