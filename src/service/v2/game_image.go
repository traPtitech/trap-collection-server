package v2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
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

var _ service.GameImageV2 = &GameImage{}

type GameImage struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameImageRepository repository.GameImageV2
	gameImageStorage    storage.GameImage
}

func NewGameImage(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImageV2,
	gameImageStorage storage.GameImage,
) *GameImage {
	return &GameImage{
		db:                  db,
		gameRepository:      gameRepository,
		gameImageRepository: gameImageRepository,
		gameImageStorage:    gameImageStorage,
	}
}

func (gameImage *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, gameID values.GameID) (*domain.GameImage, error) {
	var image *domain.GameImage
	err := gameImage.db.Transaction(ctx, nil, func(ctx context.Context) error {
		// TODO: v2のgameRepositoryに変更
		_, err := gameImage.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
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

			image = domain.NewGameImage(
				imageID,
				imageType,
				time.Now(),
			)

			err = gameImage.gameImageRepository.SaveGameImage(ctx, gameID, image)
			if err != nil {
				return fmt.Errorf("failed to save game image: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gameImage.gameImageStorage.SaveGameImage(ctx, filePr, imageID)
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
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return image, nil
}

func (gameImage *GameImage) GetGameImages(ctx context.Context, gameID values.GameID) ([]*domain.GameImage, error) {
	_, err := gameImage.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	images, err := gameImage.gameImageRepository.GetGameImages(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game images: %w", err)
	}

	return images, nil
}

func (gameImage *GameImage) GetGameImage(ctx context.Context, gameID values.GameID, imageID values.GameImageID) (values.GameImageTmpURL, error) {
	_, err := gameImage.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	var url *url.URL
	err = gameImage.db.Transaction(ctx, nil, func(ctx context.Context) error {
		image, err := gameImage.gameImageRepository.GetGameImage(ctx, imageID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameImageID
		}
		if err != nil {
			return fmt.Errorf("failed to get game image file: %w", err)
		}

		if image.GameID != gameID {
			// gameIdに対応したゲームにゲーム画像が紐づいていない場合も、
			// 念の為閲覧権限がないゲームに紐づいた画像IDを知ることができないようにするため、
			// 画像が存在しない場合と同じErrInvalidGameImageIDを返す
			return service.ErrInvalidGameImageID
		}

		url, err = gameImage.gameImageStorage.GetTempURL(ctx, image.GameImage, time.Minute)
		if err != nil {
			return fmt.Errorf("failed to get game image: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return url, nil
}

func (gameImage *GameImage) GetGameImageMeta(ctx context.Context, gameID values.GameID, imageID values.GameImageID) (*domain.GameImage, error) {
	_, err := gameImage.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	image, err := gameImage.gameImageRepository.GetGameImage(ctx, imageID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameImageID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game image file: %w", err)
	}

	if image.GameID != gameID {
		// gameIdに対応したゲームにゲーム画像が紐づいていない場合も、
		// 念の為閲覧権限がないゲームに紐づいた画像IDを知ることができないようにするため、
		// 画像が存在しない場合と同じErrInvalidGameImageIDを返す
		return nil, service.ErrInvalidGameImageID
	}

	return image.GameImage, nil
}
