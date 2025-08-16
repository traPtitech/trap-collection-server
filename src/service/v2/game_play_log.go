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

type GamePlayLog struct {
	db                    repository.DB
	gamePlayLogRepository repository.GamePlayLogV2
	editionRepository     repository.Edition
	gameRepository        repository.GameV2
	gameVersionRepository repository.GameVersionV2
}

func NewGamePlayLog(
	db repository.DB,
	gamePlayLogRepository repository.GamePlayLogV2,
	editionRepository repository.Edition,
	gameRepository repository.GameV2,
	gameVersionRepository repository.GameVersionV2,
) *GamePlayLog {
	return &GamePlayLog{
		db:                    db,
		gamePlayLogRepository: gamePlayLogRepository,
		editionRepository:     editionRepository,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
	}
}

func (g *GamePlayLog) CreatePlayLog(ctx context.Context, editionID values.LauncherVersionID, gameID values.GameID, gameVersionID values.GameVersionID, startTime time.Time) (*domain.GamePlayLog, error) {
	_, err := g.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, service.ErrInvalidEdition
		}
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	_, err = g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, service.ErrInvalidGame
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameVersions, err := g.gameVersionRepository.GetGameVersionsByIDs(ctx, []values.GameVersionID{gameVersionID}, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game version: %w", err)
	}
	if len(gameVersions) != 1 {
		return nil, service.ErrInvalidGameVersion
	}

	playLog := domain.NewGamePlayLog(editionID, gameID, gameVersionID, startTime)

	err = g.gamePlayLogRepository.CreateGamePlayLog(ctx, playLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create game play log: %w", err)
	}

	return playLog, nil
}

func (g *GamePlayLog) UpdatePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error {
	playLog, err := g.gamePlayLogRepository.GetGamePlayLog(ctx, playLogID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidPlayLogID
		}
		return fmt.Errorf("failed to get game play log: %w", err)
	}

	if endTime.Before(playLog.StartTime) {
		return service.ErrInvalidEndTime
	}

	err = g.gamePlayLogRepository.UpdateGamePlayLogEndTime(ctx, playLogID, endTime)
	if err != nil {
		return fmt.Errorf("failed to update game play log end time: %w", err)
	}

	return nil
}

func (g *GamePlayLog) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	_, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, service.ErrInvalidGame
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	if gameVersionID != nil {
		gameVersions, err := g.gameVersionRepository.GetGameVersionsByIDs(ctx, []values.GameVersionID{*gameVersionID}, repository.LockTypeNone)
		if err != nil {
			return nil, fmt.Errorf("failed to get game version: %w", err)
		}
		if len(gameVersions) != 1 {
			return nil, service.ErrInvalidGameVersion
		}
	}

	stats, err := g.gamePlayLogRepository.GetGamePlayStats(ctx, gameID, gameVersionID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get game play stats: %w", err)
	}

	return stats, nil
}

func (g *GamePlayLog) GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error) {
	edition, err := g.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, service.ErrInvalidEdition
		}
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	stats, err := g.gamePlayLogRepository.GetEditionPlayStats(ctx, editionID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get edition play stats: %w", err)
	}

	stats.EditionName = edition.GetName()

	return stats, nil
}
