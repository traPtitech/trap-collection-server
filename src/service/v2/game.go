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

type Game struct {
	db                 repository.DB
	gameRepository     repository.GameV2
	gameManagementRole repository.GameManagementRole
	user               *User
}

func NewGame(
	db repository.DB,
	gameRepository repository.GameV2,
	gameManagementRole repository.GameManagementRole,
	userUtils *User,
) *Game {
	return &Game{
		db:                 db,
		gameRepository:     gameRepository,
		gameManagementRole: gameManagementRole,
		user:               userUtils,
	}
}

func (g *Game) CreateGame(ctx context.Context, session *domain.OIDCSession, name values.GameName, description values.GameDescription, owners []values.TraPMemberName, maintainers []values.TraPMemberName) (*service.GameInfoV2, error) {
	user, err := g.user.getMe(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	game := domain.NewGame(values.NewGameID(), name, description, time.Now())

	var ownersInfo []*service.UserInfo
	var maintainersInfo []*service.UserInfo

	err = g.db.Transaction(ctx, nil, func(ctx context.Context) error {
		err := g.gameRepository.SaveGame(ctx, game)
		if err != nil {
			return fmt.Errorf("failed to save game: %w", err)
		}

		activeUsers, err := g.user.getActiveUsers(ctx, session) //ユーザー名=>uuidの変換のために全アクティブユーザーを取得
		if err != nil {
			return fmt.Errorf("failed to get active users: %w", err)
		}

		activeUsersMap := make(map[values.TraPMemberName]values.TraPMemberID, len(activeUsers))
		for _, activeUser := range activeUsers {
			activeUsersMap[activeUser.GetName()] = activeUser.GetID()
		}

		var ownersID []values.TraPMemberID
		for _, owner := range owners {
			if owner == user.GetName() { //ログイン中のユーザーがownersに含まれていたらエラー
				return service.ErrOverlapBetweenUserAndOwners
			}
			if ownerID, ok := activeUsersMap[owner]; ok { //ownerが存在するユーザーが確かめる
				ownersID = append(ownersID, ownerID)
			}
			ownerInfo := service.NewUserInfo(
				activeUsersMap[owner],
				owner,
				values.TrapMemberStatusActive,
			)
			ownersInfo = append(ownersInfo, ownerInfo)
		}
		owners = append(owners, user.GetName()) //ログイン中のユーザーをownersに追加

		ownersMap := make(map[values.TraPMemberName]struct{})
		for _, owner := range owners {
			ownersMap[owner] = struct{}{}
		}

		var maintainersID []values.TraPMemberID
		for _, maintainer := range maintainers {
			if _, ok := ownersMap[maintainer]; ok { //ownerとmaintainerは重複しない
				return service.ErrOverlapBetweenOwnersAndMaintainers
			}

			if maintainerID, ok := activeUsersMap[maintainer]; ok { //ユーザーが存在するか確認
				maintainersID = append(maintainersID, maintainerID)

				maintainerInfo := service.NewUserInfo(
					activeUsersMap[maintainer],
					maintainer,
					values.TrapMemberStatusActive,
				)
				maintainersInfo = append(maintainersInfo, maintainerInfo)
			}

		}

		err = g.gameManagementRole.AddGameManagementRoles(
			ctx,
			game.GetID(),
			ownersID,
			values.GameManagementRoleAdministrator)
		if err != nil {
			return fmt.Errorf("failed to add management role 'owner': %w", err)
		}

		err = g.gameManagementRole.AddGameManagementRoles(
			ctx,
			game.GetID(),
			maintainersID,
			values.GameManagementRoleCollaborator)
		if err != nil {
			return fmt.Errorf("failed to add management role 'maintainer': %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	gameInfo := &service.GameInfoV2{
		Game:        game,
		Owners:      ownersInfo,
		Maintainers: maintainersInfo,
	}
	return gameInfo, nil
}

func (g *Game) GetGame(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) (*service.GameInfoV2, error) {
	game, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGame
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	//管理者たちを取得
	administrators, err := g.gameManagementRole.GetGameManagersByGameID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game management role: %w", err)
	}

	activeUsers, err := g.user.getActiveUsers(ctx, session) //ユーザー名=>uuidの変換のために全アクティブユーザーを取得
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	activeUsersMap := make(map[values.TraPMemberID]values.TraPMemberName, len(activeUsers))
	for _, activeUser := range activeUsers {
		activeUsersMap[activeUser.GetID()] = activeUser.GetName()
	}

	var ownersInfo []*service.UserInfo
	var maintainersInfo []*service.UserInfo
	for _, administrator := range administrators {
		switch administrator.Role {
		case values.GameManagementRoleAdministrator:
			if ownerName, ok := activeUsersMap[administrator.UserID]; ok {
				ownerInfo := service.NewUserInfo(
					administrator.UserID,
					ownerName,
					values.TrapMemberStatusActive,
				)
				ownersInfo = append(ownersInfo, ownerInfo)
			}
		case values.GameManagementRoleCollaborator:
			if maintainerName, ok := activeUsersMap[administrator.UserID]; ok {
				maintainerInfo := service.NewUserInfo(
					administrator.UserID,
					maintainerName,
					values.TrapMemberStatusActive,
				)
				maintainersInfo = append(maintainersInfo, maintainerInfo)
			}
		default:
			fmt.Println("invalid administrator role")
		}
	}

	gameInfo := &service.GameInfoV2{
		Game:        game,
		Owners:      ownersInfo,
		Maintainers: maintainersInfo,
	}

	return gameInfo, nil
}

func (g *Game) GetGames(ctx context.Context, limit int, offset int) (int, []*domain.Game, error) {
	games, gameNumber, err := g.gameRepository.GetGames(ctx, limit, offset)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get games: %w", err)
	}
	if len(games) == 0 {
		return 0, []*domain.Game{}, nil
	}
	return gameNumber, games, nil
}

func (g *Game) GetMyGames(ctx context.Context, session *domain.OIDCSession, limit int, offset int) (int, []*domain.Game, error) {
	user, err := g.user.getMe(ctx, session)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get user: %w", err)
	}

	myGames, gameNumber, err := g.gameRepository.GetGamesByUser(ctx, user.GetID(), limit, offset)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get game IDs: %w", err)
	}

	if len(myGames) == 0 {
		return 0, []*domain.Game{}, nil
	}

	return gameNumber, myGames, nil
}

func (g *Game) UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription) (*domain.Game, error) { //V1と変わらず
	var game *domain.Game
	err := g.db.Transaction(ctx, nil, func(ctx context.Context) error {
		var err error
		game, err = g.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGame
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		// 変更がなければ何もしない
		if game.GetName() == name && game.GetDescription() == description {
			return nil
		}

		game.SetName(name)
		game.SetDescription(description)

		err = g.gameRepository.UpdateGame(ctx, game)
		if err != nil {
			return fmt.Errorf("failed to save game: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return game, nil
}

func (g *Game) DeleteGame(ctx context.Context, gameID values.GameID) error { //V1と変わらない
	err := g.gameRepository.RemoveGame(ctx, gameID)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return service.ErrNoGame
	}
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}
