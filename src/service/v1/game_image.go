package v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameImage struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameImageRepository repository.GameImage
	gameImageStorage    storage.GameImage
}

func NewGameImage(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImage,
	gameImageStorage storage.GameImage,
) *GameImage {
	return &GameImage{
		db:                  db,
		gameRepository:      gameRepository,
		gameImageRepository: gameImageRepository,
		gameImageStorage:    gameImageStorage,
	}
}

func (gi *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, gameID values.GameID) error {
	err := gi.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gi.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		buf := bytes.NewBuffer(nil)
		r := io.TeeReader(reader, buf)
		fType, err := filetype.MatchReader(r)
		if err != nil {
			return fmt.Errorf("failed to get file type: %w", err)
		}

		var imageType values.GameImageType
		switch fType.Extension {
		case matchers.TypeJpeg.Extension:
			imageType = values.GameImageTypeJpeg
		case matchers.TypePng.Extension:
			imageType = values.GameImageTypePng
		case matchers.TypeGif.Extension:
			imageType = values.GameImageTypeGif
		default:
			return service.ErrInvalidFormat
		}

		image := domain.NewGameImage(
			values.NewGameImageID(),
			imageType,
		)

		err = gi.gameImageRepository.SaveGameImage(ctx, gameID, image)
		if err != nil {
			return fmt.Errorf("failed to save game image: %w", err)
		}

		err = gi.gameImageStorage.SaveGameImage(ctx, buf, image)
		if err != nil {
			return fmt.Errorf("failed to save game image file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (gi *GameImage) GetGameImage(ctx context.Context, writer io.Writer, gameID values.GameID) error {
	err := gi.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gi.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		image, err := gi.gameImageRepository.GetGameImage(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGameImage
		}
		if err != nil {
			return fmt.Errorf("failed to get game image: %w", err)
		}

		err = gi.gameImageStorage.GetGameImage(ctx, writer, image)
		if err != nil {
			return fmt.Errorf("failed to get game image file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}
