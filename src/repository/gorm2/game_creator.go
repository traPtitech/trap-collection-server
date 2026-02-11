package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

var _ repository.GameCreator = (*GameCreator)(nil)

type GameCreator struct {
	db *DB
}

func NewGameCreator(db *DB) *GameCreator {
	return &GameCreator{
		db: db,
	}
}

func (gc *GameCreator) GetGameCreatorsByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error) {
	db, err := gc.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db: %w", err)
	}

	var gameCreators []schema.GameCreatorTable
	err = db.Preload("CreatorJobs").
		Preload("CustomCreatorJobs").
		Where("game_creators.game_id = ?", uuid.UUID(gameID)).
		Order("game_creators.created_at ASC").
		Find(&gameCreators).Error
	if err != nil {
		return nil, fmt.Errorf("find game creators: %w", err)
	}

	result := make([]*domain.GameCreatorWithJobs, 0, len(gameCreators))
	for _, gc := range gameCreators {
		jobs := make([]*domain.GameCreatorJob, 0, len(gc.CreatorJobs))
		for _, job := range gc.CreatorJobs {
			jobs = append(jobs,
				domain.NewGameCreatorJob(
					values.GameCreatorJobID(job.ID),
					values.GameCreatorJobDisplayName(job.DisplayName),
					job.CreatedAt,
				))
		}
		customJobs := make([]*domain.GameCreatorCustomJob, 0, len(gc.CustomCreatorJobs))
		for _, job := range gc.CustomCreatorJobs {
			customJobs = append(customJobs,
				domain.NewGameCreatorCustomJob(values.GameCreatorJobID(job.ID),
					values.GameCreatorJobDisplayName(job.DisplayName),
					values.GameID(job.GameID),
					job.CreatedAt,
				))
		}

		result = append(result, domain.NewGameCreatorWithJobs(
			domain.NewGameCreator(
				values.GameCreatorID(gc.ID),
				values.TraPMemberID(gc.UserID),
				values.TraPMemberName(gc.UserName),
				gc.CreatedAt),
			jobs,
			customJobs,
		))
	}

	return result, nil
}

func (gc *GameCreator) GetGameCreatorPresetJobs(_ context.Context) ([]*domain.GameCreatorJob, error) {
	return nil, nil // TODO: implement
}

func (gc *GameCreator) GetGameCreatorCustomJobsByGameID(_ context.Context, _ values.GameID) ([]*domain.GameCreatorCustomJob, error) {
	return nil, nil // TODO: implement
}
