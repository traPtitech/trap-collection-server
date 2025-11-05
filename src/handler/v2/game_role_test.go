package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestPatchGameRole(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	game := domain.NewGame(values.NewGameID(), "game1", "game1 description", values.GameVisibilityTypePrivate, time.Now())
	user := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), "user1", values.TrapMemberStatusActive, false)
	user2 := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), "user2", values.TrapMemberStatusActive, false)
	genre := domain.NewGameGenre(values.NewGameGenreID(), "genre1", time.Now())

	var (
		roleTypeOwner                           = openapi.Owner
		roleTypeMaintainer                      = openapi.Maintainer
		roleTypeInvalid    openapi.GameRoleType = "invalid"
	)

	ownerRequestBody := &openapi.PatchGameRoleJSONRequestBody{
		Id:   openapi.UserID(user.GetID()),
		Type: &roleTypeOwner,
	}

	validAuthSession := domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))

	testCases := map[string]struct {
		gameID                        openapi.GameIDInPath
		sessionExist                  bool
		authSession                   *domain.OIDCSession
		reqBody                       *openapi.PatchGameRoleJSONRequestBody
		executeEditGameManagementRole bool
		EditGameManagementRoleErr     error
		executeGetGame                bool
		newGameInfo                   *service.GameInfoV2
		GetGameErr                    error
		isErr                         bool
		err                           error
		statusCode                    int
		expectedResponse              *openapi.Game
	}{
		"ownerでも問題なし": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:   game,
				Owners: []*service.UserInfo{user},
				Genres: []*domain.GameGenre{genre},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre.GetName())},
			},
		},
		"maintainerでも問題なし": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			sessionExist: true,
			authSession:  validAuthSession,
			reqBody: &openapi.PatchGameRoleJSONRequestBody{
				Id:   openapi.UserID(user.GetID()),
				Type: &roleTypeMaintainer,
			},
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:        game,
				Owners:      []*service.UserInfo{user2},
				Maintainers: []*service.UserInfo{user},
				Genres:      []*domain.GameGenre{genre},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user2.GetName())},
				Maintainers: &[]openapi.UserName{string(user.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre.GetName())},
			},
		},
		"複数人いても問題なし": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:        game,
				Owners:      []*service.UserInfo{user, user2},
				Maintainers: []*service.UserInfo{},
				Genres:      []*domain.GameGenre{genre},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user.GetName()), string(user2.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre.GetName())},
			},
		},
		"複数ジャンルでも問題なし": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:        game,
				Owners:      []*service.UserInfo{user},
				Maintainers: []*service.UserInfo{},
				Genres:      []*domain.GameGenre{genre, domain.NewGameGenre(values.NewGameGenreID(), "genre2", time.Now())},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre.GetName()), "genre2"},
			},
		},
		"roleTypeがnilでも問題なし": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			sessionExist: true,
			authSession:  validAuthSession,
			reqBody: &openapi.PatchGameRoleJSONRequestBody{
				Id: openapi.UserID(user.GetID()),
			},
			executeEditGameManagementRole: false,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:        game,
				Owners:      []*service.UserInfo{user},
				Maintainers: []*service.UserInfo{},
				Genres:      []*domain.GameGenre{genre},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre.GetName())},
			},
		},
		"sessionがないので401": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			sessionExist: false,
			statusCode:   http.StatusUnauthorized,
			isErr:        true,
		},
		"authSessionが無効なので401": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			sessionExist: true,
			authSession:  nil,
			statusCode:   http.StatusUnauthorized,
			isErr:        true,
		},
		"無効なroleTypeなので400": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			sessionExist: true,
			authSession:  validAuthSession,
			reqBody: &openapi.PatchGameRoleJSONRequestBody{
				Id:   openapi.UserID(user.GetID()),
				Type: &roleTypeInvalid,
			},
			statusCode: http.StatusBadRequest,
			isErr:      true,
		},
		"EditGameManagementRoleがErrNoGameManagementRoleUpdatedなので400": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			EditGameManagementRoleErr:     service.ErrNoGameManagementRoleUpdated,
			statusCode:                    http.StatusBadRequest,
			isErr:                         true,
		},
		"EditGameManagementRoleがErrNoGameなので404": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			EditGameManagementRoleErr:     service.ErrNoGame,
			statusCode:                    http.StatusNotFound,
			isErr:                         true,
		},
		"EditGameManagementRoleがErrInvalidUserIDなので400": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			EditGameManagementRoleErr:     service.ErrInvalidUserID,
			statusCode:                    http.StatusBadRequest,
			isErr:                         true,
		},
		"EditGameManagementRoleがErrCannotEditOwnersなので400": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			EditGameManagementRoleErr:     service.ErrCannotEditOwners,
			statusCode:                    http.StatusBadRequest,
			isErr:                         true,
		},
		"EditGameManagementRoleがエラーなので500": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			EditGameManagementRoleErr:     errors.New("error"),
			statusCode:                    http.StatusInternalServerError,
			isErr:                         true,
		},
		"GetGameがErrNoGameなので404": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			GetGameErr:                    service.ErrNoGame,
			statusCode:                    http.StatusNotFound,
			isErr:                         true,
		},
		"GetGameがエラーなので500": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			GetGameErr:                    errors.New("error"),
			statusCode:                    http.StatusInternalServerError,
			isErr:                         true,
		},
		"GetGameのvisibilityが無効なので500": {
			gameID:                        openapi.GameIDInPath(game.GetID()),
			sessionExist:                  true,
			authSession:                   validAuthSession,
			reqBody:                       ownerRequestBody,
			executeEditGameManagementRole: true,
			executeGetGame:                true,
			newGameInfo: &service.GameInfoV2{
				Game:   domain.NewGame(game.GetID(), game.GetName(), game.GetDescription(), values.GameVisibility(100), game.GetCreatedAt()),
				Owners: []*service.UserInfo{user},
				Genres: []*domain.GameGenre{genre},
			},
			statusCode: http.StatusInternalServerError,
			isErr:      true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockConf := mockConfig.NewMockHandler(ctrl)
			mockConf.
				EXPECT().
				SessionKey().
				Return("key", nil)
			mockConf.
				EXPECT().
				SessionSecret().
				Return("secret", nil)
			sess, err := session.NewSession(mockConf)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
				return
			}
			session, err := NewSession(sess)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
				return
			}

			c, req, rec := setupTestRequest(
				t,
				http.MethodPatch,
				fmt.Sprintf("/api/v2/games/%s/role", testCase.gameID.String()),
				withJSONBody(t, testCase.reqBody),
			)

			if testCase.sessionExist {
				setTestSession(t, c, req, rec, session, testCase.authSession)
			}

			mockGameRoleService := mock.NewMockGameRoleV2(ctrl)
			if testCase.executeEditGameManagementRole {
				var roleType values.GameManagementRole
				switch *testCase.reqBody.Type {
				case openapi.Owner:
					roleType = values.GameManagementRoleAdministrator
				case openapi.Maintainer:
					roleType = values.GameManagementRoleCollaborator
				default:
					t.Fatalf("unexpected role type: %v", *testCase.reqBody.Type)
				}
				mockGameRoleService.
					EXPECT().
					EditGameManagementRole(gomock.Any(), gomock.Any(), values.GameID(testCase.gameID), values.NewTrapMemberID(testCase.reqBody.Id), roleType).
					Return(testCase.EditGameManagementRoleErr)
			}

			mockGameService := mock.NewMockGameV2(ctrl)
			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any(), values.GameID(testCase.gameID)).
					Return(testCase.newGameInfo, testCase.GetGameErr)
			}

			gameRole := NewGameRole(mockGameRoleService, mockGameService, session)

			err = gameRole.PatchGameRole(c, testCase.gameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					}
				} else if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if testCase.isErr {
				return
			}

			var responseGame openapi.Game
			err = json.NewDecoder(rec.Body).Decode(&responseGame)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedResponse.Id, responseGame.Id)
			assert.Equal(t, testCase.expectedResponse.Name, responseGame.Name)
			assert.Equal(t, testCase.expectedResponse.Description, responseGame.Description)
			assert.Equal(t, testCase.expectedResponse.Visibility, responseGame.Visibility)
			for i := range testCase.expectedResponse.Owners {
				assert.Equal(t, testCase.expectedResponse.Owners[i], responseGame.Owners[i])
			}
			if testCase.expectedResponse.Maintainers != nil {
				for i := range *testCase.expectedResponse.Maintainers {
					assert.Equal(t, (*testCase.expectedResponse.Maintainers)[i], (*responseGame.Maintainers)[i])
				}
			}
			assert.WithinDuration(t, testCase.expectedResponse.CreatedAt, responseGame.CreatedAt, time.Second)
			if testCase.expectedResponse.Genres != nil {
				for i := range *testCase.expectedResponse.Genres {
					assert.Equal(t, (*testCase.expectedResponse.Genres)[i], (*responseGame.Genres)[i])
				}
			}
		})
	}
}

