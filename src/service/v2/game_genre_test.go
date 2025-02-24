package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGetGameGenres(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	type test struct {
		isLoginUser      bool
		gameInfosRepo    []*repository.GameGenreInfo
		GetGameGenresErr error
		gameInfos        []*service.GameGenreInfo
		isErr            bool
		expectedErr      error
	}

	gameGenre1 := domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now())
	gameGenre2 := domain.NewGameGenre(values.NewGameGenreID(), "2D", time.Now())

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameInfosRepo: []*repository.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
			gameInfos:     []*service.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
		},
		"ログインしていてもエラー無し": {
			isLoginUser:   true,
			gameInfosRepo: []*repository.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
			gameInfos:     []*service.GameGenreInfo{{GameGenre: *gameGenre1, Num: 1}},
		},
		"複数でもエラー無し": {
			gameInfosRepo: []*repository.GameGenreInfo{
				{GameGenre: *gameGenre1, Num: 1},
				{GameGenre: *gameGenre2, Num: 3},
			},
			gameInfos: []*service.GameGenreInfo{
				{GameGenre: *gameGenre1, Num: 1},
				{GameGenre: *gameGenre2, Num: 3},
			},
		},
		"GetGameGenresがエラーなのでエラー": {
			GetGameGenresErr: errors.New("test"),
			isErr:            true,
		},
	}

	visibilitiesAll := []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate}
	visibilitiesNotLogin := []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			var argVisibilities []values.GameVisibility
			if testCase.isLoginUser {
				argVisibilities = visibilitiesAll
			} else {
				argVisibilities = visibilitiesNotLogin
			}
			mockGameGenreRepository.
				EXPECT().
				GetGameGenres(gomock.Any(), gomock.InAnyOrder(argVisibilities)).
				Return(testCase.gameInfosRepo, testCase.GetGameGenresErr)

			gameInfos, err := gameGenreService.GetGameGenres(ctx, testCase.isLoginUser)

			if testCase.isErr {
				if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			assert.Len(t, gameInfos, len(testCase.gameInfos))
			for i := range gameInfos {
				assert.Equal(t, testCase.gameInfos[i].GetID(), gameInfos[i].GetID())
				assert.Equal(t, testCase.gameInfos[i].GetName(), gameInfos[i].GetName())
				assert.Equal(t, testCase.gameInfos[i].Num, gameInfos[i].Num)
				assert.WithinDuration(t, testCase.gameInfos[i].GetCreatedAt(), gameInfos[i].GetCreatedAt(), time.Second)
			}

		})
	}
}

