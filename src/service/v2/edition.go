package v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/traPtitech/trap-collection-server/pkg/option"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Edition struct {
	db                    repository.DB
	editionRepository     repository.Edition
	gameRepository        repository.GameV2
	gameVersionRepository repository.GameVersionV2
	gameFileRepository    repository.GameFileV2
}

func NewEdition(
	db repository.DB,
	editionRepository repository.Edition,
	gameRepository repository.GameV2,
	gameVersionRepository repository.GameVersionV2,
	gameFileRepository repository.GameFileV2,
) *Edition {
	return &Edition{
		db:                    db,
		editionRepository:     editionRepository,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		gameFileRepository:    gameFileRepository,
	}
}

func (edition *Edition) CreateEdition(
	ctx context.Context,
	name values.EditionName,
	questionnaireURL option.Option[values.EditionQuestionnaireURL],
	gameVersionIDs []values.GameVersionID,
) (*domain.Edition, error) {
	gameVersionMap := make(map[values.GameVersionID]struct{}, len(gameVersionIDs))
	for _, gameVersionID := range gameVersionIDs {
		if _, ok := gameVersionMap[gameVersionID]; ok {
			return nil, service.ErrDuplicateGameVersion
		}

		gameVersionMap[gameVersionID] = struct{}{}
	}

	var newEdition *domain.Edition
	if url, ok := questionnaireURL.Value(); ok {
		newEdition = domain.NewEditionWithQuestionnaire(values.NewEditionID(), name, url, time.Now())
	} else {
		newEdition = domain.NewEditionWithoutQuestionnaire(values.NewEditionID(), name, time.Now())
	}

	err := edition.db.Transaction(ctx, nil, func(ctx context.Context) error {
		gameVersions, err := edition.gameVersionRepository.GetGameVersionsByIDs(ctx, gameVersionIDs, repository.LockTypeRecord)
		if err != nil {
			return fmt.Errorf("failed to get game versions: %w", err)
		}

		if len(gameVersions) != len(gameVersionIDs) {
			return service.ErrInvalidGameVersionID
		}

		gameVersionMap := make(map[values.GameID]struct{}, len(gameVersions))
		for _, gameVersion := range gameVersions {
			if _, ok := gameVersionMap[gameVersion.GameID]; ok {
				return service.ErrDuplicateGame
			}

			gameVersionMap[gameVersion.GameID] = struct{}{}
		}

		err = edition.editionRepository.SaveEdition(ctx, newEdition)
		if err != nil {
			return fmt.Errorf("failed to save edition: %w", err)
		}

		err = edition.editionRepository.UpdateEditionGameVersions(ctx, newEdition.GetID(), gameVersionIDs)
		if err != nil {
			return fmt.Errorf("failed to update edition game versions: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return newEdition, nil
}

func (edition *Edition) GetEditions(ctx context.Context) ([]*domain.Edition, error) {
	editions, err := edition.editionRepository.GetEditions(ctx, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get editions: %w", err)
	}

	return editions, nil
}

func (edition *Edition) GetEdition(ctx context.Context, editionID values.EditionID) (*domain.Edition, error) {
	editionValue, err := edition.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidEditionID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	return editionValue, nil
}

func (edition *Edition) UpdateEdition(
	ctx context.Context,
	editionID values.EditionID,
	name values.EditionName,
	questionnaireURL option.Option[values.EditionQuestionnaireURL],
) (*domain.Edition, error) {
	var editionValue *domain.Edition
	err := edition.db.Transaction(ctx, nil, func(ctx context.Context) error {
		var err error
		editionValue, err = edition.editionRepository.GetEdition(ctx, editionID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidEditionID
		}
		if err != nil {
			return fmt.Errorf("failed to get edition: %w", err)
		}

		editionValue.SetName(name)

		if url, ok := questionnaireURL.Value(); ok {
			editionValue.SetQuestionnaireURL(url)
		} else {
			editionValue.UnsetQuestionnaireURL()
		}

		err = edition.editionRepository.UpdateEdition(ctx, editionValue)
		if err != nil {
			return fmt.Errorf("failed to save edition: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return editionValue, nil
}

func (edition *Edition) DeleteEdition(ctx context.Context, editionID values.EditionID) error {
	err := edition.editionRepository.DeleteEdition(ctx, editionID)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return service.ErrInvalidEditionID
	}
	if err != nil {
		return fmt.Errorf("failed to delete edition: %w", err)
	}

	return nil
}

func (edition *Edition) UpdateEditionGameVersions(
	ctx context.Context,
	editionID values.EditionID,
	gameVersionIDs []values.GameVersionID,
) ([]*service.GameVersionWithGame, error) {
	gameVersionMap := make(map[values.GameVersionID]struct{}, len(gameVersionIDs))
	for _, gameVersionID := range gameVersionIDs {
		if _, ok := gameVersionMap[gameVersionID]; ok {
			return nil, service.ErrDuplicateGameVersion
		}

		gameVersionMap[gameVersionID] = struct{}{}
	}

	var gameVersions []*service.GameVersionWithGame
	err := edition.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := edition.editionRepository.GetEdition(ctx, editionID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidEditionID
		}
		if err != nil {
			return fmt.Errorf("failed to get edition: %w", err)
		}

		gameVersionInfos, err := edition.gameVersionRepository.GetGameVersionsByIDs(ctx, gameVersionIDs, repository.LockTypeRecord)
		if err != nil {
			return fmt.Errorf("failed to get game versions: %w", err)
		}

		if len(gameVersionInfos) != len(gameVersionIDs) {
			return service.ErrInvalidGameVersionID
		}

		gameIDs := make([]values.GameID, 0, len(gameVersionInfos))
		// ゲームが違うゲームバージョンのみなので、重複はない
		fileIDs := []values.GameFileID{}
		gameVersionMap := make(map[values.GameID]struct{}, len(gameVersionInfos))
		for _, gameVersion := range gameVersionInfos {
			if _, ok := gameVersionMap[gameVersion.GameID]; ok {
				return service.ErrDuplicateGame
			}

			gameVersionMap[gameVersion.GameID] = struct{}{}
			gameIDs = append(gameIDs, gameVersion.GameID)
			fileIDs = append(fileIDs, gameVersion.FileIDs...)
		}

		games, err := edition.gameRepository.GetGamesByIDs(ctx, gameIDs, repository.LockTypeNone)
		if err != nil {
			return fmt.Errorf("failed to get games: %w", err)
		}

		gameMap := make(map[values.GameID]*domain.Game, len(games))
		for _, game := range games {
			gameMap[game.GetID()] = game
		}

		files, err := edition.gameFileRepository.GetGameFilesWithoutTypes(ctx, fileIDs, repository.LockTypeNone)
		if err != nil {
			return fmt.Errorf("failed to get game files: %w", err)
		}

		fileMap := make(map[values.GameFileID]*domain.GameFile, len(files))
		for _, file := range files {
			fileMap[file.GetID()] = file.GameFile
		}

		gameVersions = make([]*service.GameVersionWithGame, 0, len(gameVersionInfos))
		for _, gameVersion := range gameVersionInfos {
			game, ok := gameMap[gameVersion.GameID]
			if !ok {
				return errors.New("game not found")
			}

			assets := &service.Assets{
				URL: gameVersion.URL,
			}
			for _, id := range gameVersion.FileIDs {
				file, ok := fileMap[id]
				if !ok {
					log.Printf("error: game file not found(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
					continue
				}

				switch file.GetFileType() {
				case values.GameFileTypeWindows:
					if _, ok := assets.Windows.Value(); ok {
						log.Printf("error: duplicate file type windows(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
						continue
					}

					assets.Windows = option.NewOption(file.GetID())
				case values.GameFileTypeMac:
					if _, ok := assets.Mac.Value(); ok {
						log.Printf("error: duplicate file type mac(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
						continue
					}

					assets.Mac = option.NewOption(file.GetID())
				case values.GameFileTypeJar:
					if _, ok := assets.Jar.Value(); ok {
						log.Printf("error: duplicate file type jar(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
						continue
					}

					assets.Jar = option.NewOption(file.GetID())
				default:
					log.Printf("error: invalid game file type: game_id=%s, game_version_id=%s, game_file_id=%s, file_type=%d\n", gameVersion.GameID, gameVersion.GetID(), id, file.GetFileType())
					continue
				}
			}

			gameVersions = append(gameVersions, &service.GameVersionWithGame{
				GameVersion: service.GameVersionInfo{
					GameVersion: gameVersion.GameVersion,
					ImageID:     gameVersion.ImageID,
					VideoID:     gameVersion.VideoID,
					Assets:      assets,
				},
				Game: game,
			})
		}

		err = edition.editionRepository.UpdateEditionGameVersions(ctx, editionID, gameVersionIDs)
		if err != nil {
			return fmt.Errorf("failed to update edition game versions: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return gameVersions, nil
}

func (edition *Edition) GetEditionGameVersions(ctx context.Context, editionID values.EditionID) ([]*service.GameVersionWithGame, error) {
	_, err := edition.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidEditionID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	gameVersions, err := edition.editionRepository.GetEditionGameVersions(ctx, editionID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game versions: %w", err)
	}

	fileIDs := []values.GameFileID{}
	gameIDs := make([]values.GameID, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		gameIDs = append(gameIDs, gameVersion.GameID)
		fileIDs = append(fileIDs, gameVersion.FileIDs...)
	}

	games, err := edition.gameRepository.GetGamesByIDs(ctx, gameIDs, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	gameMap := make(map[values.GameID]*domain.Game, len(games))
	for _, game := range games {
		gameMap[game.GetID()] = game
	}

	files, err := edition.gameFileRepository.GetGameFilesWithoutTypes(ctx, fileIDs, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game files: %w", err)
	}

	fileMap := make(map[values.GameFileID]*domain.GameFile, len(files))
	for _, file := range files {
		fileMap[file.GetID()] = file.GameFile
	}

	gameVersionInfos := make([]*service.GameVersionWithGame, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		game, ok := gameMap[gameVersion.GameID]
		if !ok {
			return nil, errors.New("game not found")
		}

		assets := &service.Assets{
			URL: gameVersion.URL,
		}
		for _, id := range gameVersion.FileIDs {
			file, ok := fileMap[id]
			if !ok {
				log.Printf("error: game file not found(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
				continue
			}

			switch file.GetFileType() {
			case values.GameFileTypeWindows:
				if _, ok := assets.Windows.Value(); ok {
					log.Printf("error: duplicate file type windows(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
					continue
				}

				assets.Windows = option.NewOption(file.GetID())
			case values.GameFileTypeMac:
				if _, ok := assets.Mac.Value(); ok {
					log.Printf("error: duplicate file type mac(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
					continue
				}

				assets.Mac = option.NewOption(file.GetID())
			case values.GameFileTypeJar:
				if _, ok := assets.Jar.Value(); ok {
					log.Printf("error: duplicate file type jar(game_id=%s, game_version_id=%s, game_file_id=%s)\n", gameVersion.GameID, gameVersion.GetID(), id)
					continue
				}

				assets.Jar = option.NewOption(file.GetID())
			default:
				log.Printf("error: invalid game file type: game_id=%s, game_version_id=%s, game_file_id=%s, file_type=%d\n", gameVersion.GameID, gameVersion.GetID(), id, file.GetFileType())
				continue
			}
		}

		gameVersionInfos = append(gameVersionInfos, &service.GameVersionWithGame{
			GameVersion: service.GameVersionInfo{
				GameVersion: gameVersion.GameVersion,
				ImageID:     gameVersion.ImageID,
				VideoID:     gameVersion.VideoID,
				Assets:      assets,
			},
			Game: game,
		})
	}

	return gameVersionInfos, nil
}
