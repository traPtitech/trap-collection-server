package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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

func TestLauncherAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(session, mockLauncherAuthService, mockOIDCService)

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

func TestCheckTrapMemberAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherAuthService := mock.NewMockLauncherAuth(ctrl)
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(session, mockLauncherAuthService, mockOIDCService)

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
	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	middleware := NewMiddleware(session, mockLauncherAuthService, mockOIDCService)

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
