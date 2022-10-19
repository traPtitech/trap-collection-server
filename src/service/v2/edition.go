package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/pkg/types"
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
}

func NewEdition(
	db repository.DB,
	editionRepository repository.Edition,
	gameRepository repository.GameV2,
	gameVersionRepository repository.GameVersionV2,
) *Edition {
	return &Edition{
		db:                    db,
		editionRepository:     editionRepository,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
	}
}

func (edition *Edition) CreateEdition(
	ctx context.Context,
	name values.LauncherVersionName,
	questionnaireURL types.Option[values.LauncherVersionQuestionnaireURL],
	gameVersionIDs []values.GameVersionID,
) (*domain.LauncherVersion, error) {
	var newEdition *domain.LauncherVersion
	if url, ok := questionnaireURL.Value(); ok {
		newEdition = domain.NewLauncherVersionWithQuestionnaire(values.NewLauncherVersionID(), name, url, time.Now())
	} else {
		newEdition = domain.NewLauncherVersionWithoutQuestionnaire(values.NewLauncherVersionID(), name, time.Now())
	}

	err := edition.db.Transaction(ctx, nil, func(ctx context.Context) error {
		gameVersions, err := edition.gameVersionRepository.GetGameVersionsByIDs(ctx, gameVersionIDs, repository.LockTypeRecord)
		if err != nil {
			return fmt.Errorf("failed to get game versions: %w", err)
		}

		if len(gameVersions) != len(gameVersionIDs) {
			return service.ErrInvalidGameVersionID
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
