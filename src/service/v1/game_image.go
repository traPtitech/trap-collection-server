package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"golang.org/x/sync/errgroup"
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

		imageID := values.NewGameImageID()

		eg, ctx := errgroup.WithContext(ctx)
		fileTypePr, fileTypePw := io.Pipe()
		filePr, filePw := io.Pipe()

		eg.Go(func() error {
			defer fileTypePr.Close()

			fType, err := filetype.MatchReader(fileTypePr)
			if err != nil {
				return fmt.Errorf("failed to get file type: %w", err)
			}

			_, err = io.ReadAll(fileTypePr)
			if err != nil {
				return fmt.Errorf("failed to read file type: %w", err)
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
				imageID,
				imageType,
				time.Now(),
			)

			err = gi.gameImageRepository.SaveGameImage(ctx, gameID, image)
			if err != nil {
				return fmt.Errorf("failed to save game image: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gi.gameImageStorage.SaveGameImage(ctx, filePr, imageID)
			if err != nil {
				return fmt.Errorf("failed to save game image file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePw.Close()
			defer fileTypePw.Close()

			mw := io.MultiWriter(fileTypePw, filePw)
			_, err = io.Copy(mw, reader)
			if err != nil {
				return fmt.Errorf("failed to copy image: %w", err)
			}

			return nil
		})

		err = eg.Wait()
		if err != nil {
			return fmt.Errorf("failed to save game image: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (gi *GameImage) GetGameImage(ctx context.Context, gameID values.GameID) (values.GameImageTmpURL, error) {
	_, err := gi.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	image, err := gi.gameImageRepository.GetLatestGameImage(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameImage
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game image: %w", err)
	}

	tmpURL, err := gi.gameImageStorage.GetTmpURL(ctx, image, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to get game image temp url: %W", err)
	}

	return tmpURL, nil
}