func TestDeleteGameGenre(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	type test struct {
		ID                 values.GameGenreID
		RemoveGameGenreErr error
		isErr              bool
		expectedErr        error
	}

	genreID := values.NewGameGenreID()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			ID: genreID,
		},
		"RemoveGameGenreがErrNoRecordDeletedなのでエラー": {
			ID:                 genreID,
			RemoveGameGenreErr: repository.ErrNoRecordDeleted,
			isErr:              true,
			expectedErr:        service.ErrNoGameGenre,
		},
		"RemoveGameGenreが他のエラーなのでエラー": {
			ID:                 genreID,
			RemoveGameGenreErr: errors.New("error"),
			isErr:              true,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockGameGenreRepository.
				EXPECT().
				RemoveGameGenre(ctx, testCase.ID).
				Return(testCase.RemoveGameGenreErr)

			err := gameGenreService.DeleteGameGenre(ctx, genreID)

			if !testCase.isErr {
				assert.NoError(t, err)
				return
			}

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestUpdateGameGenres(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	type test struct {
		gameID                        values.GameID
		gameGenreNames                []values.GameGenreName
		executeGetGameGenresWithNames bool
		GetGameGenresWithNamesResult  []*domain.GameGenre
		GetGameGenresWithNamesErr     error
		executeSaveGameGenres         bool
		SaveGameGenresErr             error
		executeRegisterGenresToGame   bool
		RegisterGenresToGameErr       error
		isErr                         bool
		expectedErr                   error
	}

	gameGenreName1 := values.GameGenreName("3D")
	gameGenreName2 := values.GameGenreName("2D")

	gameGenre1 := domain.NewGameGenre(values.NewGameGenreID(), gameGenreName1, time.Now())
	gameGenre2 := domain.NewGameGenre(values.NewGameGenreID(), gameGenreName2, time.Now())

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1},
			executeSaveGameGenres:         true,
			executeRegisterGenresToGame:   true,
		},
		"ジャンル名が重複しているのでエラー": {
			gameID:         values.NewGameID(),
			gameGenreNames: []values.GameGenreName{gameGenreName1, gameGenreName1},
			isErr:          true,
			expectedErr:    service.ErrDuplicateGameGenre,
		},
		"GetGameGenresWithNamesがErrRecordNotFoundでもエラー無し": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{},
			GetGameGenresWithNamesErr:     repository.ErrRecordNotFound,
			executeSaveGameGenres:         true,
			executeRegisterGenresToGame:   true,
		},
		"GetGameGenresWithNamesがErrRecordNotFound以外のエラーなのでエラー": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{},
			GetGameGenresWithNamesErr:     errors.New("test"),
			isErr:                         true,
		},
		"全てが既存のジャンルでもエラー無し": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1, gameGenre2},
			executeRegisterGenresToGame:   true,
		},
		"SaveGameGenresがErrDuplicatedUniqueKeyなのでエラー": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1},
			executeSaveGameGenres:         true,
			SaveGameGenresErr:             repository.ErrDuplicatedUniqueKey,
			isErr:                         true,
			expectedErr:                   service.ErrDuplicateGameGenre,
		},
		"SaveGameGenresが他のエラーなのでエラー": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1},
			executeSaveGameGenres:         true,
			SaveGameGenresErr:             errors.New("test"),
			isErr:                         true,
		},
		"RegisterGenresToGameがErrRecordNotFoundなのでエラー": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1},
			executeSaveGameGenres:         true,
			executeRegisterGenresToGame:   true,
			RegisterGenresToGameErr:       repository.ErrRecordNotFound,
			isErr:                         true,
			expectedErr:                   service.ErrNoGame,
		},
		"RegisterGenresToGameが他のエラーなのでエラー": {
			gameID:                        values.NewGameID(),
			gameGenreNames:                []values.GameGenreName{gameGenreName1, gameGenreName2},
			executeGetGameGenresWithNames: true,
			GetGameGenresWithNamesResult:  []*domain.GameGenre{gameGenre1},
			executeSaveGameGenres:         true,
			executeRegisterGenresToGame:   true,
			RegisterGenresToGameErr:       errors.New("test"),
			isErr:                         true,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			if testCase.executeGetGameGenresWithNames {
				mockGameGenreRepository.
					EXPECT().
					GetGameGenresWithNames(ctx, testCase.gameGenreNames).
					Return(testCase.GetGameGenresWithNamesResult, testCase.GetGameGenresWithNamesErr)
			}

			if testCase.executeSaveGameGenres {
				mockGameGenreRepository.
					EXPECT().
					SaveGameGenres(ctx, gomock.Len(len(testCase.gameGenreNames)-len(testCase.GetGameGenresWithNamesResult))).
					Return(testCase.SaveGameGenresErr)
			}

			if testCase.executeRegisterGenresToGame {
				mockGameGenreRepository.
					EXPECT().
					RegisterGenresToGame(ctx, testCase.gameID, gomock.Len(len(testCase.gameGenreNames))).
					Return(testCase.RegisterGenresToGameErr)
			}

			err := gameGenreService.UpdateGameGenres(ctx, testCase.gameID, testCase.gameGenreNames)

			if !testCase.isErr {
				assert.NoError(t, err)
				return
			}

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestUpdateGameGenre(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreRepository := mockRepository.NewMockGameGenre(ctrl)
	mockDB := mockRepository.NewMockDB(ctrl)

	gameGenreService := NewGameGenre(mockDB, mockGameGenreRepository)

	gameGenreID := values.NewGameGenreID()

	testCases := map[string]struct {
		gameGenre              *domain.GameGenre
		newGameGenreName       values.GameGenreName
		GetGameGenreErr        error
		executeUpdateGameGenre bool
		UpdateGameGenreErr     error
		executeGetGenreGames   bool
		games                  []*domain.Game
		GetGenreGamesErr       error
		genreInfo              *service.GameGenreInfo
		isError                bool
		wantErr                error
	}{
		"特に問題ないのでエラー無し": {
			gameGenre:              domain.NewGameGenre(gameGenreID, "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			executeGetGenreGames:   true,
			games: []*domain.Game{
				domain.NewGame(values.NewGameID(), "game", "description", values.GameVisibilityTypePublic, time.Now()),
			},
			genreInfo: &service.GameGenreInfo{
				GameGenre: *domain.NewGameGenre(gameGenreID, "2D", time.Now()),
				Num:       1,
			},
		},
		"GetGameGenreがErrRecordNotFoundなのでErrNoGameGenre": {
			gameGenre:        domain.NewGameGenre(gameGenreID, "3D", time.Now()),
			newGameGenreName: "2D",
			GetGameGenreErr:  repository.ErrRecordNotFound,
			isError:          true,
			wantErr:          service.ErrNoGameGenre,
		},
		"GetGameGenreがエラーなのでエラー": {
			gameGenre:        domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now()),
			newGameGenreName: "2D",
			GetGameGenreErr:  errors.New("test"),
			isError:          true,
		},
		"値に変更が無いのでErrNoGameGenreUpdated": {
			gameGenre:        domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now()),
			newGameGenreName: values.GameGenreName("3D"),
			isError:          true,
			wantErr:          service.ErrNoGameGenreUpdated,
		},
		"UpdateGameGenreがErrDuplicatedUniqueKeyなのでErrDuplicateGameGenre": {
			gameGenre:              domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			UpdateGameGenreErr:     repository.ErrDuplicatedUniqueKey,
			isError:                true,
			wantErr:                service.ErrDuplicateGameGenreName,
		},
		"UpdateGameGenreがErrNoRecordUpdatedなのでErrNoGameGenreUpdated": {
			gameGenre:              domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now()),
			newGameGenreName:       "2D", // 本来はこの値が異なるからErrNoRecordUpdatedにはならない
			executeUpdateGameGenre: true,
			UpdateGameGenreErr:     repository.ErrNoRecordUpdated,
			isError:                true,
			wantErr:                service.ErrNoGameGenreUpdated,
		},
		"UpdateGameGenreがエラーなのでエラー": {
			gameGenre:              domain.NewGameGenre(values.NewGameGenreID(), "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			UpdateGameGenreErr:     errors.New("test"),
			isError:                true,
		},
		"GetGenreGamesがnilなので0件でエラー無し": {
			gameGenre:              domain.NewGameGenre(gameGenreID, "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			executeGetGenreGames:   true,
			games:                  nil,
			genreInfo: &service.GameGenreInfo{
				GameGenre: *domain.NewGameGenre(gameGenreID, "2D", time.Now()),
				Num:       0,
			},
		},
		"GetGenreGamesが複数でもエラー無し": {
			gameGenre:              domain.NewGameGenre(gameGenreID, "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			executeGetGenreGames:   true,
			games: []*domain.Game{
				domain.NewGame(values.NewGameID(), "game1", "description1", values.GameVisibilityTypePublic, time.Now()),
				domain.NewGame(values.NewGameID(), "game2", "description2", values.GameVisibilityTypePublic, time.Now()),
			},
			genreInfo: &service.GameGenreInfo{
				GameGenre: *domain.NewGameGenre(gameGenreID, "2D", time.Now()),
				Num:       2,
			},
		},
		"GetGenreGamesがエラーなのでエラー": {
			gameGenre:              domain.NewGameGenre(gameGenreID, "3D", time.Now()),
			newGameGenreName:       "2D",
			executeUpdateGameGenre: true,
			executeGetGenreGames:   true,
			GetGenreGamesErr:       errors.New("test"),
			isError:                true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mockGameGenreRepository.
				EXPECT().
				GetGameGenre(gomock.Any(), testCase.gameGenre.GetID()).
				Return(testCase.gameGenre, testCase.GetGameGenreErr)

			if testCase.executeUpdateGameGenre {
				mockGameGenreRepository.
					EXPECT().
					UpdateGameGenre(gomock.Any(), domain.NewGameGenre(testCase.gameGenre.GetID(), testCase.newGameGenreName, testCase.gameGenre.GetCreatedAt())).
					Return(testCase.UpdateGameGenreErr)
			}

			if testCase.executeGetGenreGames {
				mockGameGenreRepository.
					EXPECT().
					GetGamesByGenreID(gomock.Any(), testCase.gameGenre.GetID()).
					Return(testCase.games, testCase.GetGenreGamesErr)
			}

			genreInfo, err := gameGenreService.UpdateGameGenre(ctx, testCase.gameGenre.GetID(), testCase.newGameGenreName)

			if testCase.isError {
				if testCase.wantErr != nil {
					assert.ErrorIs(t, err, testCase.wantErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil {
				return
			}

			assert.Equal(t, testCase.genreInfo.GetID(), genreInfo.GetID())
			assert.Equal(t, testCase.genreInfo.GetName(), genreInfo.GetName())
			assert.WithinDuration(t, testCase.genreInfo.GetCreatedAt(), genreInfo.GetCreatedAt(), time.Second)
			assert.Equal(t, testCase.genreInfo.Num, genreInfo.Num)
		})
	}
}
