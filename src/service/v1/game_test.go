package v1

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
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description     string
		name            values.GameName
		gameDescription values.GameDescription
		SaveGameErr     error
		isErr           bool
		err             error
	}

	testCases := []test{
		{
			description:     "特に問題ないのでエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "nameが空でもエラーなし",
			name:            values.GameName(""),
			gameDescription: values.GameDescription("test"),
		},
		{
			description:     "descriptionが空でもエラーなし",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription(""),
		},
		{
			description:     "CreateGameがエラーなのでエラー",
			name:            values.GameName("test"),
			gameDescription: values.GameDescription("test"),
			SaveGameErr:     errors.New("test"),
			isErr:           true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				SaveGame(gomock.Any(), gomock.Any()).
				Return(testCase.SaveGameErr)

			game, err := gameVersionService.CreateGame(ctx, testCase.name, testCase.gameDescription)

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

			assert.Equal(t, testCase.name, game.GetName())
			assert.Equal(t, testCase.gameDescription, game.GetDescription())
			assert.WithinDuration(t, time.Now(), game.GetCreatedAt(), time.Second)
		})
	}
}

func TestUpdateGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

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
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

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

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description                 string
		gameID                      values.GameID
		game                        *domain.Game
		GetGameErr                  error
		executeGetLatestGameVersion bool
		gameVersion                 *domain.GameVersion
		GetLatestGameVersionErr     error
		isErr                       bool
		err                         error
	}

	gameID := values.NewGameID()

	gameVersionID := values.NewGameVersionID()

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
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				gameVersionID,
				"v1.0.0",
				"game version description",
				time.Now(),
			),
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
			description: "ゲームバージョンが存在しなくても問題なし",
			gameID:      gameID,
			game: domain.NewGame(
				gameID,
				"game name",
				"game description",
				time.Now(),
			),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     repository.ErrRecordNotFound,
		},
		{
			description: "GetLatestGameVersionがエラーなのでエラー",
			gameID:      gameID,
			game: domain.NewGame(
				gameID,
				"game name",
				"game description",
				time.Now(),
			),
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(testCase.game, testCase.GetGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.gameVersion, testCase.GetLatestGameVersionErr)
			}

			gameInfo, err := gameVersionService.GetGame(ctx, testCase.gameID)

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

			if testCase.gameVersion != nil {
				assert.Equal(t, testCase.gameVersion, gameInfo.LatestVersion)

				assert.Equal(t, testCase.gameVersion.GetID(), gameInfo.LatestVersion.GetID())
				assert.Equal(t, testCase.gameVersion.GetName(), gameInfo.LatestVersion.GetName())
				assert.Equal(t, testCase.gameVersion.GetDescription(), gameInfo.LatestVersion.GetDescription())
				assert.WithinDuration(t, testCase.gameVersion.GetCreatedAt(), gameInfo.LatestVersion.GetCreatedAt(), time.Second)
			} else {
				assert.Nil(t, gameInfo.LatestVersion)
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
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description                  string
		games                        []*domain.Game
		GetGamesErr                  error
		executeGetLatestGameVersions bool
		gameVersionMap               map[values.GameID]*domain.GameVersion
		GetLatestGameVersionsErr     error
		gameInfos                    []*service.GameInfo
		isErr                        bool
		err                          error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			gameVersionMap: map[values.GameID]*domain.GameVersion{
				gameID1: domain.NewGameVersion(
					gameVersionID1,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
			},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
			},
		},
		{
			description: "バージョンが存在しないゲームがあっても問題なし",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			gameVersionMap:               map[values.GameID]*domain.GameVersion{},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: nil,
				},
			},
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			games:       []*domain.Game{},
			gameInfos:   []*service.GameInfo{},
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
			executeGetLatestGameVersions: true,
			gameVersionMap: map[values.GameID]*domain.GameVersion{
				gameID1: domain.NewGameVersion(
					gameVersionID1,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
				gameID2: domain.NewGameVersion(
					gameVersionID2,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
			},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
			},
		},
		{
			description: "GetGamesがエラーなのでエラー",
			GetGamesErr: errors.New("error"),
			isErr:       true,
		},
		{
			description: "GetLatestGameVersionsByGameIDsがエラーなのでエラー",
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			GetLatestGameVersionsErr:     errors.New("error"),
			isErr:                        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGames(gomock.Any()).
				Return(testCase.games, testCase.GetGamesErr)

			if testCase.executeGetLatestGameVersions {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersionsByGameIDs(gomock.Any(), gomock.Any(), repository.LockTypeNone).
					Return(testCase.gameVersionMap, testCase.GetLatestGameVersionsErr)
			}

			gameInfos, err := gameVersionService.GetGames(ctx)

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

			assert.Len(t, gameInfos, len(testCase.gameInfos))

			for i, gameInfo := range gameInfos {
				assert.Equal(t, testCase.gameInfos[i].Game.GetID(), gameInfo.Game.GetID())
				assert.Equal(t, testCase.gameInfos[i].Game.GetName(), gameInfo.Game.GetName())
				assert.Equal(t, testCase.gameInfos[i].Game.GetDescription(), gameInfo.Game.GetDescription())
				assert.WithinDuration(t, testCase.gameInfos[i].Game.GetCreatedAt(), gameInfo.Game.GetCreatedAt(), time.Second)

				if testCase.gameInfos[i].LatestVersion == nil {
					assert.Nil(t, gameInfo.LatestVersion)
				} else {
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetID(), gameInfo.LatestVersion.GetID())
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetName(), gameInfo.LatestVersion.GetName())
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetDescription(), gameInfo.LatestVersion.GetDescription())
					assert.WithinDuration(t, testCase.gameInfos[i].LatestVersion.GetCreatedAt(), gameInfo.LatestVersion.GetCreatedAt(), time.Second)
				}
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
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	mockUserCache := mockCache.NewMockUser(ctrl)
	mockUserAuth := mockAuth.NewMockUser(ctrl)

	userUtils := NewUserUtils(mockUserAuth, mockUserCache)

	gameVersionService := NewGame(mockDB, mockGameRepository, mockGameVersionRepository, userUtils)

	type test struct {
		description                  string
		authSession                  *domain.OIDCSession
		user                         *service.UserInfo
		isGetMeErr                   bool
		executeGetGamesByUser        bool
		games                        []*domain.Game
		GetGamesByUserErr            error
		executeGetLatestGameVersions bool
		gameVersionMap               map[values.GameID]*domain.GameVersion
		GetLatestGameVersionsErr     error
		gameInfos                    []*service.GameInfo
		isErr                        bool
		err                          error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			gameVersionMap: map[values.GameID]*domain.GameVersion{
				gameID1: domain.NewGameVersion(
					gameVersionID1,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
			},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
			},
		},
		{
			description: "バージョンが存在しないゲームがあっても問題なし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			gameVersionMap:               map[values.GameID]*domain.GameVersion{},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: nil,
				},
			},
		},
		{
			description: "ゲームが存在しなくてもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			executeGetGamesByUser: true,
			games:                 []*domain.Game{},
			gameInfos:             []*service.GameInfo{},
		},
		{
			description: "ゲームが複数でもエラーなし",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
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
			executeGetLatestGameVersions: true,
			gameVersionMap: map[values.GameID]*domain.GameVersion{
				gameID1: domain.NewGameVersion(
					gameVersionID1,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
				gameID2: domain.NewGameVersion(
					gameVersionID2,
					"v1.0.0",
					"game version description",
					time.Now(),
				),
			},
			gameInfos: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						"game name",
						"game description",
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						"v1.0.0",
						"game version description",
						time.Now(),
					),
				},
			},
		},
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
			description: "GetGamesがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			executeGetGamesByUser: true,
			GetGamesByUserErr:     errors.New("error"),
			isErr:                 true,
		},
		{
			description: "GetLatestGameVersionsByGameIDsがエラーなのでエラー",
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			user: service.NewUserInfo(
				values.NewTrapMemberID(uuid.New()),
				"mazrean",
				values.TrapMemberStatusActive,
			),
			executeGetGamesByUser: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					"game name",
					"game description",
					time.Now(),
				),
			},
			executeGetLatestGameVersions: true,
			GetLatestGameVersionsErr:     errors.New("error"),
			isErr:                        true,
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
			}

			if testCase.executeGetGamesByUser {
				mockGameRepository.
					EXPECT().
					GetGamesByUser(gomock.Any(), testCase.user.GetID()).
					Return(testCase.games, testCase.GetGamesByUserErr)
			}

			if testCase.executeGetLatestGameVersions {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersionsByGameIDs(gomock.Any(), gomock.Any(), repository.LockTypeNone).
					Return(testCase.gameVersionMap, testCase.GetLatestGameVersionsErr)
			}

			gameInfos, err := gameVersionService.GetMyGames(ctx, testCase.authSession)

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

			assert.Len(t, gameInfos, len(testCase.gameInfos))

			for i, gameInfo := range gameInfos {
				assert.Equal(t, testCase.gameInfos[i].Game.GetID(), gameInfo.Game.GetID())
				assert.Equal(t, testCase.gameInfos[i].Game.GetName(), gameInfo.Game.GetName())
				assert.Equal(t, testCase.gameInfos[i].Game.GetDescription(), gameInfo.Game.GetDescription())
				assert.WithinDuration(t, testCase.gameInfos[i].Game.GetCreatedAt(), gameInfo.Game.GetCreatedAt(), time.Second)

				if testCase.gameInfos[i].LatestVersion == nil {
					assert.Nil(t, gameInfo.LatestVersion)
				} else {
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetID(), gameInfo.LatestVersion.GetID())
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetName(), gameInfo.LatestVersion.GetName())
					assert.Equal(t, testCase.gameInfos[i].LatestVersion.GetDescription(), gameInfo.LatestVersion.GetDescription())
					assert.WithinDuration(t, testCase.gameInfos[i].LatestVersion.GetCreatedAt(), gameInfo.LatestVersion.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}
