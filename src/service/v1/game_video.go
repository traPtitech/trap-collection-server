package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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

type GameVideo struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameVideoRepository repository.GameVideo
	gameVideoStorage    storage.GameVideo
}

func NewGameVideo(
	db repository.DB,
	gameRepository repository.Game,
	gameVideoRepository repository.GameVideo,
	gameVideoStorage storage.GameVideo,
) *GameVideo {
	return &GameVideo{
		db:                  db,
		gameRepository:      gameRepository,
		gameVideoRepository: gameVideoRepository,
		gameVideoStorage:    gameVideoStorage,
	}
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, gameID values.GameID) error {
	err := gv.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gv.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		videoID := values.NewGameVideoID()

		eg, ctx := errgroup.WithContext(ctx)
		fileTypePr, fileTypePw := io.Pipe()
		filePr, filePw := io.Pipe()

		eg.Go(func() error {
			defer fileTypePr.Close()

			fType, err := filetype.MatchReader(fileTypePr)
			if err != nil {
				return fmt.Errorf("failed to get file type: %w", err)
			}

			var videoType values.GameVideoType
			switch fType.Extension {
			case matchers.TypeMp4.Extension:
				videoType = values.GameVideoTypeMp4
			default:
				return service.ErrInvalidFormat
			}

			_, err = io.ReadAll(fileTypePr)
			if err != nil {
				return fmt.Errorf("failed to read video: %w", err)
			}

			video := domain.NewGameVideo(
				videoID,
				videoType,
				time.Now(),
			)

			err = gv.gameVideoRepository.SaveGameVideo(ctx, gameID, video)
			if err != nil {
				return fmt.Errorf("failed to save game video: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gv.gameVideoStorage.SaveGameVideo(ctx, filePr, videoID)
			if err != nil {
				return fmt.Errorf("failed to save game video file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePw.Close()
			defer fileTypePw.Close()

			mw := io.MultiWriter(fileTypePw, filePw)
			_, err = io.Copy(mw, reader)
			if err != nil {
				return fmt.Errorf("failed to copy video: %w", err)
			}

			return nil
		})

		err = eg.Wait()
		if err != nil {
			return fmt.Errorf("failed to save game video: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (gv *GameVideo) GetGameVideo(ctx context.Context, gameID values.GameID) (io.ReadCloser, error) {
	_, err := gv.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	video, err := gv.gameVideoRepository.GetLatestGameVideo(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameVideo
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game image: %w", err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		err = gv.gameVideoStorage.GetGameVideo(ctx, pw, video)
		if err != nil {
			log.Printf("error: failed to get game video: %+v", err)
		}
	}()

	return pr, nil
}
