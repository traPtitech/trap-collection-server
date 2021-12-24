package v1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"golang.org/x/sync/errgroup"
)

type GameFile struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	gameFileRepository    repository.GameFile
	gameFileStorage       storage.GameFile
}

func NewGameFile(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	gameFileRepository repository.GameFile,
	gameFileStorage storage.GameFile,
) *GameFile {
	return &GameFile{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		gameFileRepository:    gameFileRepository,
		gameFileStorage:       gameFileStorage,
	}
}

func (gf *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, gameID values.GameID, fileType values.GameFileType, entryPoint values.GameFileEntryPoint) (*domain.GameFile, error) {
	var gameFile *domain.GameFile
	err := gf.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gf.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		gameVersion, err := gf.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGameVersion
		}
		if err != nil {
			return fmt.Errorf("failed to get latest game version: %w", err)
		}

		gameFiles, err := gf.gameFileRepository.GetGameFiles(ctx, gameVersion.GetID(), []values.GameFileType{fileType})
		if err != nil {
			return fmt.Errorf("failed to get game file: %w", err)
		}

		if len(gameFiles) != 0 {
			return service.ErrGameFileAlreadyExists
		}

		gameFileID := values.NewGameFileID()

		eg, ctx := errgroup.WithContext(ctx)
		hashPr, hashPw := io.Pipe()
		filePr, filePw := io.Pipe()

		eg.Go(func() error {
			defer hashPr.Close()

			hash, err := values.NewGameFileHash(hashPr)
			if err != nil {
				return fmt.Errorf("failed to get hash: %w", err)
			}

			gameFile = domain.NewGameFile(gameFileID, fileType, entryPoint, hash, time.Now())

			err = gf.gameFileRepository.SaveGameFile(ctx, gameVersion.GetID(), gameFile)
			if err != nil {
				return fmt.Errorf("failed to save game file(repository): %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer filePr.Close()

			err = gf.gameFileStorage.SaveGameFile(ctx, filePr, gameFileID)
			if err != nil {
				return fmt.Errorf("failed to save game file(storage): %w", err)
			}

			return nil
		})

		eg.Go(func() error {
			defer hashPw.Close()
			defer filePw.Close()

			mw := io.MultiWriter(hashPw, filePw)
			_, err := io.Copy(mw, reader)
			if err != nil {
				return fmt.Errorf("failed to copy: %w", err)
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

	return gameFile, nil
}

func (gf *GameFile) GetGameFile(ctx context.Context, gameID values.GameID, environment *values.LauncherEnvironment) (io.Reader, *domain.GameFile, error) {
	_, err := gf.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameVersion, err := gf.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrNoGameVersion
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get latest game version: %w", err)
	}

	gameFiles, err := gf.gameFileRepository.GetGameFiles(ctx, gameVersion.GetID(), environment.AcceptGameFileTypes())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game file: %w", err)
	}

	gameFileTypeMap := make(map[values.GameFileType]*domain.GameFile)
	for _, gameFile := range gameFiles {
		gameFileTypeMap[gameFile.GetFileType()] = gameFile
	}

	var gameFile *domain.GameFile
	if winGameFile, ok := gameFileTypeMap[values.GameFileTypeWindows]; ok {
		gameFile = winGameFile
	} else if macGameFile, ok := gameFileTypeMap[values.GameFileTypeMac]; ok {
		gameFile = macGameFile
	} else if jarGameFile, ok := gameFileTypeMap[values.GameFileTypeJar]; ok {
		gameFile = jarGameFile
	} else {
		return nil, nil, service.ErrNoGameFile
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		err = gf.gameFileStorage.GetGameFile(ctx, pw, gameFile)
		if err != nil {
			log.Printf("error: failed to get game file: %v\n", err)
		}
	}()

	return pr, gameFile, nil
}
