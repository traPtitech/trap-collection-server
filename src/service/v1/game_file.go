package v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
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

		buf := bytes.NewBuffer(nil)
		tr := io.TeeReader(reader, buf)
		hash, err := values.NewGameFileHash(tr)
		if err != nil {
			return fmt.Errorf("failed to get hash: %w", err)
		}

		_, err = io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("failed to read all: %w", err)
		}

		gameFileID := values.NewGameFileID()
		gameFile = domain.NewGameFile(gameFileID, fileType, entryPoint, hash)

		err = gf.gameFileRepository.SaveGameFile(ctx, gameVersion.GetID(), gameFile)
		if err != nil {
			return fmt.Errorf("failed to save game file(repository): %w", err)
		}

		err = gf.gameFileStorage.SaveGameFile(ctx, buf, gameFile)
		if err != nil {
			return fmt.Errorf("failed to save game file(storage): %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return gameFile, nil
}

func (gf *GameFile) GetGameFile(ctx context.Context, writer io.Writer, gameID values.GameID, environment *values.LauncherEnvironment) (*domain.GameFile, error) {
	_, err := gf.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameVersion, err := gf.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game version: %w", err)
	}

	gameFiles, err := gf.gameFileRepository.GetGameFiles(ctx, gameVersion.GetID(), environment.AcceptGameFileTypes())
	if err != nil {
		return nil, fmt.Errorf("failed to get game file: %w", err)
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
		return nil, service.ErrNoGameFile
	}

	err = gf.gameFileStorage.GetGameFile(ctx, writer, gameFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get game file(storage): %w", err)
	}

	return gameFile, nil
}
