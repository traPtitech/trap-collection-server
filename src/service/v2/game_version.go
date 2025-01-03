package v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.GameVersionV2 = &GameVersion{}

type GameVersion struct {
	db                    repository.DB
	gameRepository        repository.GameV2
	gameImageRepository   repository.GameImageV2
	gameVideoRepository   repository.GameVideoV2
	gameFileRepository    repository.GameFileV2
	gameVersionRepository repository.GameVersionV2
}

func NewGameVersion(
	db repository.DB,
	gameRepository repository.GameV2,
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

		// 既存のゲームバージョンの名前と一致していた場合はエラーを返す
		_, currentGameVersions, err := gameVersion.gameVersionRepository.GetGameVersions(ctx, gameID, 0, 0, repository.LockTypeNone)
		if err != nil {
			return fmt.Errorf("failed to get game versions: %w", err)
		}
		for _, currentGameVersion := range currentGameVersions {
			if currentGameVersion.GetName() == name {
				return service.ErrDuplicateGameVersion
			}
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
			gameFiles, err := gameVersion.gameFileRepository.GetGameFilesWithoutTypes(ctx, fileIDs, repository.LockTypeRecord)
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

func (gameVersion *GameVersion) GetGameVersions(ctx context.Context, gameID values.GameID, params *service.GetGameVersionsParams) (uint, []*service.GameVersionInfo, error) {
	var (
		limit  uint
		offset uint
	)
	if params == nil {
		limit = 0
		offset = 0
	} else {
		if params.Limit == 0 {
			return 0, nil, service.ErrInvalidLimit
		}

		limit = params.Limit
		offset = params.Offset
	}

	num, gameVersions, err := gameVersion.gameVersionRepository.GetGameVersions(ctx, gameID, limit, offset, repository.LockTypeNone)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	fileIDMap := make(map[values.GameFileID]struct{}, len(gameVersions))
	fileIDs := make([]values.GameFileID, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		for _, fileID := range gameVersion.FileIDs {
			if _, ok := fileIDMap[fileID]; ok {
				continue
			}

			fileIDs = append(fileIDs, fileID)

			fileIDMap[fileID] = struct{}{}
		}
	}

	gameFileMap := make(map[values.GameFileID]*domain.GameFile, len(fileIDs))
	if len(fileIDs) > 0 {
		gameFiles, err := gameVersion.gameFileRepository.GetGameFilesWithoutTypes(ctx, fileIDs, repository.LockTypeNone)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to get game files: %w", err)
		}

		for _, gameFile := range gameFiles {
			gameFileMap[gameFile.GameFile.GetID()] = gameFile.GameFile
		}
	}

	gameVersionInfos := make([]*service.GameVersionInfo, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		assets := &service.Assets{
			URL: gameVersion.URL,
		}
		for _, id := range gameVersion.FileIDs {
			gameFile, ok := gameFileMap[id]
			if !ok {
				log.Printf("error: game file not found(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, gameVersion.GetID(), id)
				continue
			}

			switch gameFile.GetFileType() {
			case values.GameFileTypeWindows:
				if _, ok := assets.Windows.Value(); ok {
					log.Printf("error: duplicate file type windows(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, gameVersion.GetID(), id)
					continue
				}

				assets.Windows = types.NewOption(gameFile.GetID())
			case values.GameFileTypeMac:
				if _, ok := assets.Mac.Value(); ok {
					log.Printf("error: duplicate file type mac(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, gameVersion.GetID(), id)
					continue
				}

				assets.Mac = types.NewOption(gameFile.GetID())
			case values.GameFileTypeJar:
				if _, ok := assets.Jar.Value(); ok {
					log.Printf("error: duplicate file type jar(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, gameVersion.GetID(), id)
					continue
				}

				assets.Jar = types.NewOption(gameFile.GetID())
			default:
				log.Printf("invalid game file type: game_id=%s, game_version_id=%s, game_file_id=%s, file_type=%d\n", gameID, gameVersion.GetID(), id, gameFile.GetFileType())
				continue
			}
		}

		gameVersionInfos = append(gameVersionInfos, &service.GameVersionInfo{
			GameVersion: gameVersion.GameVersion,
			Assets:      assets,
			ImageID:     gameVersion.ImageID,
			VideoID:     gameVersion.VideoID,
		})
	}

	return num, gameVersionInfos, nil
}

func (gameVersion *GameVersion) GetLatestGameVersion(ctx context.Context, gameID values.GameID) (*service.GameVersionInfo, error) {
	_, err := gameVersion.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	version, err := gameVersion.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game version: %w", err)
	}

	gameFileMap := make(map[values.GameFileID]*domain.GameFile, len(version.FileIDs))
	if len(version.FileIDs) != 0 {
		gameFiles, err := gameVersion.gameFileRepository.GetGameFilesWithoutTypes(ctx, version.FileIDs, repository.LockTypeNone)
		if err != nil {
			return nil, fmt.Errorf("failed to get game files: %w", err)
		}

		for _, gameFile := range gameFiles {
			gameFileMap[gameFile.GameFile.GetID()] = gameFile.GameFile
		}
	}

	assets := &service.Assets{
		URL: version.URL,
	}
	for _, id := range version.FileIDs {
		gameFile, ok := gameFileMap[id]
		if !ok {
			log.Printf("error: game file not found(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, version.GetID(), id)
			continue
		}

		switch gameFile.GetFileType() {
		case values.GameFileTypeWindows:
			if _, ok := assets.Windows.Value(); ok {
				log.Printf("error: duplicate file type windows(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, version.GetID(), id)
				continue
			}

			assets.Windows = types.NewOption(gameFile.GetID())
		case values.GameFileTypeMac:
			if _, ok := assets.Mac.Value(); ok {
				log.Printf("error: duplicate file type mac(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, version.GetID(), id)
				continue
			}

			assets.Mac = types.NewOption(gameFile.GetID())
		case values.GameFileTypeJar:
			if _, ok := assets.Jar.Value(); ok {
				log.Printf("error: duplicate file type jar(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameID, version.GetID(), id)
				continue
			}

			assets.Jar = types.NewOption(gameFile.GetID())
		default:
			log.Printf("invalid game file type: game_id=%s, game_version_id=%s, game_file_id=%s, file_type=%d\n", gameID, version.GetID(), id, gameFile.GetFileType())
			continue
		}
	}

	return &service.GameVersionInfo{
		GameVersion: version.GameVersion,
		Assets:      assets,
		ImageID:     version.ImageID,
		VideoID:     version.VideoID,
	}, nil
}
