package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.GameVersionV2 = &GameVersion{}

type GameVersion struct {
	db repository.DB
	// TODO: v2のgameRepositoryに変更
	gameRepository        repository.Game
	gameImageRepository   repository.GameImageV2
	gameVideoRepository   repository.GameVideoV2
	gameFileRepository    repository.GameFileV2
	gameVersionRepository repository.GameVersionV2
}

func NewGameVersion(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImageV2,
	gameVideoRepository repository.GameVideoV2,
	gameFileRepository repository.GameFileV2,
	gameVersionRepository repository.GameVersionV2,
) *GameVersion {
	return &GameVersion{
		db:                    db,
		gameRepository:        gameRepository,
		gameImageRepository:   gameImageRepository,
		gameVideoRepository:   gameVideoRepository,
		gameFileRepository:    gameFileRepository,
		gameVersionRepository: gameVersionRepository,
	}
}

func (gameVersion *GameVersion) CreateGameVersion(
	ctx context.Context,
	gameID values.GameID,
	name values.GameVersionName,
	description values.GameVersionDescription,
	imageID values.GameImageID,
	videoID values.GameVideoID,
	assets *service.Assets,
) (*service.GameVersionInfo, error) {
	fileIDs := make([]values.GameFileID, 0, 3)
	// fileの種類確認用のmap
	fileTypeMap := make(map[values.GameFileID]values.GameFileType, 3)
	windowsFileID, windowsFileOk := assets.Windows.Value()
	if windowsFileOk {
		fileIDs = append(fileIDs, windowsFileID)
		fileTypeMap[windowsFileID] = values.GameFileTypeWindows
	}

	macFileID, macFileOk := assets.Mac.Value()
	if macFileOk {
		fileIDs = append(fileIDs, macFileID)
		fileTypeMap[macFileID] = values.GameFileTypeMac
	}

	jarFileID, jarFileOk := assets.Jar.Value()
	if jarFileOk {
		fileIDs = append(fileIDs, jarFileID)
		fileTypeMap[jarFileID] = values.GameFileTypeJar
	}

	_, urlOk := assets.URL.Value()
	if !urlOk && !windowsFileOk && !macFileOk && !jarFileOk {
		return nil, service.ErrNoAsset
	}

	version := domain.NewGameVersion(
		values.NewGameVersionID(),
		name,
		description,
		time.Now(),
	)

	err := gameVersion.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameVersion.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return err
		}

		gameImage, err := gameVersion.gameImageRepository.GetGameImage(ctx, imageID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameImageID
		}
		if err != nil {
			return fmt.Errorf("failed to get game image: %w", err)
		}

		if gameImage.GameID != gameID {
			// 権限がない人からgame imageが存在していることがわからないように、
			// gameが存在しない場合と同じエラーを返す
			return service.ErrInvalidGameImageID
		}

		gameVideo, err := gameVersion.gameVideoRepository.GetGameVideo(ctx, videoID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameVideoID
		}
		if err != nil {
			return fmt.Errorf("failed to get game video: %w", err)
		}

		if gameVideo.GameID != gameID {
			// 権限がない人からgame videoが存在していることがわからないように、
			// gameが存在しない場合と同じエラーを返す
			return service.ErrInvalidGameVideoID
		}

		if len(fileIDs) != 0 {
			gameFiles, err := gameVersion.gameFileRepository.GetGameFiles(ctx, fileIDs, repository.LockTypeRecord)
			if err != nil {
				return fmt.Errorf("failed to get game files: %w", err)
			}

			gameFileMap := make(map[values.GameFileID]*domain.GameFile, len(gameFiles))
			for _, gameFile := range gameFiles {
				if gameFile.GameID != gameID {
					// 権限がない人からgame fileが存在していることがわからないように、
					// gameが存在しない場合と同じエラーを返す
					return service.ErrInvalidGameFileID
				}

				gameFileMap[gameFile.GameFile.GetID()] = gameFile.GameFile
			}

			for id, fileType := range fileTypeMap {
				gameFile, ok := gameFileMap[id]
				if !ok {
					return service.ErrInvalidGameFileID
				}

				if gameFile.GetFileType() != fileType {
					return service.ErrInvalidGameFileType
				}
			}
		}

		err = gameVersion.gameVersionRepository.CreateGameVersion(ctx, gameID, imageID, videoID, assets.URL, fileIDs, version)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return &service.GameVersionInfo{
		GameVersion: version,
		Assets:      assets,
		ImageID:     imageID,
		VideoID:     videoID,
	}, nil
}
