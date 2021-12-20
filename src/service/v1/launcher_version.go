package v1

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

type LauncherVersion struct {
	db                        repository.DB
	launcherVersionRepository repository.LauncherVersion
	gameRepository            repository.Game
}

func NewLauncherVersion(
	db repository.DB,
	launcherVersionRepository repository.LauncherVersion,
	gameRepository repository.Game,
) *LauncherVersion {
	return &LauncherVersion{
		db:                        db,
		launcherVersionRepository: launcherVersionRepository,
		gameRepository:            gameRepository,
	}
}

func (lv *LauncherVersion) CreateLauncherVersion(ctx context.Context, name values.LauncherVersionName, questionnaireURL values.LauncherVersionQuestionnaireURL) (*domain.LauncherVersion, error) {
	var launcherVersion *domain.LauncherVersion
	if questionnaireURL == nil {
		launcherVersion = domain.NewLauncherVersionWithoutQuestionnaire(
			values.NewLauncherVersionID(),
			name,
			time.Now(),
		)
	} else {
		launcherVersion = domain.NewLauncherVersionWithQuestionnaire(
			values.NewLauncherVersionID(),
			name,
			questionnaireURL,
			time.Now(),
		)
	}

	err := lv.launcherVersionRepository.CreateLauncherVersion(ctx, launcherVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher version: %w", err)
	}

	return launcherVersion, nil
}

func (lv *LauncherVersion) GetLauncherVersions(ctx context.Context) ([]*domain.LauncherVersion, error) {
	launcherVersions, err := lv.launcherVersionRepository.GetLauncherVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher versions: %w", err)
	}

	return launcherVersions, nil
}

func (lv *LauncherVersion) GetLauncherVersion(ctx context.Context, id values.LauncherVersionID) (*domain.LauncherVersion, []*domain.Game, error) {
	launcherVersion, err := lv.launcherVersionRepository.GetLauncherVersion(ctx, id, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrNoLauncherVersion
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	games, err := lv.gameRepository.GetGamesByLauncherVersion(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get games: %w", err)
	}

	return launcherVersion, games, nil
}

func (lv *LauncherVersion) AddGamesToLauncherVersion(ctx context.Context, id values.LauncherVersionID, gameIDs []values.GameID) (*domain.LauncherVersion, []*domain.Game, error) {
	var launcherVersion *domain.LauncherVersion
	var games []*domain.Game
	err := lv.db.Transaction(ctx, nil, func(ctx context.Context) error {
		var err error
		launcherVersion, err = lv.launcherVersionRepository.GetLauncherVersion(ctx, id, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoLauncherVersion
		}
		if err != nil {
			return fmt.Errorf("failed to get launcher version: %w", err)
		}

		games, err = lv.gameRepository.GetGamesByLauncherVersion(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get games: %w", err)
		}

		gameMap := make(map[values.GameID]struct{}, len(games))
		for _, game := range games {
			gameMap[game.GetID()] = struct{}{}
		}

		// 既にあるゲーム、新たに追加するゲームとの重複チェック
		for _, gameID := range gameIDs {
			if _, ok := gameMap[gameID]; ok {
				return service.ErrDuplicateGame
			}

			gameMap[gameID] = struct{}{}
		}

		newGames, err := lv.gameRepository.GetGamesByIDs(ctx, gameIDs, repository.LockTypeRecord)
		if err != nil {
			return fmt.Errorf("failed to get games: %w", err)
		}

		if len(newGames) != len(gameIDs) {
			return service.ErrNoGame
		}

		err = lv.launcherVersionRepository.AddGamesToLauncherVersion(ctx, id, gameIDs)
		if err != nil {
			return fmt.Errorf("failed to add games to launcher version: %w", err)
		}

		games = append(games, newGames...)

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return launcherVersion, games, nil
}

func (lv *LauncherVersion) GetLauncherVersionCheckList(ctx context.Context, launcherVersionID values.LauncherVersionID, env *values.LauncherEnvironment) ([]*service.CheckListItem, error) {
	_, err := lv.launcherVersionRepository.GetLauncherVersion(ctx, launcherVersionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoLauncherVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	gameInfos, err := lv.gameRepository.GetGameInfosByLauncherVersion(ctx, launcherVersionID, env.AcceptGameFileTypes())
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	checkList := make([]*service.CheckListItem, 0, len(gameInfos))
	for _, gameInfo := range gameInfos {
		// url,fileのうち、fileが存在する場合はurlを無視、存在しない場合はurlを使う
		var url *domain.GameURL
		var file *domain.GameFile
		switch {
		case len(gameInfo.LatestFiles) != 0:
			for _, latestFile := range gameInfo.LatestFiles {
				// Windows,Mac専用のファイルを優先
				if latestFile.GetFileType() == values.GameFileTypeWindows || latestFile.GetFileType() == values.GameFileTypeMac {
					file = latestFile
					break
				}

				// Windows,Mac優先のため、その他の場合はbreakしない
				file = latestFile
			}
		case gameInfo.LatestURL != nil:
			url = gameInfo.LatestURL
		}

		checkList = append(checkList, &service.CheckListItem{
			Game:          gameInfo.Game,
			LatestVersion: gameInfo.LatestVersion,
			LatestURL:     url,
			LatestFile:    file,
			LatestImage:   gameInfo.LatestImage,
			LatestVideo:   gameInfo.LatestVideo,
		})
	}

	return checkList, nil
}
