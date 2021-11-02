package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

type CallChecker struct {
	IsCalled bool
}

func (cc *CallChecker) Handler(c echo.Context) error {
	cc.IsCalled = true

	return c.NoContent(http.StatusOK)
}

func TestTrapMemberAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description        string
		isOk               bool
		isCheckTraPAuthErr bool
		isCalled           bool
		statusCode         int
	}

	testCases := []test{
		{
			description: "okかつエラーなしなので通す",
			isOk:        true,
			isCalled:    true,
			statusCode:  http.StatusOK,
		},
		{
			description: "okでないなので401",
			isOk:        false,
			statusCode:  http.StatusUnauthorized,
		},
		{
			description:        "CheckLauncherAuthがエラーなので401",
			isCheckTraPAuthErr: true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			var traPAuthErr error
			if testCase.isOk {
				traPAuthErr = nil
			} else if testCase.isCheckTraPAuthErr {
				traPAuthErr = errors.New("error")
			} else {
				traPAuthErr = service.ErrOIDCSessionExpired
			}

			accessToken := "access token"
			sess, err := session.store.New(req, session.key)
			if err != nil {
				t.Fatal(err)
			}

			sess.Values[accessTokenSessionKey] = accessToken
			sess.Values[expiresAtSessionKey] = time.Now()

			err = sess.Save(req, rec)
			if err != nil {
				t.Fatalf("failed to save session: %v", err)
			}

			setCookieHeader(c)

			mockOIDCService.
				EXPECT().
				TraPAuth(gomock.Any(), gomock.Any()).
				Return(traPAuthErr)

			callChecker := CallChecker{}

			e.HTTPErrorHandler(middleware.TrapMemberAuthMiddleware(callChecker.Handler)(c), c)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.Equal(t, testCase.isCalled, callChecker.IsCalled, testCase.description)
		})
	}
}

func TestLauncherAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description            string
		isOk                   bool
		isCheckLauncherAuthErr bool
		isCalled               bool
		statusCode             int
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	testCases := []test{
		{
			description: "okかつエラーなしなので通す",
			isOk:        true,
			isCalled:    true,
			statusCode:  http.StatusOK,
		},
		{
			description: "okでないので401",
			isOk:        false,
			statusCode:  http.StatusUnauthorized,
		},
		{
			description:            "CheckLauncherAuthがエラーなので401",
			isCheckLauncherAuthErr: true,
			statusCode:             http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			var launcherAuthErr error
			if testCase.isCheckLauncherAuthErr {
				launcherAuthErr = errors.New("error")
			} else if testCase.isOk {
				launcherAuthErr = nil
			} else {
				launcherAuthErr = service.ErrInvalidLauncherSessionAccessToken
			}

			req.Header.Set(echo.HeaderAuthorization, "Bearer "+string(accessToken))
			mockLauncherAuthService.
				EXPECT().
				LauncherAuth(c.Request().Context(), accessToken).
				Return(&domain.LauncherUser{}, &domain.LauncherVersion{}, launcherAuthErr)

			callChecker := CallChecker{}

			e.HTTPErrorHandler(middleware.LauncherAuthMiddleware(callChecker.Handler)(c), c)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.Equal(t, testCase.isCalled, callChecker.IsCalled, testCase.description)
		})
	}
}

func TestBothAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description            string
		isCheckLauncherAuthOk  bool
		isCheckLauncherAuthErr bool
		executeTraPAuth        bool
		isCheckTraPAuthOk      bool
		isCheckTraPAuthErr     bool
		isCalled               bool
		statusCode             int
	}

	testCases := []test{
		{
			description:           "LauncherAuthがokなので通す",
			isCheckLauncherAuthOk: true,
			isCalled:              true,
			statusCode:            http.StatusOK,
		},
		{
			description:            "CheckLauncherAuthがエラーなので500",
			isCheckLauncherAuthErr: true,
			statusCode:             http.StatusInternalServerError,
		},
		{
			description:           "LauncherAuthがokでなくてもTraPAuthがokなので通す",
			isCheckLauncherAuthOk: false,
			executeTraPAuth:       true,
			isCheckTraPAuthOk:     true,
			isCalled:              true,
			statusCode:            http.StatusOK,
		},
		{
			description:           "TraPAuthがエラーなので500",
			isCheckLauncherAuthOk: false,
			executeTraPAuth:       true,
			isCheckTraPAuthOk:     false,
			isCheckTraPAuthErr:    true,
			statusCode:            http.StatusInternalServerError,
		},
		{
			description:           "LauncherAuth、TraPAuthともにfalseなので401",
			isCheckLauncherAuthOk: false,
			executeTraPAuth:       true,
			isCheckTraPAuthOk:     false,
			statusCode:            http.StatusUnauthorized,
		},
	}

	launcherAccessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			var launcherAuthErr error
			if testCase.isCheckLauncherAuthErr {
				launcherAuthErr = errors.New("error")
			} else if testCase.isCheckLauncherAuthOk {
				launcherAuthErr = nil
			} else {
				launcherAuthErr = service.ErrInvalidLauncherSessionAccessToken
			}
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+string(launcherAccessToken))
			mockLauncherAuthService.
				EXPECT().
				LauncherAuth(c.Request().Context(), launcherAccessToken).
				Return(&domain.LauncherUser{}, &domain.LauncherVersion{}, launcherAuthErr)

			if testCase.executeTraPAuth {
				var traPAuthErr error
				if testCase.isCheckTraPAuthOk {
					traPAuthErr = nil
				} else if testCase.isCheckTraPAuthErr {
					traPAuthErr = errors.New("error")
				} else {
					traPAuthErr = service.ErrOIDCSessionExpired
				}

				traPAuthAccessToken := "access token"
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				sess.Values[accessTokenSessionKey] = traPAuthAccessToken
				sess.Values[expiresAtSessionKey] = time.Now()

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				mockOIDCService.
					EXPECT().
					TraPAuth(gomock.Any(), gomock.Any()).
					Return(traPAuthErr)
			}

			callChecker := CallChecker{}

			e.HTTPErrorHandler(middleware.BothAuthMiddleware(callChecker.Handler)(c), c)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.Equal(t, testCase.isCalled, callChecker.IsCalled)
		})
	}
}

func TestCheckTrapMemberAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description      string
		sessionExist     bool
		authSessionExist bool
		accessToken      string
		expiresAt        time.Time
		executeTraPAuth  bool
		TraPAuthErr      error
		isOk             bool
		message          string
		isErr            bool
		err              error
	}

	testCases := []test{
		{
			description:      "特に問題ないのでtrue",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "access token",
			expiresAt:        time.Now(),
			executeTraPAuth:  true,
			isOk:             true,
		},
		{
			description:  "セッションがないのでfalse",
			sessionExist: false,
			isOk:         false,
			message:      "no access token",
		},
		{
			description:      "authSessionがないのでfalse",
			sessionExist:     true,
			authSessionExist: false,
			isOk:             false,
			message:          "no access token",
		},
		{
			description:      "TraPAuthがErrOIDCSessionExpiredなのでfalse",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "access token",
			expiresAt:        time.Now(),
			executeTraPAuth:  true,
			TraPAuthErr:      service.ErrOIDCSessionExpired,
			isOk:             false,
			message:          "access token is expired",
		},
		{
			description:      "TraPAuthがエラー(ErrOIDCSessionExpired以外)なのでエラー",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "access token",
			expiresAt:        time.Now(),
			executeTraPAuth:  true,
			TraPAuthErr:      errors.New("error"),
			isErr:            true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSessionExist {
					sess.Values[accessTokenSessionKey] = testCase.accessToken
					sess.Values[expiresAtSessionKey] = testCase.expiresAt
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)
			}

			if testCase.executeTraPAuth {
				mockOIDCService.
					EXPECT().
					TraPAuth(gomock.Any(), gomock.Any()).
					Return(testCase.TraPAuthErr)
			}

			ok, message, err := middleware.checkTrapMemberAuth(c)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.isOk, ok)
			assert.Equal(t, testCase.message, message)
		})
	}
}

func TestCheckLauncherAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description         string
		authorizationHeader string
		executeLauncherAuth bool
		accessToken         values.LauncherSessionAccessToken
		launcherUser        *domain.LauncherUser
		launcherVersion     *domain.LauncherVersion
		LauncherAuthErr     error
		setValues           bool
		ok                  bool
		message             string
		isErr               bool
		err                 error
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	productKey1, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description:         "Authorizationヘッダーが空文字なのでfalse",
			authorizationHeader: "",
			ok:                  false,
			message:             "invalid authorization header: ",
		},
		{
			description:         "AuthorizationヘッダーがBearerトークンでないのでfalse",
			authorizationHeader: "Basic",
			ok:                  false,
			message:             "invalid authorization header: Basic",
		},
		{
			description:         "accessTokenの形式が不正なのでfalse",
			authorizationHeader: "Bearer a",
			ok:                  false,
			message:             "invalid access token: a",
		},
		{
			description:         "accessTokenが空文字なのでfalse",
			authorizationHeader: "Bearer ",
			ok:                  false,
			message:             "invalid access token: ",
		},
		{
			description:         "LauncherAuthがErrInvalidLauncherSessionAccessTokenなのでfalse",
			authorizationHeader: "Bearer " + string(accessToken),
			executeLauncherAuth: true,
			accessToken:         accessToken,
			LauncherAuthErr:     service.ErrInvalidLauncherSessionAccessToken,
			ok:                  false,
			message:             "invalid access token",
		},
		{
			description:         "LauncherAuthがErrLauncherSessionAccessTokenExpiredなのでfalse",
			authorizationHeader: "Bearer " + string(accessToken),
			executeLauncherAuth: true,
			accessToken:         accessToken,
			LauncherAuthErr:     service.ErrLauncherSessionAccessTokenExpired,
			ok:                  false,
			message:             "access token expired",
		},
		{
			description:         "LauncherAuthがエラーなのでエラー",
			authorizationHeader: "Bearer " + string(accessToken),
			executeLauncherAuth: true,
			accessToken:         accessToken,
			LauncherAuthErr:     errors.New("error"),
			ok:                  false,
			isErr:               true,
		},
		{
			description:         "LauncherAuthが成功",
			authorizationHeader: "Bearer " + string(accessToken),
			executeLauncherAuth: true,
			accessToken:         accessToken,
			launcherUser: domain.NewLauncherUser(
				values.NewLauncherUserID(),
				productKey1,
			),
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				values.NewLauncherVersionID(),
				values.NewLauncherVersionName("test"),
				time.Now(),
			),
			setValues: true,
			ok:        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			req.Header.Set(echo.HeaderAuthorization, testCase.authorizationHeader)

			if testCase.executeLauncherAuth {
				mockLauncherAuthService.
					EXPECT().
					LauncherAuth(c.Request().Context(), testCase.accessToken).
					Return(testCase.launcherUser, testCase.launcherVersion, testCase.LauncherAuthErr)
			}

			ok, message, err := middleware.checkLauncherAuth(c)

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

			assert.Equal(t, testCase.ok, ok)
			assert.Equal(t, testCase.message, message)

			if testCase.setValues {
				launcherUser, err := getLauncherUser(c)
				assert.NoError(t, err)
				assert.Equal(t, testCase.launcherUser, launcherUser)

				launcherVersion, err := getLauncherVersion(c)
				assert.NoError(t, err)
				assert.Equal(t, testCase.launcherVersion, launcherVersion)
			}
		})
	}
}

func TestGameMaintainerAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description           string
		sessionExist          bool
		authSession           *domain.OIDCSession
		strGameID             string
		executeUpdateGameAuth bool
		gameID                values.GameID
		UpdateGameAuthErr     error
		isCalled              bool
		statusCode            int
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:  "特に問題ないので通過",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:             uuid.UUID(gameID).String(),
			executeUpdateGameAuth: true,
			gameID:                gameID,
			isCalled:              true,
			statusCode:            http.StatusOK,
		},
		{
			description:  "セッションがないので500",
			sessionExist: false,
			strGameID:    uuid.UUID(gameID).String(),
			isCalled:     false,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionが存在しないので500",
			sessionExist: true,
			strGameID:    uuid.UUID(gameID).String(),
			isCalled:     false,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "gameIDが不正なので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:  "invalid",
			isCalled:   false,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "ErrInvalidGameIDなので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:             uuid.UUID(gameID).String(),
			executeUpdateGameAuth: true,
			gameID:                gameID,
			UpdateGameAuthErr:     service.ErrInvalidGameID,
			isCalled:              false,
			statusCode:            http.StatusBadRequest,
		},
		{
			description:  "ErrForbiddenなので403",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:             uuid.UUID(gameID).String(),
			executeUpdateGameAuth: true,
			gameID:                gameID,
			UpdateGameAuthErr:     service.ErrForbidden,
			isCalled:              false,
			statusCode:            http.StatusForbidden,
		},
		{
			description:  "UpdateGameAuthがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:             uuid.UUID(gameID).String(),
			executeUpdateGameAuth: true,
			gameID:                gameID,
			UpdateGameAuthErr:     errors.New("error"),
			isCalled:              false,
			statusCode:            http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/games/%s", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/games/:gameID")
			c.SetParamNames("gameID")
			c.SetParamValues(testCase.strGameID)

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

			if testCase.executeUpdateGameAuth {
				mockGameAuthService.
					EXPECT().
					UpdateGameAuth(gomock.Any(), gomock.Any(), testCase.gameID).
					Return(testCase.UpdateGameAuthErr)
			}

			callChecker := CallChecker{}

			e.HTTPErrorHandler(middleware.GameMaintainerAuthMiddleware(callChecker.Handler)(c), c)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.Equal(t, testCase.isCalled, callChecker.IsCalled)
		})
	}
}

func TestGameOwnerAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockGameAuthService := mock.NewMockGameAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(
		session,
		mockLauncherAuthService,
		mockGameAuthService,
		mockOIDCService,
	)

	type test struct {
		description                         string
		sessionExist                        bool
		authSession                         *domain.OIDCSession
		strGameID                           string
		executeUpdateGameManagementRoleAuth bool
		gameID                              values.GameID
		UpdateGameManagementRoleAuthErr     error
		isCalled                            bool
		statusCode                          int
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:  "特に問題ないので通過",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:                           uuid.UUID(gameID).String(),
			executeUpdateGameManagementRoleAuth: true,
			gameID:                              gameID,
			isCalled:                            true,
			statusCode:                          http.StatusOK,
		},
		{
			description:  "セッションがないので500",
			sessionExist: false,
			strGameID:    uuid.UUID(gameID).String(),
			isCalled:     false,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionが存在しないので500",
			sessionExist: true,
			strGameID:    uuid.UUID(gameID).String(),
			isCalled:     false,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "gameIDが不正なので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:  "invalid",
			isCalled:   false,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "ErrInvalidGameIDなので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:                           uuid.UUID(gameID).String(),
			executeUpdateGameManagementRoleAuth: true,
			gameID:                              gameID,
			UpdateGameManagementRoleAuthErr:     service.ErrInvalidGameID,
			isCalled:                            false,
			statusCode:                          http.StatusBadRequest,
		},
		{
			description:  "ErrForbiddenなので403",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:                           uuid.UUID(gameID).String(),
			executeUpdateGameManagementRoleAuth: true,
			gameID:                              gameID,
			UpdateGameManagementRoleAuthErr:     service.ErrForbidden,
			isCalled:                            false,
			statusCode:                          http.StatusForbidden,
		},
		{
			description:  "UpdateGameAuthがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"access token",
				time.Now().Add(time.Hour),
			),
			strGameID:                           uuid.UUID(gameID).String(),
			executeUpdateGameManagementRoleAuth: true,
			gameID:                              gameID,
			UpdateGameManagementRoleAuthErr:     errors.New("error"),
			isCalled:                            false,
			statusCode:                          http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/games/%s", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/games/:gameID")
			c.SetParamNames("gameID")
			c.SetParamValues(testCase.strGameID)

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

			if testCase.executeUpdateGameManagementRoleAuth {
				mockGameAuthService.
					EXPECT().
					UpdateGameManagementRoleAuth(gomock.Any(), gomock.Any(), testCase.gameID).
					Return(testCase.UpdateGameManagementRoleAuthErr)
			}

			callChecker := CallChecker{}

			e.HTTPErrorHandler(middleware.GameOwnerAuthMiddleware(callChecker.Handler)(c), c)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.Equal(t, testCase.isCalled, callChecker.IsCalled)
		})
	}
}
