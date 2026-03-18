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

type GameCreator struct {
	gameCreatorRepo repository.GameCreator
	gameRepository  repository.GameV2
	db              repository.DB
	user            *User
}

func NewGameCreator(gameCreatorRepo repository.GameCreator, gameRepository repository.GameV2, db repository.DB, user *User) *GameCreator {
	return &GameCreator{
		gameCreatorRepo: gameCreatorRepo,
		gameRepository:  gameRepository,
		db:              db,
		user:            user,
	}
}

func (gc *GameCreator) GetGameCreators(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error) {
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("get game: %w", err)
	}

	creators, err := gc.gameCreatorRepo.GetGameCreatorsByGameID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("get game creators by game id: %w", err)
	}

	return creators, nil
}

func (gc *GameCreator) GetGameCreatorJobs(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorJob, []*domain.GameCreatorCustomJob, error) {
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get game: %w", err)
	}

	presetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get game creator preset jobs: %w", err)
	}

	customJobs, err := gc.gameCreatorRepo.GetGameCreatorCustomJobsByGameID(ctx, gameID)
	if err != nil {
		return nil, nil, fmt.Errorf("get game creator custom jobs by game id: %w", err)
	}

	return presetJobs, customJobs, nil
}

func (gc *GameCreator) EditGameCreators(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, inputs []*service.EditGameCreatorJobInput) error {
	// ゲームの存在確認
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrInvalidGameID
	}
	if err != nil {
		return fmt.Errorf("get game: %w", err)
	}

	// 存在するユーザーidを指定しているか確認
	// TODO: 今の実装は現役しか取得できないので、凍結済みユーザーを取得できるようにする
	users, err := gc.user.getActiveUsers(ctx, session)
	if err != nil {
		return fmt.Errorf("get active users: %w", err)
	}
	usersMap := make(map[values.TraPMemberID]*service.UserInfo, len(users))
	for _, user := range users {
		usersMap[user.GetID()] = user
	}
	for _, inputUser := range inputs {
		if _, ok := usersMap[inputUser.UserID]; !ok {
			return service.ErrInvalidUserID
		}
	}

	// 新しいcustom jobが既存のcustom jobと重複しないかチェック
	newCustomJobNames := make(map[values.GameCreatorJobDisplayName]struct{}, len(inputs))
	for _, input := range inputs {
		for _, jobName := range input.NewCustomJobNames {
			newCustomJobNames[jobName] = struct{}{}
		}
	}
	existingCustomJobs, err := gc.gameCreatorRepo.GetGameCreatorCustomJobsByGameID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("get game creator custom jobs by game id: %w", err)
	}
	for _, job := range existingCustomJobs {
		if _, ok := newCustomJobNames[job.GetDisplayName()]; ok {
			return service.ErrDuplicateCustomJobDisplayName
		}
	}

	// ユーザー内に同じjob idが含まれないかのチェック
	for _, userInput := range inputs {
		jobIDsMap := make(map[values.GameCreatorJobID]struct{}, len(userInput.Jobs))
		for _, jobID := range userInput.Jobs {
			if _, ok := jobIDsMap[jobID]; ok {
				return service.ErrDuplicateGameCreatorJobID
			}
			jobIDsMap[jobID] = struct{}{}
		}
	}

	presetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobs(ctx)
	if err != nil {
		return fmt.Errorf("get game creator preset jobs: %w", err)
	}
	presetJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(presetJobs))
	for _, job := range presetJobs {
		presetJobsMap[job.GetID()] = job
	}

	// まだDBにないcreatorを作成するために、差分をチェックする
	creatorUserIDs := make([]values.TraPMemberID, len(inputs))
	for i, input := range inputs {
		creatorUserIDs[i] = input.UserID
	}
	existingCreators, err := gc.gameCreatorRepo.GetCreatorsByUserIDs(ctx, creatorUserIDs)
	if err != nil {
		return fmt.Errorf("get creators by user ids: %w", err)
	}
	existingUserCreatorMap := make(map[values.TraPMemberID]*domain.GameCreator, len(existingCreators))
	for _, creator := range existingCreators {
		existingUserCreatorMap[creator.GetUserID()] = creator
	}

	err = gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		newCustomJobs := make([]*domain.GameCreatorCustomJob, 0, len(newCustomJobNames))
		for newCustomJobName := range newCustomJobNames {
			newCustomJobs = append(newCustomJobs, domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), newCustomJobName, gameID, time.Now()))
		}
		err := gc.gameCreatorRepo.CreateGameCreatorCustomJobs(ctx, newCustomJobs)
		if err != nil {
			return fmt.Errorf("create game creator custom jobs: %w", err)
		}
		newCustomJosMap := make(map[values.GameCreatorJobDisplayName]values.GameCreatorJobID, len(newCustomJobs))
		for _, job := range newCustomJobs {
			newCustomJosMap[job.GetDisplayName()] = job.GetID()
		}

		newCreators := make([]*domain.GameCreator, 0, len(creatorUserIDs)-len(existingCreators))
		for _, userID := range creatorUserIDs {
			if _, ok := existingUserCreatorMap[userID]; !ok {
				newCreator := domain.NewGameCreator(values.NewGameCreatorID(), userID, usersMap[userID].GetName(), time.Now())
				newCreators = append(newCreators, newCreator)
			}
		}
		err = gc.gameCreatorRepo.CreateGameCreators(ctx, newCreators)
		if err != nil {
			return fmt.Errorf("create game creators: %w", err)
		}

		userCreatorMap := make(map[values.TraPMemberID]*domain.GameCreator, len(inputs))
		for userID, creator := range existingUserCreatorMap {
			userCreatorMap[userID] = creator
		}
		for _, creator := range newCreators {
			userCreatorMap[creator.GetUserID()] = creator
		}

		presetJobsRelations := make(map[values.GameCreatorID][]values.GameCreatorJobID, len(inputs))
		for _, input := range inputs {
			if len(input.Jobs) == 0 {
				continue
			}
			creator, ok := userCreatorMap[input.UserID]
			if !ok {
				// 起きないはず
				return fmt.Errorf("user creator not found: %s", input.UserID)
			}
			presetJobIDs := make([]values.GameCreatorJobID, 0, len(input.Jobs))
			for _, jobID := range input.Jobs {
				if _, ok := presetJobsMap[jobID]; !ok {
					// custom jobなので含めない
					continue
				}
				presetJobIDs = append(presetJobIDs, jobID)
			}
			presetJobsRelations[creator.GetID()] = presetJobIDs
		}
		err = gc.gameCreatorRepo.UpsertGameCreatorPresetJobsRelations(ctx, presetJobsRelations)
		if err != nil {
			return fmt.Errorf("upsert game creator preset jobs relations: %w", err)
		}

		customJobRelations := make(map[values.GameCreatorID][]values.GameCreatorJobID, len(inputs))
		for _, input := range inputs {
			creator, ok := userCreatorMap[input.UserID]
			if !ok {
				// 起きないはず
				return fmt.Errorf("user creator not found: %s", input.UserID)
			}
			customJobIDs := make([]values.GameCreatorJobID, 0, len(input.Jobs))
			for _, jobID := range input.Jobs {
				if _, ok := presetJobsMap[jobID]; ok {
					// preset jobなので含めない
					continue
				}
				customJobIDs = append(customJobIDs, jobID)
			}
			for _, newCustomJobName := range input.NewCustomJobNames {
				newCustomJobID, ok := newCustomJosMap[newCustomJobName]
				if !ok {
					// 起きないはず
					return fmt.Errorf("new custom job id not found: %s", newCustomJobName)
				}
				customJobIDs = append(customJobIDs, newCustomJobID)
			}
			customJobRelations[creator.GetID()] = customJobIDs
		}
		err = gc.gameCreatorRepo.UpsertGameCreatorCustomJobsRelations(ctx, customJobRelations)
		if err != nil {
			return fmt.Errorf("upsert game creator custom jobs relations: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}