func TestDeleteGameRole(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	game := domain.NewGame(values.NewGameID(), "game1", "game1 description", values.GameVisibilityTypePrivate, time.Now())
	user := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), "user1", values.TrapMemberStatusActive, false)
	user2 := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), "user2", values.TrapMemberStatusActive, false)
	genre1 := domain.NewGameGenre(values.NewGameGenreID(), "genre1", time.Now())

	validAuthSession := domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))

	testCases := map[string]struct {
		gameID                          openapi.GameIDInPath
		userID                          openapi.UserIDInPath
		sessionExist                    bool
		authSession                     *domain.OIDCSession
		executeDeleteGameManagementRole bool
		DeleteGameManagementRoleErr     error
		executeGetGame                  bool
		newGameInfo                     *service.GameInfoV2
		GetGameErr                      error
		isErr                           bool
		err                             error
		statusCode                      int
		expectedResponse                *openapi.Game
	}{
		"特に問題なく削除できる": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			executeGetGame:                  true,
			newGameInfo: &service.GameInfoV2{
				Game:        game,
				Owners:      []*service.UserInfo{user},
				Maintainers: []*service.UserInfo{user2},
				Genres:      []*domain.GameGenre{genre1},
			},
			expectedResponse: &openapi.Game{
				Id:          openapi.GameID(game.GetID()),
				Name:        openapi.GameName(game.GetName()),
				Description: openapi.GameDescription(game.GetDescription()),
				Visibility:  openapi.Private,
				Owners:      []openapi.UserName{string(user.GetName())},
				CreatedAt:   game.GetCreatedAt(),
				Genres:      &[]openapi.GameGenreName{string(genre1.GetName())},
			},
		},
		"sessionがないので401": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			userID:       openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist: false,
			statusCode:   http.StatusUnauthorized,
			isErr:        true,
		},
		"authSessionが無効なので401": {
			gameID:       openapi.GameIDInPath(game.GetID()),
			userID:       openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist: true,
			authSession:  nil,
			statusCode:   http.StatusUnauthorized,
			isErr:        true,
		},
		"RemoveGameManagementRoleがErrInvalidRoleなので404": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			DeleteGameManagementRoleErr:     service.ErrInvalidRole,
			statusCode:                      http.StatusNotFound,
			isErr:                           true,
		},
		"RemoveGameManagementRoleがErrCannotDeleteOwnerなので400": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(user.GetID()),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			DeleteGameManagementRoleErr:     service.ErrCannotDeleteOwner,
			statusCode:                      http.StatusBadRequest,
			isErr:                           true,
		},
		"RemoveGameManagementRoleがErrNoGameなので400": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			DeleteGameManagementRoleErr:     service.ErrNoGame,
			statusCode:                      http.StatusNotFound,
			isErr:                           true,
		},
		"RemoveGameManagementRoleがエラーなので500": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			DeleteGameManagementRoleErr:     errors.New("error"),
			statusCode:                      http.StatusInternalServerError,
			isErr:                           true,
		},
		"GetGameがErrNoGameなので404": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			executeGetGame:                  true,
			GetGameErr:                      service.ErrNoGame,
			statusCode:                      http.StatusNotFound,
			isErr:                           true,
		},
		"GetGameがエラーなので500": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			executeGetGame:                  true,
			GetGameErr:                      errors.New("error"),
			statusCode:                      http.StatusInternalServerError,
			isErr:                           true,
		},
		"GetGameのvisibilityが無効なので500": {
			gameID:                          openapi.GameIDInPath(game.GetID()),
			userID:                          openapi.UserIDInPath(values.NewTrapMemberID(uuid.New())),
			sessionExist:                    true,
			authSession:                     validAuthSession,
			executeDeleteGameManagementRole: true,
			executeGetGame:                  true,
			newGameInfo: &service.GameInfoV2{
				Game:        domain.NewGame(game.GetID(), game.GetName(), game.GetDescription(), values.GameVisibility(100), game.GetCreatedAt()),
				Owners:      []*service.UserInfo{user},
				Maintainers: []*service.UserInfo{user2},
				Genres:      []*domain.GameGenre{genre1},
			},
			statusCode: http.StatusInternalServerError,
			isErr:      true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockConf := mockConfig.NewMockHandler(ctrl)
			mockConf.
				EXPECT().
				SessionKey().
				Return("key", nil)
			mockConf.
				EXPECT().
				SessionSecret().
				Return("secret", nil)
			sess, err := session.NewSession(mockConf)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
				return
			}
			session, err := NewSession(sess)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
				return
			}

			c, req, rec := setupTestRequest(
				t,
				http.MethodDelete,
				fmt.Sprintf("/api/v2/games/%s/role/%s", testCase.gameID.String(), testCase.userID.String()),
				nil,
			)

			if testCase.sessionExist {
				setTestSession(t, c, req, rec, session, testCase.authSession)
			}

			mockGameRoleService := mock.NewMockGameRoleV2(ctrl)
			if testCase.executeDeleteGameManagementRole {
				mockGameRoleService.
					EXPECT().
					RemoveGameManagementRole(gomock.Any(), values.GameID(testCase.gameID), values.NewTrapMemberID(testCase.userID)).
					Return(testCase.DeleteGameManagementRoleErr)
			}

			mockGameService := mock.NewMockGameV2(ctrl)
			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any(), values.GameID(testCase.gameID)).
					Return(testCase.newGameInfo, testCase.GetGameErr)
			}

			gameRole := NewGameRole(mockGameRoleService, mockGameService, session)

			err = gameRole.DeleteGameRole(c, testCase.gameID, testCase.userID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if assert.ErrorAs(t, err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					}
				} else if testCase.err != nil {
					assert.ErrorIs(t, err, testCase.err)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if testCase.isErr {
				return
			}

			var responseGame openapi.Game
			err = json.NewDecoder(rec.Body).Decode(&responseGame)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedResponse.Id, responseGame.Id)
			assert.Equal(t, testCase.expectedResponse.Name, responseGame.Name)
			assert.Equal(t, testCase.expectedResponse.Description, responseGame.Description)
			assert.Equal(t, testCase.expectedResponse.Visibility, responseGame.Visibility)
			for i := range testCase.expectedResponse.Owners {
				assert.Equal(t, testCase.expectedResponse.Owners[i], responseGame.Owners[i])
			}
			if testCase.expectedResponse.Maintainers != nil {
				for i := range *testCase.expectedResponse.Maintainers {
					assert.Equal(t, (*testCase.expectedResponse.Maintainers)[i], (*responseGame.Maintainers)[i])
				}
			}
			assert.WithinDuration(t, testCase.expectedResponse.CreatedAt, responseGame.CreatedAt, time.Second)
			if testCase.expectedResponse.Genres != nil {
				for i := range *testCase.expectedResponse.Genres {
					assert.Equal(t, (*testCase.expectedResponse.Genres)[i], (*responseGame.Genres)[i])
				}
			}
		})
	}
}
