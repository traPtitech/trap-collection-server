package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
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
