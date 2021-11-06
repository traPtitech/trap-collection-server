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

		buf := bytes.NewBuffer(nil)
		r := io.TeeReader(reader, buf)
		fType, err := filetype.MatchReader(r)
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

		_, err = io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read video: %w", err)
		}

		video := domain.NewGameVideo(
			values.NewGameVideoID(),
			videoType,
		)

		err = gv.gameVideoRepository.SaveGameVideo(ctx, gameID, video)
		if err != nil {
			return fmt.Errorf("failed to save game video: %w", err)
		}

		err = gv.gameVideoStorage.SaveGameVideo(ctx, buf, video)
		if err != nil {
			return fmt.Errorf("failed to save game video file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (gv *GameVideo) GetGameVideo(ctx context.Context, writer io.Writer, gameID values.GameID) error {
	err := gv.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gv.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		video, err := gv.gameVideoRepository.GetLatestGameVideo(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGameImage
		}
		if err != nil {
			return fmt.Errorf("failed to get game image: %w", err)
		}

		err = gv.gameVideoStorage.GetGameVideo(ctx, writer, video)
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
