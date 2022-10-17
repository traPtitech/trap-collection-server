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

var _ service.GameVideoV2 = &GameVideo{}

type GameVideo struct {
	db                  repository.DB
	gameRepository      repository.GameV2
	gameVideoRepository repository.GameVideoV2
	gameVideoStorage    storage.GameVideo
}

func NewGameVideo(
	db repository.DB,
	gameRepository repository.GameV2,
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

func (gameVideo *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, gameID values.GameID) (*domain.GameVideo, error) {
	var video *domain.GameVideo
	err := gameVideo.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameVideo.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
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

			_, err = io.ReadAll(fileTypePr)
			if err != nil {
				return fmt.Errorf("failed to read file type: %w", err)
			}

			var videoType values.GameVideoType
			if fType.Extension == matchers.TypeMp4.Extension {
				videoType = values.GameVideoTypeMp4
			} else {
				return service.ErrInvalidFormat
			}

			video = domain.NewGameVideo(
				videoID,
				videoType,
				time.Now(),
			)

			err = gameVideo.gameVideoRepository.SaveGameVideo(ctx, gameID, video)
			if err != nil {
				return fmt.Errorf("failed to save game video: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gameVideo.gameVideoStorage.SaveGameVideo(ctx, filePr, videoID)
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
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return video, nil
}

func (gameVideo *GameVideo) GetGameVideos(ctx context.Context, gameID values.GameID) ([]*domain.GameVideo, error) {
	_, err := gameVideo.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	videos, err := gameVideo.gameVideoRepository.GetGameVideos(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game videos: %w", err)
	}

	return videos, nil
}

func (gameVideo *GameVideo) GetGameVideo(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (values.GameVideoTmpURL, error) {
	_, err := gameVideo.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	var url *url.URL
	err = gameVideo.db.Transaction(ctx, nil, func(ctx context.Context) error {
		video, err := gameVideo.gameVideoRepository.GetGameVideo(ctx, videoID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameVideoID
		}
		if err != nil {
			return fmt.Errorf("failed to get game video file: %w", err)
		}

		if video.GameID != gameID {
			// gameIdに対応したゲームにゲーム動画が紐づいていない場合も、
			// 念の為閲覧権限がないゲームに紐づいた動画IDを知ることができないようにするため、
			// 動画が存在しない場合と同じErrInvalidGameVideoIDを返す
			return service.ErrInvalidGameVideoID
		}

		url, err = gameVideo.gameVideoStorage.GetTempURL(ctx, video.GameVideo, time.Minute)
		if err != nil {
			return fmt.Errorf("failed to get game video: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return url, nil
}

func (gameVideo *GameVideo) GetGameVideoMeta(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (*domain.GameVideo, error) {
	return nil, nil
}
