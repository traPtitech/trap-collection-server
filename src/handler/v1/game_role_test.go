package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/openapi"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostMaintainer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameRoleHandler := NewGameRole(session, mockGameAuthService)

	type test struct {
		description                 string
		sessionExist                bool
		authSession                 *domain.OIDCSession
		strGameID                   string
		maintainers                 []string
		executeAddGameCollaborators bool
		gameID                      values.GameID
		userIDs                     []values.TraPMemberID
		AddGameCollaboratorsErr     error
		isErr                       bool
		err                         error
		statusCode                  int
	}

	gameID := values.NewGameID()

	userIDs1 := []values.TraPMemberID{
		values.NewTrapMemberID(uuid.New()),
	}
	maintainers1 := make([]string, 0, len(userIDs1))
	for _, userID := range userIDs1 {
		maintainers1 = append(maintainers1, uuid.UUID(userID).String())
	}

	userIDs2 := []values.TraPMemberID{
		values.NewTrapMemberID(uuid.New()),
		values.NewTrapMemberID(uuid.New()),
	}
	maintainers2 := make([]string, 0, len(userIDs1))
	for _, userID := range userIDs2 {
		maintainers2 = append(maintainers2, uuid.UUID(userID).String())
	}

	testCases := []test{
		{
			description:  "特に問題ないので問題なし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:                   uuid.UUID(gameID).String(),
			maintainers:                 maintainers1,
			executeAddGameCollaborators: true,
			gameID:                      gameID,
			userIDs:                     userIDs1,
		},
		{
			description:  "セッションがないので500",
			sessionExist: false,
			strGameID:    uuid.UUID(gameID).String(),
			maintainers:  maintainers1,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionがないので400",
			sessionExist: true,
			strGameID:    uuid.UUID(gameID).String(),
			maintainers:  maintainers1,
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		{
			description:  "gameIDの形式がuuidでないので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:   "invalid",
			maintainers: maintainers1,
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:  "userIDの形式がuuidでないので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:   uuid.UUID(gameID).String(),
			maintainers: []string{"invalid"},
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:  "maintainerが複数でも問題なし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:                   uuid.UUID(gameID).String(),
			maintainers:                 maintainers2,
			executeAddGameCollaborators: true,
			gameID:                      gameID,
			userIDs:                     userIDs2,
		},
		{
			description:  "ErrInvalidGameIDなので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:                   uuid.UUID(gameID).String(),
			maintainers:                 maintainers1,
			executeAddGameCollaborators: true,
			gameID:                      gameID,
			userIDs:                     userIDs1,
			AddGameCollaboratorsErr:     service.ErrInvalidGameID,
			isErr:                       true,
			statusCode:                  http.StatusBadRequest,
		},
		{
			description:  "ErrInvalidUserIDなので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:                   uuid.UUID(gameID).String(),
			maintainers:                 maintainers1,
			executeAddGameCollaborators: true,
			gameID:                      gameID,
			userIDs:                     userIDs1,
			AddGameCollaboratorsErr:     service.ErrInvalidUserID,
			isErr:                       true,
			statusCode:                  http.StatusBadRequest,
		},
		{
			description:  "AddGameCollaboratorsがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:                   uuid.UUID(gameID).String(),
			maintainers:                 maintainers1,
			executeAddGameCollaborators: true,
			gameID:                      gameID,
			userIDs:                     userIDs1,
			AddGameCollaboratorsErr:     errors.New("AddGameCollaborators error"),
			isErr:                       true,
			statusCode:                  http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/game/maintainer", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.store.Get(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				c.Set(sessionContextKey, sess)
			}

			if testCase.executeAddGameCollaborators {
				mockGameAuthService.
					EXPECT().
					AddGameCollaborators(gomock.Any(), gomock.Any(), testCase.gameID, testCase.userIDs).
					Return(testCase.AddGameCollaboratorsErr)
			}

			err := gameRoleHandler.PostMaintainer(c, testCase.strGameID, &openapi.Maintainers{
				Maintainers: testCase.maintainers,
			})

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

func TestGetMaintainer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameRoleHandler := NewGameRole(session, mockGameAuthService)

	type test struct {
		description            string
		sessionExist           bool
		authSession            *domain.OIDCSession
		strGameID              string
		executeGetGameManagers bool
		gameID                 values.GameID
		gameManagers           []*service.GameManager
		GetGameManagersErr     error
		maintainers            []*openapi.Maintainer
		isErr                  bool
		err                    error
		statusCode             int
	}

	gameID := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
			},
			maintainers: []*openapi.Maintainer{
				{
					Id:   uuid.UUID(userID1).String(),
					Name: "mazrean",
					Role: 1,
				},
			},
		},
		{
			description:  "セッションがないので500",
			sessionExist: false,
			strGameID:    uuid.UUID(gameID).String(),
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionがないので400",
			sessionExist: true,
			strGameID:    uuid.UUID(gameID).String(),
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		{
			description:  "gameIDの形式がuuidでないので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:  "invalid",
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "ErrInvalidGameIDなので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			GetGameManagersErr:     service.ErrInvalidGameID,
			isErr:                  true,
			statusCode:             http.StatusBadRequest,
		},
		{
			description:  "GetGameManagersがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			GetGameManagersErr:     errors.New("error"),
			isErr:                  true,
			statusCode:             http.StatusInternalServerError,
		},
		{
			description:  "roleがcollaboratorでもエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleCollaborator,
				},
			},
			maintainers: []*openapi.Maintainer{
				{
					Id:   uuid.UUID(userID1).String(),
					Name: "mazrean",
					Role: 0,
				},
			},
		},
		{
			// 実際には起き得ないが、念のため確認
			description:  "roleが誤っているので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       100,
				},
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:  "管理者が複数でもエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			gameManagers: []*service.GameManager{
				{
					UserID:     userID1,
					UserName:   "mazrean",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleAdministrator,
				},
				{
					UserID:     userID2,
					UserName:   "mazrean2",
					UserStatus: values.TrapMemberStatusActive,
					Role:       values.GameManagementRoleCollaborator,
				},
			},
			maintainers: []*openapi.Maintainer{
				{
					Id:   uuid.UUID(userID1).String(),
					Name: "mazrean",
					Role: 1,
				},
				{
					Id:   uuid.UUID(userID2).String(),
					Name: "mazrean2",
					Role: 0,
				},
			},
		},
		{
			description:  "管理者がいなくてもエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			strGameID:              uuid.UUID(gameID).String(),
			executeGetGameManagers: true,
			gameID:                 gameID,
			gameManagers:           []*service.GameManager{},
			maintainers:            []*openapi.Maintainer{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/game/maintainer", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.store.Get(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				c.Set(sessionContextKey, sess)
			}

			if testCase.executeGetGameManagers {
				mockGameAuthService.
					EXPECT().
					GetGameManagers(gomock.Any(), gomock.Any(), testCase.gameID).
					Return(testCase.gameManagers, testCase.GetGameManagersErr)
			}

			maintainers, err := gameRoleHandler.GetMaintainer(c, testCase.strGameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Len(t, maintainers, len(testCase.maintainers))

			for i, maintainer := range maintainers {
				assert.Equal(t, testCase.maintainers[i].Id, maintainer.Id)
				assert.Equal(t, testCase.maintainers[i].Name, maintainer.Name)
				assert.Equal(t, testCase.maintainers[i].Role, maintainer.Role)
			}
		})
	}
}
