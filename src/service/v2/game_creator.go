package v2

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
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

func (gc *GameCreator) CreateGameCustomJob(ctx context.Context, gameID values.GameID, displayName values.GameCreatorJobDisplayName) (*domain.GameCreatorCustomJob, error) {
	var customJob *domain.GameCreatorCustomJob
	err := gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("get game: %w", err)
		}

		existingCustomJobs, err := gc.gameCreatorRepo.GetGameCreatorCustomJobsByGameID(ctx, gameID)
		if err != nil {
			return fmt.Errorf("get game creator custom jobs by game id: %w", err)
		}
		for _, existingCustomJob := range existingCustomJobs {
			if existingCustomJob.GetDisplayName() == displayName {
				return service.ErrDuplicateCustomJobDisplayName
			}
		}

		customJob = domain.NewGameCreatorCustomJob(
			values.NewGameCreatorJobID(),
			displayName,
			gameID,
			time.Now(),
		)
		err = gc.gameCreatorRepo.CreateGameCreatorCustomJobs(ctx, []*domain.GameCreatorCustomJob{customJob})
		if err != nil {
			return fmt.Errorf("create game creator custom job: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction: %w", err)
	}

	return customJob, nil
}

func (gc *GameCreator) DeleteGameCreator(ctx context.Context, gameID values.GameID, creatorID values.GameCreatorID) error {
	err := gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		creator, err := gc.gameCreatorRepo.GetGameCreatorByID(ctx, creatorID)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameCreatorID
		}
		if err != nil {
			return fmt.Errorf("get game creator by id: %w", err)
		}

		if creator.GetGameID() != gameID {
			return service.ErrInvalidGameCreatorGamePair
		}

		err = gc.gameCreatorRepo.DeleteGameCreatorCustomJobs(ctx, creatorID)
		if err != nil {
			return fmt.Errorf("delete game creator custom jobs: %w", err)
		}
		err = gc.gameCreatorRepo.DeleteGameCreatorPresetJobs(ctx, creatorID)
		if err != nil {
			return fmt.Errorf("delete game creator preset jobs: %w", err)
		}

		err = gc.gameCreatorRepo.DeleteGameCreator(ctx, gameID, creatorID)
		if errors.Is(err, repository.ErrNoRecordDeleted) {
			// ここまでの処理で弾けているはずだが一応
			return service.ErrInvalidGameCreatorGamePair
		}
		if err != nil {
			return fmt.Errorf("delete game creator: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}

func (gc *GameCreator) CreateGameCreator(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userID values.TraPMemberID) (*domain.GameCreator, error) {
	users, err := gc.user.getAllUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	index := slices.IndexFunc(users, func(user *service.UserInfo) bool {
		return user.GetID() == userID
	})
	if index == -1 {
		return nil, service.ErrInvalidUserID
	}
	user := users[index]

	creator := domain.NewGameCreator(values.NewGameCreatorID(), userID, gameID, user.GetName(), time.Now())
	err = gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err = gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("get game: %w", err)
		}

		err = gc.gameCreatorRepo.CreateGameCreators(ctx, []*domain.GameCreator{creator})
		if errors.Is(err, repository.ErrDuplicatedUniqueKey) {
			return service.ErrDuplicateGameCreator
		}
		if err != nil {
			return fmt.Errorf("create game creator: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return creator, nil
}

func (gc *GameCreator) SetGameCreatorJobs(ctx context.Context, gameID values.GameID, creatorID values.GameCreatorID, jobIDs []values.GameCreatorJobID) error {
	err := gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		creator, err := gc.gameCreatorRepo.GetGameCreatorByID(ctx, creatorID)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameCreatorID
		}
		if err != nil {
			return fmt.Errorf("get game creator by id: %w", err)
		}

		if creator.GetGameID() != gameID {
			return service.ErrInvalidGameCreatorGamePair
		}

		jobIDsMap := make(map[values.GameCreatorJobID]struct{}, len(jobIDs))
		for _, jobID := range jobIDs {
			jobIDsMap[jobID] = struct{}{}
		}

		allPresetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobs(ctx)
		if err != nil {
			return fmt.Errorf("get game creator preset jobs: %w", err)
		}
		allPresetJobIDsMap := make(map[values.GameCreatorJobID]struct{}, len(allPresetJobs))
		for _, job := range allPresetJobs {
			allPresetJobIDsMap[job.GetID()] = struct{}{}
		}

		presetJobIDs, customJobIDs := make([]values.GameCreatorJobID, 0, len(jobIDs)), make([]values.GameCreatorJobID, 0, len(jobIDs))
		for _, jobID := range jobIDs {
			if _, isPreset := allPresetJobIDsMap[jobID]; isPreset {
				presetJobIDs = append(presetJobIDs, jobID)
			} else {
				customJobIDs = append(customJobIDs, jobID)
			}
		}

		customJobs, err := gc.gameCreatorRepo.GetGameCreatorCustomJobsByCreatorID(ctx, creatorID)
		if err != nil {
			return fmt.Errorf("get game creator custom jobs by creator id: %w", err)
		}

		presetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobsByCreatorID(ctx, creatorID)
		if err != nil {
			return fmt.Errorf("get game creator preset jobs by creator id: %w", err)
		}

		deletingPresetJobIDs := make([]values.GameCreatorJobID, 0, len(presetJobs))
		for _, presetJob := range presetJobs {
			presetJobID := presetJob.GetID()
			if _, exists := jobIDsMap[presetJobID]; !exists {
				deletingPresetJobIDs = append(deletingPresetJobIDs, presetJobID)
			}
		}
		err = gc.gameCreatorRepo.DeleteGameCreatorPresetJobsByCreatorID(ctx, creatorID, deletingPresetJobIDs)
		if err != nil {
			return fmt.Errorf("delete game creator preset jobs by creator id: %w", err)
		}

		deletingCustomJobIDs := make([]values.GameCreatorJobID, 0, len(customJobs))
		for _, customJob := range customJobs {
			customJobID := customJob.GetID()
			if _, exists := jobIDsMap[customJobID]; !exists {
				deletingCustomJobIDs = append(deletingCustomJobIDs, customJobID)
			}
		}
		err = gc.gameCreatorRepo.DeleteGameCreatorCustomJobsByCreatorID(ctx, creatorID, deletingCustomJobIDs)
		if err != nil {
			return fmt.Errorf("delete game creator custom jobs by creator id: %w", err)
		}

		if err := gc.gameCreatorRepo.UpsertGameCreatorPresetJobsRelations(ctx, map[values.GameCreatorID][]values.GameCreatorJobID{
			creatorID: presetJobIDs,
		}); err != nil {
			if errors.Is(err, repository.ErrForeignKeyViolated) {
				return service.ErrInvalidPresetJobID
			}
			return fmt.Errorf("upsert game creator preset jobs relations: %w", err)
		}

		if err := gc.gameCreatorRepo.UpsertGameCreatorCustomJobsRelations(ctx, map[values.GameCreatorID][]values.GameCreatorJobID{
			creatorID: customJobIDs,
		}); err != nil {
			if errors.Is(err, repository.ErrForeignKeyViolated) {
				return service.ErrInvalidGameAndCustomJobPair
			}
			return fmt.Errorf("upsert game creator custom jobs relations: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}

func (gc *GameCreator) EditGameCreators(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, inputs []*service.EditGameCreatorJobInput) error {
	err := gc.validateGameExists(ctx, gameID)
	if err != nil {
		return err
	}

	presetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobs(ctx)
	if err != nil {
		return fmt.Errorf("get game creator preset jobs: %w", err)
	}
	presetJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(presetJobs))
	for _, job := range presetJobs {
		presetJobsMap[job.GetID()] = job
	}

	validatedInput, err := gc.validateEditGameCreatorsInput(ctx, session, gameID, presetJobsMap, inputs)
	if err != nil {
		return err
	}

	inputData, err := gc.loadEditGameCreatorsInputData(ctx, gameID, inputs)
	if err != nil {
		return err
	}

	err = gc.db.Transaction(ctx, nil, func(ctx context.Context) error {
		return gc.applyEditGameCreators(ctx, gameID, presetJobsMap, inputs, validatedInput, inputData)
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}

type editGameCreatorsValidatedInput struct {
	usersMap          map[values.TraPMemberID]*service.UserInfo
	newCustomJobNames map[values.GameCreatorJobDisplayName]struct{}
}

func (gc *GameCreator) validateGameExists(ctx context.Context, gameID values.GameID) error {
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrInvalidGameID
	}
	if err != nil {
		return fmt.Errorf("get game: %w", err)
	}

	return nil
}

func (gc *GameCreator) validateEditGameCreatorsInput(
	ctx context.Context,
	session *domain.OIDCSession,
	gameID values.GameID,
	presetJobsMap map[values.GameCreatorJobID]*domain.GameCreatorJob,
	inputs []*service.EditGameCreatorJobInput,
) (*editGameCreatorsValidatedInput, error) {
	activeMembers, err := gc.user.getAllUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	activeMembersMap := make(map[values.TraPMemberID]*service.UserInfo, len(activeMembers))
	for _, user := range activeMembers {
		activeMembersMap[user.GetID()] = user
	}
	for _, input := range inputs {
		if _, ok := activeMembersMap[input.UserID]; !ok {
			return nil, service.ErrInvalidUserID
		}
	}
	inputUsersMap := make(map[values.TraPMemberID]struct{}, len(inputs))
	for _, input := range inputs {
		if _, ok := inputUsersMap[input.UserID]; ok {
			return nil, service.ErrDuplicateUserID
		}
		inputUsersMap[input.UserID] = struct{}{}
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
		return nil, fmt.Errorf("get game creator custom jobs by game id: %w", err)
	}
	for _, job := range existingCustomJobs {
		if _, ok := newCustomJobNames[job.GetDisplayName()]; ok {
			return nil, service.ErrDuplicateCustomJobDisplayName
		}
	}

	// 入力の job id が既存の job id と一致するかチェック
	jobIDsMap := make(map[values.GameCreatorJobID]struct{}, len(existingCustomJobs))
	for _, job := range existingCustomJobs {
		jobIDsMap[job.GetID()] = struct{}{}
	}
	for _, input := range inputs {
		for _, jobID := range input.Jobs {
			_, isPresetJob := presetJobsMap[jobID]
			_, isCustomJob := jobIDsMap[jobID]
			if !isPresetJob && !isCustomJob {
				return nil, service.ErrInvalidGameCreatorJobID
			}
		}
	}

	// 1人のユーザー内に同じjob idが含まれないかのチェック
	for _, input := range inputs {
		jobIDsMap := make(map[values.GameCreatorJobID]struct{}, len(input.Jobs))
		for _, jobID := range input.Jobs {
			if _, ok := jobIDsMap[jobID]; ok {
				return nil, service.ErrDuplicateGameCreatorJobID
			}
			jobIDsMap[jobID] = struct{}{}
		}
	}

	return &editGameCreatorsValidatedInput{
		usersMap:          activeMembersMap,
		newCustomJobNames: newCustomJobNames,
	}, nil
}

type editGameCreatorsInputData struct {
	creatorUserIDs         []values.TraPMemberID
	existingUserCreatorMap map[values.TraPMemberID]*domain.GameCreator
}

func (gc *GameCreator) loadEditGameCreatorsInputData(
	ctx context.Context,
	gameID values.GameID,
	inputs []*service.EditGameCreatorJobInput,
) (*editGameCreatorsInputData, error) {
	creatorUserIDs := make([]values.TraPMemberID, len(inputs))
	for i, input := range inputs {
		creatorUserIDs[i] = input.UserID
	}

	existingCreators, err := gc.gameCreatorRepo.GetGameCreatorsByUserIDs(ctx, gameID, creatorUserIDs)
	if err != nil {
		return nil, fmt.Errorf("get creators by user ids: %w", err)
	}

	existingUserCreatorMap := make(map[values.TraPMemberID]*domain.GameCreator, len(existingCreators))
	for _, creator := range existingCreators {
		existingUserCreatorMap[creator.GetUserID()] = creator
	}

	return &editGameCreatorsInputData{
		creatorUserIDs:         creatorUserIDs,
		existingUserCreatorMap: existingUserCreatorMap,
	}, nil
}

func (gc *GameCreator) applyEditGameCreators(
	ctx context.Context,
	gameID values.GameID,
	presetJobsMap map[values.GameCreatorJobID]*domain.GameCreatorJob,
	inputs []*service.EditGameCreatorJobInput,
	validatedInput *editGameCreatorsValidatedInput,
	inputData *editGameCreatorsInputData,
) error {
	newCustomJobs := make([]*domain.GameCreatorCustomJob, 0, len(validatedInput.newCustomJobNames))
	for newCustomJobName := range validatedInput.newCustomJobNames {
		newCustomJobs = append(newCustomJobs, domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), newCustomJobName, gameID, time.Now()))
	}
	err := gc.gameCreatorRepo.CreateGameCreatorCustomJobs(ctx, newCustomJobs)
	if err != nil {
		return fmt.Errorf("create game creator custom jobs: %w", err)
	}

	newCustomJobsMap := make(map[values.GameCreatorJobDisplayName]values.GameCreatorJobID, len(newCustomJobs))
	for _, job := range newCustomJobs {
		newCustomJobsMap[job.GetDisplayName()] = job.GetID()
	}

	newCreators := make([]*domain.GameCreator, 0, len(inputData.creatorUserIDs)-len(inputData.existingUserCreatorMap))
	for _, userID := range inputData.creatorUserIDs {
		if _, ok := inputData.existingUserCreatorMap[userID]; !ok {
			newCreator := domain.NewGameCreator(values.NewGameCreatorID(), userID, gameID, validatedInput.usersMap[userID].GetName(), time.Now())
			newCreators = append(newCreators, newCreator)
		}
	}
	err = gc.gameCreatorRepo.CreateGameCreators(ctx, newCreators)
	if err != nil {
		return fmt.Errorf("create game creators: %w", err)
	}

	userCreatorMap := make(map[values.TraPMemberID]*domain.GameCreator, len(inputs))
	maps.Copy(userCreatorMap, inputData.existingUserCreatorMap)
	for _, creator := range newCreators {
		userCreatorMap[creator.GetUserID()] = creator
	}

	presetRelations, err := buildPresetJobRelations(inputs, userCreatorMap, presetJobsMap)
	if err != nil {
		return err
	}
	err = gc.gameCreatorRepo.UpsertGameCreatorPresetJobsRelations(ctx, presetRelations)
	if err != nil {
		return fmt.Errorf("upsert game creator preset jobs relations: %w", err)
	}

	customRelations, err := buildCustomJobRelations(inputs, userCreatorMap, presetJobsMap, newCustomJobsMap)
	if err != nil {
		return err
	}
	err = gc.gameCreatorRepo.UpsertGameCreatorCustomJobsRelations(ctx, customRelations)
	if err != nil {
		return fmt.Errorf("upsert game creator custom jobs relations: %w", err)
	}

	return nil
}

func buildPresetJobRelations(
	inputs []*service.EditGameCreatorJobInput,
	userCreatorMap map[values.TraPMemberID]*domain.GameCreator,
	presetJobsMap map[values.GameCreatorJobID]*domain.GameCreatorJob,
) (map[values.GameCreatorID][]values.GameCreatorJobID, error) {
	presetRelations := make(map[values.GameCreatorID][]values.GameCreatorJobID, len(inputs))
	for _, input := range inputs {
		creator, ok := userCreatorMap[input.UserID]
		if !ok {
			// 起きないはず
			return nil, fmt.Errorf("user creator not found: %s", input.UserID)
		}

		presetJobIDs := make([]values.GameCreatorJobID, 0, len(input.Jobs))
		for _, jobID := range input.Jobs {
			if _, ok := presetJobsMap[jobID]; !ok {
				continue
			}
			presetJobIDs = append(presetJobIDs, jobID)
		}
		presetRelations[creator.GetID()] = presetJobIDs
	}

	return presetRelations, nil
}

func buildCustomJobRelations(
	inputs []*service.EditGameCreatorJobInput,
	userCreatorMap map[values.TraPMemberID]*domain.GameCreator,
	presetJobsMap map[values.GameCreatorJobID]*domain.GameCreatorJob,
	newCustomJobsMap map[values.GameCreatorJobDisplayName]values.GameCreatorJobID,
) (map[values.GameCreatorID][]values.GameCreatorJobID, error) {
	customRelations := make(map[values.GameCreatorID][]values.GameCreatorJobID, len(inputs))
	for _, input := range inputs {
		creator, ok := userCreatorMap[input.UserID]
		if !ok {
			// 起きないはず
			return nil, fmt.Errorf("user creator not found: %s", input.UserID)
		}

		customJobIDs := make([]values.GameCreatorJobID, 0, len(input.Jobs)+len(input.NewCustomJobNames))
		for _, jobID := range input.Jobs {
			if _, ok := presetJobsMap[jobID]; ok {
				continue
			}
			customJobIDs = append(customJobIDs, jobID)
		}

		for _, newCustomJobName := range input.NewCustomJobNames {
			newCustomJobID, ok := newCustomJobsMap[newCustomJobName]
			if !ok {
				return nil, fmt.Errorf("new custom job id not found: %s", newCustomJobName)
			}
			customJobIDs = append(customJobIDs, newCustomJobID)
		}

		customRelations[creator.GetID()] = customJobIDs
	}

	return customRelations, nil
}
