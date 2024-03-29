package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockAuth "github.com/traPtitech/trap-collection-server/src/auth/mock"
	"github.com/traPtitech/trap-collection-server/src/cache"
	mockCache "github.com/traPtitech/trap-collection-server/src/cache/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	user := NewUser(mockUserAuth, mockUserCache)

	gameService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		user,
	)

	type test struct {
		description                               string
		authSession                               *domain.OIDCSession
		user                                      *service.UserInfo
		isGetMeErr                                bool
		name                                      values.GameName
		gameDescription                           values.GameDescription
		owners                                    []values.TraPMemberName
		maintainers                               []values.TraPMemberName
		executeSaveGame                           bool
		SaveGameErr                               error
		executeAddGameManagementRolesAdmin        bool
		executeAddGameManagementRolesCollaborator bool
		AddGameManagementRoleAdminErr             error
		AddGameManagementRoleCollabErr            error
		expectedOwners                            []values.TraPMemberName
		isErr                                     bool
		err                                       error
	}

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())
	userID3 := values.NewTrapMemberID(uuid.New())
	userID4 := values.NewTrapMemberID(uuid.New())

	activeUsers := []*service.UserInfo{
		service.NewUserInfo(
			userID1,
			"ikura-hamu",
			values.TrapMemberStatusActive,
		),
		service.NewUserInfo(
			userID2,
			"mazrean",
			values.TrapMemberStatusActive,
		),
		service.NewUserInfo(
			userID3,
			"pikachu",
			values.TrapMemberStatusActive,
		),
		service.NewUserInfo(
			userID4,
			"JichouP",
			values.TrapMemberStatusActive,
		),
	}

	testCases := []test{
		{
			description: "ユーザー情報の取得に失敗したのでエラー",
			authSession: domain.NewOIDCSession("access token", time.Now().Add(time.Hour)),
			isGetMeErr:  true,
			isErr:       true,
		},
		{
			description: "特に問題ないのでエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				userID1,
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test description"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "nameが空でもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName(""),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "descriptionが空でもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription(""),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "ownersが複数いてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean", "JichouP"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "JichouP", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "ownersがいなくてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "maintainersが複数いてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu", "JichouP"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "maintainersがいなくてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription(""),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "ownerとユーザーが同じなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"ikura-hamu"},
			maintainers:     []values.TraPMemberName{"pikachu"},
			executeSaveGame: true,
			isErr:           true,
			err:             service.ErrOverlapInOwners,
		},
		{
			description: "maintainerとユーザーが同じなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"mazrean"},
			maintainers:     []values.TraPMemberName{"ikura-hamu"},
			executeSaveGame: true,
			isErr:           true,
			err:             service.ErrOverlapBetweenOwnersAndMaintainers,
		},
		{
			description: "ownersに同じ人が含まれているのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"mazrean", "mazrean"},
			maintainers:     []values.TraPMemberName{"pikachu"},
			executeSaveGame: true,
			isErr:           true,
			err:             service.ErrOverlapInOwners,
		},
		{
			description: "maintainersに同じ人が含まれているのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"mazrean"},
			maintainers:     []values.TraPMemberName{"pikachu", "pikachu"},
			executeSaveGame: true,
			isErr:           true,
			err:             service.ErrOverlapInMaintainers,
		},
		{
			description: "ownersとmaintainersに同じ人がいるのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"pikachu"},
			maintainers:     []values.TraPMemberName{"pikachu"},
			executeSaveGame: true,
			isErr:           true,
			err:             service.ErrOverlapBetweenOwnersAndMaintainers,
		},
		{
			description: "ownersにactiveUserでない人が含まれるが問題なし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"s9"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			expectedOwners:                     []values.TraPMemberName{"ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "maintainersにactiveUserでない人が含まれるが問題なし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"s9"},
			expectedOwners:                     []values.TraPMemberName{"mazrean", "ikura-hamu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
		},
		{
			description: "SaveGameがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			owners:          []values.TraPMemberName{"mazrean"},
			maintainers:     []values.TraPMemberName{"pikachu"},
			executeSaveGame: true,
			SaveGameErr:     errors.New("test"),
			isErr:           true,
		},
		{
			description: "AddGameManagementRolesがownerの追加でエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			AddGameManagementRoleAdminErr:      errors.New("test"),
			isErr:                              true,
		},
		{
			description: "AddGameManagementRolesがmaintainerの追加でエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"ikura-hamu",
				values.TrapMemberStatusActive,
			),
			name:                               values.GameName("test"),
			gameDescription:                    values.GameDescription("test"),
			owners:                             []values.TraPMemberName{"mazrean"},
			maintainers:                        []values.TraPMemberName{"pikachu"},
			executeSaveGame:                    true,
			executeAddGameManagementRolesAdmin: true,
			executeAddGameManagementRolesCollaborator: true,
			AddGameManagementRoleCollabErr:            errors.New("test"),
			isErr:                                     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isGetMeErr {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(nil, cache.ErrCacheMiss)
				mockUserAuth.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession).
					Return(nil, errors.New("error"))
			} else {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(testCase.user, nil)

				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(activeUsers, nil).AnyTimes()
			}

			if testCase.executeSaveGame {
				mockGameRepository.
					EXPECT().
					SaveGame(gomock.Any(), gomock.Any()).
					Return(testCase.SaveGameErr)
			}

			if testCase.executeAddGameManagementRolesAdmin {
				mockGameManagementRoleRepository.
					EXPECT().
					AddGameManagementRoles(gomock.Any(), gomock.Any(), gomock.Any(), values.GameManagementRoleAdministrator).
					Return(testCase.AddGameManagementRoleAdminErr)
			}
			if testCase.executeAddGameManagementRolesCollaborator {
				mockGameManagementRoleRepository.
					EXPECT().
					AddGameManagementRoles(gomock.Any(), gomock.Any(), gomock.Any(), values.GameManagementRoleCollaborator).
					Return(testCase.AddGameManagementRoleCollabErr)
			}

			game, err := gameService.CreateGame(ctx, testCase.authSession, testCase.name, testCase.gameDescription, testCase.owners, testCase.maintainers)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.name, game.Game.GetName())
			assert.Equal(t, testCase.gameDescription, game.Game.GetDescription())
			for i := 0; i < len(game.Owners); i++ {
				assert.Equal(t, testCase.expectedOwners[i], game.Owners[i].GetName())
			}
			for i := 0; i < len(game.Maintainers); i++ {
				assert.Equal(t, testCase.maintainers[i], game.Maintainers[i].GetName())
			}
			assert.WithinDuration(t, time.Now(), game.Game.GetCreatedAt(), time.Second)

		})
	}
}

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUser(mockUserAuth, mockUserCache)

	gameService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		userUtils,
	)

	type test struct {
		description                    string
		gameID                         values.GameID
		game                           *domain.Game
		executeGetActiveUsers          bool
		GetGameErr                     error
		executeGetGameManagersByGameID bool
		administrators                 []*repository.UserIDAndManagementRole
		GetGameManagersByGameIDErr     error
		executeGetGenresByGameID       bool
		genres                         []*domain.GameGenre
		GetGenresByGameIDErr           error
		isErr                          bool
		err                            error
	}

	gameID := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())

	gameGenreID := values.NewGameGenreID()
	gameGenreName := values.NewGameGenreName("ジャンル")

	activeUsers := []*service.UserInfo{
		service.NewUserInfo(
			userID1,
			"ikura-hamu",
			values.TrapMemberStatusActive,
		),
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID,
			game: domain.NewGame(
				gameID,
				"game name",
				"game description",
				time.Now(),
			),
			executeGetActiveUsers:          true,
			executeGetGameManagersByGameID: true,
			administrators: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetGenresByGameID: true,
			genres:                   []*domain.GameGenre{domain.NewGameGenre(gameGenreID, gameGenreName, time.Now().Add(-time.Hour))},
		},
		{
			description: "ゲームが存在しないのでErrNoGame",
			gameID:      gameID,
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrNoGame,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      gameID,
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                    "GetGameManagementRolesByGameIDがエラーなのでエラー",
			gameID:                         gameID,
			executeGetGameManagersByGameID: true,
			GetGameManagersByGameIDErr:     errors.New("error"),
			isErr:                          true,
		},
		{
			description:                    "GetGameGenresByGameIDがエラーなのでエラー",
			gameID:                         gameID,
			executeGetGameManagersByGameID: true,
			executeGetActiveUsers:          true,
			executeGetGenresByGameID:       true,
			GetGenresByGameIDErr:           errors.New("error"),
			isErr:                          true,
		},
		{
			description: "Genreが空でも問題ない",
			gameID:      gameID,
			game: domain.NewGame(
				gameID,
				"game name",
				"game description",
				time.Now(),
			),
			executeGetActiveUsers:          true,
			executeGetGameManagersByGameID: true,
			administrators: []*repository.UserIDAndManagementRole{
				{
					UserID: userID1,
					Role:   values.GameManagementRoleAdministrator,
				},
			},
			executeGetGenresByGameID: true,
			genres:                   []*domain.GameGenre{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(testCase.game, testCase.GetGameErr)

			if testCase.executeGetActiveUsers {
				mockUserCache.
					EXPECT().
					GetActiveUsers(gomock.Any()).
					Return(activeUsers, nil)
			}
			if testCase.executeGetGameManagersByGameID {
				mockGameManagementRoleRepository.
					EXPECT().
					GetGameManagersByGameID(ctx, testCase.gameID).
					Return(testCase.administrators, testCase.GetGameManagersByGameIDErr)
			}
			if testCase.executeGetGenresByGameID {
				mockGameGenreRepository.
					EXPECT().
					GetGenresByGameID(ctx, testCase.gameID).
					Return(testCase.genres, testCase.GetGenresByGameIDErr)
			}

			gameInfo, err := gameService.GetGame(ctx, domain.NewOIDCSession("access token", time.Now().Add(time.Hour)), testCase.gameID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.game, gameInfo.Game)

			assert.Equal(t, testCase.game.GetID(), gameInfo.Game.GetID())
			assert.Equal(t, testCase.game.GetName(), gameInfo.Game.GetName())
			assert.Equal(t, testCase.game.GetDescription(), gameInfo.Game.GetDescription())
			assert.WithinDuration(t, testCase.game.GetCreatedAt(), gameInfo.Game.GetCreatedAt(), time.Second)

			if testCase.administrators != nil {
				for i := 0; i < len(testCase.administrators); i++ {
					assert.Equal(t, testCase.administrators[i].UserID, gameInfo.Owners[i].GetID())
				}
			} else {
				assert.Nil(t, gameInfo.Maintainers)
			}

			if testCase.genres != nil {
				for i := 0; i < len(testCase.genres); i++ {
					assert.Equal(t, testCase.genres[i], gameInfo.Genres[i])

					assert.Equal(t, testCase.genres[i].GetID(), gameInfo.Genres[i].GetID())
					assert.Equal(t, testCase.genres[i].GetName(), gameInfo.Genres[i].GetName())
					assert.WithinDuration(t, testCase.genres[i].GetCreatedAt(), gameInfo.Genres[i].GetCreatedAt(), time.Second)
				}
			} else {
				assert.Nil(t, gameInfo.Genres)
			}
		})
	}
}

