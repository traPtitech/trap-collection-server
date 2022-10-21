package v2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"golang.org/x/sync/errgroup"
)

type GameFile struct {
	db                 repository.DB
	gameRepository     repository.GameV2
	gameFileRepository repository.GameFileV2
	gameFileStorage    storage.GameFile
}

func NewGameFile(
	db repository.DB,
	gameRepository repository.GameV2,
	gameFileRepository repository.GameFileV2,
	gameFileStorage storage.GameFile,
) *GameFile {
	return &GameFile{
		db:                 db,
		gameRepository:     gameRepository,
		gameFileRepository: gameFileRepository,
		gameFileStorage:    gameFileStorage,
	}
}

func (gameFile *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, gameID values.GameID, fileType values.GameFileType, entryPoint values.GameFileEntryPoint) (*domain.GameFile, error) {
	var file *domain.GameFile
	err := gameFile.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		fileID := values.NewGameFileID()

		eg, ctx := errgroup.WithContext(ctx)
		hashPr, hashPw := io.Pipe()
		filePr, filePw := io.Pipe()

		eg.Go(func() error {
			defer hashPr.Close()

			hash, err := values.NewGameFileHash(hashPr)
			if err != nil {
				return fmt.Errorf("failed to get hash: %w", err)
			}

			file = domain.NewGameFile(
				fileID,
				fileType,
				entryPoint,
				hash,
				time.Now(),
			)

			err = gameFile.gameFileRepository.SaveGameFile(ctx, gameID, file)
			if err != nil {
				return fmt.Errorf("failed to save game file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gameFile.gameFileStorage.SaveGameFile(ctx, filePr, fileID)
			if err != nil {
				return fmt.Errorf("failed to save game file: %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer hashPw.Close()
			defer filePw.Close()

			mw := io.MultiWriter(hashPw, filePw)
			_, err = io.Copy(mw, reader)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}

			return nil
		})

		err = eg.Wait()
		if err != nil {
			return fmt.Errorf("failed to save game file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return file, nil
}

func (gameFile *GameFile) GetGameFile(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (values.GameFileTmpURL, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	var url *url.URL
	err = gameFile.db.Transaction(ctx, nil, func(ctx context.Context) error {
		file, err := gameFile.gameFileRepository.GetGameFile(ctx, fileID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameFileID
		}
		if err != nil {
			return fmt.Errorf("failed to get game file: %w", err)
		}

		if file.GameID != gameID {
			// gameIdに対応したゲームにゲームファイルが紐づいていない場合も、
			// 念の為閲覧権限がないゲームに紐づいたファイルIDを知ることができないようにするため、
			// ファイルが存在しない場合と同じErrInvalidGameFileIDを返す
			return service.ErrInvalidGameFileID
		}

		url, err = gameFile.gameFileStorage.GetTempURL(ctx, file.GameFile, time.Minute)
		if err != nil {
			return fmt.Errorf("failed to get game file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return url, nil
}

func (gameFile *GameFile) GetGameFiles(ctx context.Context, gameID values.GameID, environment *values.LauncherEnvironment) ([]*domain.GameFile, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameFiles, err := gameFile.gameFileRepository.GetGameFiles(ctx, gameID, repository.LockTypeNone, environment.AcceptGameFileTypes())
	if err != nil {
		return nil, fmt.Errorf("failed to get game files: %w", err)
	}

	return gameFiles, nil
}

func (gameFile *GameFile) GetGameFileMeta(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (*domain.GameFile, error) {
	_, err := gameFile.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	file, err := gameFile.gameFileRepository.GetGameFile(ctx, fileID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameFileID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game file: %w", err)
	}

	if file.GameID != gameID {
		// gameIdに対応したゲームにゲームファイルが紐づいていない場合も、
		// 念の為閲覧権限がないゲームに紐づいたファイルIDを知ることができないようにするため、
		// ファイルが存在しない場合と同じErrInvalidGameFileIDを返す
		return nil, service.ErrInvalidGameFileID
	}

	return file.GameFile, nil
}
