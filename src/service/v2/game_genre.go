package v2

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameGenre struct {
	db                  repository.DB
	gameGenreRepository repository.GameGenre
}

func NewGameGenre(db repository.DB, gameGenreRepository repository.GameGenre) *GameGenre {
	return &GameGenre{
		db:                  db,
		gameGenreRepository: gameGenreRepository,
	}
}

var _ service.GameGenre = &GameGenre{}

func (gameGenre *GameGenre) GetGameGenres(ctx context.Context, isLoginUser bool) ([]*service.GameGenreInfo, error) {
	var visibilities []values.GameVisibility
	if !isLoginUser {
		visibilities = []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited}
	} else {
		visibilities = []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate}

	}

	gameInfosRepo, err := gameGenre.gameGenreRepository.GetGameGenres(ctx, visibilities)
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

func (gameGenre *GameGenre) UpdateGameGenres(ctx context.Context, gameID values.GameID, gameGenreNames []values.GameGenreName) error {
	// 重複するジャンルがあったらエラー
	if len(slices.Compact[[]values.GameGenreName](gameGenreNames)) != len(gameGenreNames) {
		return service.ErrDuplicateGameGenre
	}

	err := gameGenre.db.Transaction(ctx, nil, func(ctx context.Context) error {
		existingGenres, err := gameGenre.gameGenreRepository.GetGameGenresWithNames(ctx, gameGenreNames)
		if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
			return fmt.Errorf("failed to get game genres: %w", err)
		}

		newGameGenres := make([]*domain.GameGenre, 0, len(gameGenreNames))
		if len(existingGenres) != len(gameGenreNames) {
			existingGenresMap := make(map[values.GameGenreName]struct{}, len(existingGenres))
			for i := range existingGenres {
				existingGenresMap[existingGenres[i].GetName()] = struct{}{}
			}

			for i := range gameGenreNames {
				if _, ok := existingGenresMap[gameGenreNames[i]]; !ok {
					newGameGenres = append(newGameGenres, domain.NewGameGenre(values.NewGameGenreID(), gameGenreNames[i], time.Now()))
				}
			}

			err = gameGenre.gameGenreRepository.SaveGameGenres(ctx, newGameGenres)
			if errors.Is(err, repository.ErrDuplicatedUniqueKey) {
				return service.ErrDuplicateGameGenre
			}
			if err != nil {
				return fmt.Errorf("failed to save game genres: %w", err)
			}
		}

		gameGenreIDs := make([]values.GameGenreID, 0, len(existingGenres)+len(newGameGenres))
		for i := range existingGenres {
			gameGenreIDs = append(gameGenreIDs, existingGenres[i].GetID())
		}
		for i := range newGameGenres {
			gameGenreIDs = append(gameGenreIDs, newGameGenres[i].GetID())
		}

		err = gameGenre.gameGenreRepository.RegisterGenresToGame(ctx, gameID, gameGenreIDs)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGame
		}
		if err != nil {
			return fmt.Errorf("failed to register genres to game: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