func TestGetGames(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUser(mockUserAuth, mockUserCache)

	gameService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		userUtils,
	)

	type test struct {
		description     string
		limit           int
		offset          int
		n               int
		executeGetGames bool
		games           []*domain.Game
		GetGamesErr     error
		isErr           bool
		err             error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game descriptiion",
					time.Now(),
				),
			},
			executeGetGames: true,
			limit:           0,
			offset:          0,
			n:               1,
		},
		{
			description:     "ゲームが存在しなくてもエラーなし",
			games:           []*domain.Game{},
			limit:           0,
			offset:          0,
			n:               0,
			executeGetGames: true,
		},
		{
			description: "ゲームが複数でもエラーなし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
				domain.NewGame(
					gameID2,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:           0,
			offset:          0,
			n:               2,
			executeGetGames: true,
		},
		{
			description: "limitが設定されてもエラーなし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:           1,
			offset:          0,
			n:               1,
			executeGetGames: true,
		},
		{
			description: "limitとoffsetが両方設定されてもエラーなし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:           1,
			offset:          1,
			n:               2,
			executeGetGames: true,
		},
		{
			description: "offsetだけ設定されているのでエラー",
			limit:       0,
			offset:      1,
			isErr:       true,
			err:         service.ErrOffsetWithoutLimit,
		},

		{
			description:     "GetGamesがエラーなのでエラー",
			GetGamesErr:     errors.New("error"),
			isErr:           true,
			executeGetGames: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetGames {
				mockGameRepository.
					EXPECT().
					GetGames(gomock.Any(), testCase.limit, testCase.offset).
					Return(testCase.games, testCase.n, testCase.GetGamesErr)
			}

			n, games, err := gameService.GetGames(ctx, testCase.limit, testCase.offset)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.n, n)
			assert.Len(t, games, len(testCase.games))

			for i, game := range games {
				assert.Equal(t, testCase.games[i].GetID(), game.GetID())
				assert.Equal(t, testCase.games[i].GetName(), game.GetName())
				assert.Equal(t, testCase.games[i].GetDescription(), game.GetDescription())
				assert.WithinDuration(t, testCase.games[i].GetCreatedAt(), game.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetMyGames(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUser(mockUserAuth, mockUserCache)

	gameService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		userUtils,
	)

	type test struct {
		description           string
		authSession           *domain.OIDCSession
		user                  *service.UserInfo
		isGetMeErr            bool
		executeGetGamesByUser bool
		GetGamesByUserErr     error
		limit                 int
		offset                int
		n                     int
		games                 []*domain.Game
		GetGamesErr           error
		isErr                 bool
		err                   error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	user := service.NewUserInfo(
		values.NewTrapMemberID(uuid.New()),
		"ikura-hamu",
		values.TrapMemberStatusActive,
	)

	testCases := []test{
		{
			description: "getMeがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			isGetMeErr: true,
			isErr:      true,
		},
		{
			description: "特に問題ないのでエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game descriptiion",
					time.Now(),
				),
			},
			limit:  0,
			offset: 0,
			n:      1,
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			games:                 []*domain.Game{},
			limit:                 0,
			offset:                0,
			n:                     0,
		},
		{
			description: "ゲームが複数でもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
				domain.NewGame(
					gameID2,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:  0,
			offset: 0,
			n:      2,
		},
		{
			description: "limitが設定されてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:  1,
			offset: 0,
			n:      1,
		},
		{
			description: "limitとoffsetが両方設定されてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			limit:  1,
			offset: 1,
			n:      1,
		},
		{
			description: "offsetだけが設定されているのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:   user,
			limit:  0,
			offset: 1,
			isErr:  true,
			err:    service.ErrOffsetWithoutLimit,
		},
		{
			description: "GetGamesByUserがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user:                  user,
			executeGetGamesByUser: true,
			GetGamesByUserErr:     errors.New("error"),
			isErr:                 true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isGetMeErr {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(nil, cache.ErrCacheMiss)
				mockUserAuth.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession).
					Return(nil, errors.New("error"))
			} else {
				mockUserCache.
					EXPECT().
					GetMe(gomock.Any(), testCase.authSession.GetAccessToken()).
					Return(testCase.user, nil).
					AnyTimes()
			}

			if testCase.executeGetGamesByUser {
				mockGameRepository.
					EXPECT().
					GetGamesByUser(gomock.Any(), testCase.user.GetID(), testCase.limit, testCase.offset).
					Return(testCase.games, testCase.n, testCase.GetGamesByUserErr)
			}

			n, games, err := gameService.GetMyGames(ctx, testCase.authSession, testCase.limit, testCase.offset)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Len(t, games, len(testCase.games))
			assert.Equal(t, testCase.n, n)

			for i, game := range games {
				assert.Equal(t, testCase.games[i].GetID(), game.GetID())
				assert.Equal(t, testCase.games[i].GetName(), game.GetName())
				assert.Equal(t, testCase.games[i].GetDescription(), game.GetDescription())
				assert.WithinDuration(t, testCase.games[i].GetCreatedAt(), game.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestUpdateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUser(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		userUtils,
	)

	type test struct {
		description       string
		gameID            values.GameID
		name              values.GameName
		gameDescription   values.GameDescription
		game              *domain.Game
		GetGameErr        error
		executeUpdateGame bool
		UpdateGameErr     error
		isErr             bool
		err               error
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:     "特に問題ないのでエラーなし",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "nameの変更なしでもエラーなし",
			gameID:          gameID,
			name:            values.GameName("before"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "descriptionの変更なしでもエラーなし",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("before"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
		},
		{
			description:     "変更なしでも問題なし",
			gameID:          gameID,
			name:            values.GameName("before"),
			gameDescription: values.GameDescription("before"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
		},
		{
			description:     "ゲームが存在しないのでErrNoGame",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			GetGameErr:      repository.ErrRecordNotFound,
			isErr:           true,
			err:             service.ErrNoGame,
		},
		{
			description:     "GetGameがエラーなのでエラー",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			GetGameErr:      errors.New("error"),
			isErr:           true,
		},
		{
			description:     "UpdateGameがErrNoRecordUpdatedなのでエラー",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
			UpdateGameErr:     repository.ErrNoRecordUpdated,
			isErr:             true,
		},
		{
			description:     "UpdateGameがエラーなのでエラー",
			gameID:          gameID,
			name:            values.GameName("after"),
			gameDescription: values.GameDescription("after"),
			game: domain.NewGame(
				gameID,
				values.GameName("before"),
				values.GameDescription("before"),
				time.Now(),
			),
			executeUpdateGame: true,
			UpdateGameErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(testCase.game, testCase.GetGameErr)

			if testCase.executeUpdateGame {
				mockGameRepository.
					EXPECT().
					UpdateGame(gomock.Any(), testCase.game).
					Return(testCase.UpdateGameErr)
			}

			game, err := gameVersionService.UpdateGame(ctx, testCase.gameID, testCase.name, testCase.gameDescription)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.game, game)

			assert.Equal(t, testCase.game.GetID(), game.GetID())
			assert.Equal(t, testCase.name, game.GetName())
			assert.Equal(t, testCase.gameDescription, game.GetDescription())
			assert.Equal(t, testCase.game.GetCreatedAt(), game.GetCreatedAt())
		})
	}
}

func TestDeleteGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameManagementRoleRepository := mockRepository.NewMockGameManagementRole(ctrl)
	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUser(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(
		mockDB,
		mockGameRepository,
		mockGameManagementRoleRepository,
		mockGameGenreRepository,
		userUtils,
	)

	type test struct {
		description   string
		gameID        values.GameID
		RemoveGameErr error
		isErr         bool
		err           error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      values.NewGameID(),
		},
		{
			description:   "ゲームが存在しないのでErrNoGame",
			gameID:        values.NewGameID(),
			RemoveGameErr: repository.ErrNoRecordDeleted,
			isErr:         true,
			err:           service.ErrNoGame,
		},
		{
			description:   "RemoveGameがエラーなのでエラー",
			gameID:        values.NewGameID(),
			RemoveGameErr: errors.New("error"),
			isErr:         true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				RemoveGame(gomock.Any(), testCase.gameID).
				Return(testCase.RemoveGameErr)

			err := gameVersionService.DeleteGame(ctx, testCase.gameID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
