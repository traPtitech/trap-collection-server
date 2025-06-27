package v2

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetCallback(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)

	baseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		t.Errorf("Error parsing base URL: %v", err)
	}

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		TraqBaseURL().
		Return(baseURL, nil)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	oauth, err := NewOAuth2(mockConf, session, mockOIDCService)
	if err != nil {
		t.Fatalf("failed to create oauth: %v", err)
	}

	type test struct {
		description       string
		strCode           string
		sessionExist      bool
		codeVerifierExist bool
		codeVerifier      string
		executeCallback   bool
		code              values.OIDCAuthorizationCode
		authSession       *domain.OIDCSession
		CallbackErr       error
		isErr             bool
		err               error
		statusCode        int
	}

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			strCode:           "code",
			sessionExist:      true,
			codeVerifierExist: true,
			codeVerifier:      "codeVerifier",
			executeCallback:   true,
			code:              values.NewOIDCAuthorizationCode("code"),
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
		},
		{
			description:  "セッションがないので400",
			strCode:      "code",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		{
			description:       "codeVerifierがないので400",
			strCode:           "code",
			sessionExist:      true,
			codeVerifierExist: false,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "CallbackがErrInvalidAuthStateOrCodeなので400",
			strCode:           "code",
			sessionExist:      true,
			codeVerifierExist: true,
			codeVerifier:      "codeVerifier",
			executeCallback:   true,
			code:              values.NewOIDCAuthorizationCode("code"),
			CallbackErr:       service.ErrInvalidAuthStateOrCode,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "Callbackがエラーなので500",
			strCode:           "code",
			sessionExist:      true,
			codeVerifierExist: true,
			codeVerifier:      "codeVerifier",
			executeCallback:   true,
			code:              values.NewOIDCAuthorizationCode("code"),
			CallbackErr:       errors.New("error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, req, rec := setupTestRequest(t, http.MethodPost, "/oauth2/callback", nil)

			if testCase.sessionExist {
				values := make([]sessionValue, 0)
				if testCase.codeVerifierExist {
					values = append(values, sessionValue{
						key:   codeVerifierSessionKey,
						value: testCase.codeVerifier,
					})
				}
				setTestSession(t, c, req, rec, session, nil, values...)
			}

			if testCase.executeCallback {
				mockOIDCService.
					EXPECT().
					Callback(gomock.Any(), gomock.Any(), testCase.code).
					Return(testCase.authSession, testCase.CallbackErr)
			}

			err := oauth.GetCallback(c, openapi.GetCallbackParams{
				Code: testCase.strCode,
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
			if err != nil {
				return
			}

			sess, err := session.get(c)
			if err != nil {
				t.Fatalf("failed to get session: %v", err)
			}

			_, err = session.getAuthSession(sess)
			assert.NoError(t, err)
		})
	}
}

func TestGetCode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)

	baseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		t.Errorf("Error parsing base URL: %v", err)
	}

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		TraqBaseURL().
		Return(baseURL, nil)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	oauth, err := NewOAuth2(mockConf, session, mockOIDCService)
	if err != nil {
		t.Fatalf("failed to create oauth: %v", err)
	}

	type test struct {
		description         string
		client              *domain.OIDCClient
		authState           *domain.OIDCAuthState
		AuthorizeErr        error
		sessionExist        bool
		scheme              string
		host                string
		path                string
		clientID            string
		codeChallenge       string
		codeChallengeMethod string
		responseType        string
		isErr               bool
		err                 error
		statusCode          int
	}

	codeVerifier, err := values.NewOIDCCodeVerifier()
	if err != nil {
		t.Fatalf("failed to create codeVerifier: %v", err)
	}
	codeChallenge, err := codeVerifier.GetCodeChallenge(values.OIDCCodeChallengeMethodSha256)
	if err != nil {
		t.Fatalf("failed to get codeChallenge:%v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			client: domain.NewOIDCClient(
				"clientID",
			),
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			sessionExist:        true,
			isErr:               true,
			statusCode:          http.StatusSeeOther,
			scheme:              "https",
			host:                "q.trap.jp",
			path:                "/api/v3/oauth2/authorize",
			clientID:            "clientID",
			codeChallenge:       string(codeChallenge),
			codeChallengeMethod: "S256",
			responseType:        "code",
		},
		{
			description:  "Authorizeがエラーなので500",
			AuthorizeErr: errors.New("error"),
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			// 実際には発生し得ないが念の為テスト項目に入れている
			description: "CodeChallengeMethodがSHA256以外なので500",
			client: domain.NewOIDCClient(
				"clientID",
			),
			authState: domain.NewOIDCAuthState(
				100,
				codeVerifier,
			),
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description: "sessionが存在しなくてもエラーなし",
			client: domain.NewOIDCClient(
				"clientID",
			),
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			sessionExist:        false,
			isErr:               true,
			statusCode:          http.StatusSeeOther,
			scheme:              "https",
			host:                "q.trap.jp",
			path:                "/api/v3/oauth2/authorize",
			clientID:            "clientID",
			codeChallenge:       string(codeChallenge),
			codeChallengeMethod: "S256",
			responseType:        "code",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, req, rec := setupTestRequest(t, http.MethodPost, "/oauth2/callback", nil)

			if testCase.sessionExist {
				setTestSession(t, c, req, rec, session, nil)
			}

			mockOIDCService.
				EXPECT().
				GenerateAuthState(gomock.Any()).
				Return(testCase.client, testCase.authState, testCase.AuthorizeErr)

			err := oauth.GetCode(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)

						strRedirectURL := c.Response().Header().Get("Location")
						redirectURL, err := url.Parse(strRedirectURL)
						if err != nil {
							t.Fatalf("failed to parse redirectURL: %v", err)
						}

						assert.Equal(t, testCase.scheme, redirectURL.Scheme)
						assert.Equal(t, testCase.host, redirectURL.Host)
						assert.Equal(t, testCase.path, redirectURL.Path)
						assert.Equal(t, testCase.clientID, redirectURL.Query().Get("client_id"))
						assert.Equal(t, testCase.codeChallenge, redirectURL.Query().Get("code_challenge"))
						assert.Equal(t, testCase.codeChallengeMethod, redirectURL.Query().Get("code_challenge_method"))
						assert.Equal(t, testCase.responseType, redirectURL.Query().Get("response_type"))
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

			setCookieHeader(c)

			sess, err := session.get(c)
			if err != nil {
				t.Fatalf("failed to get session: %v", err)
			}

			_, err = session.getCodeVerifier(sess)
			assert.NoError(t, err)
		})
	}
}

func TestPostLogout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)

	baseURL, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		t.Errorf("Error parsing base URL: %v", err)
	}

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		TraqBaseURL().
		Return(baseURL, nil)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	oauth, err := NewOAuth2(mockConf, session, mockOIDCService)
	if err != nil {
		t.Fatalf("failed to create oauth: %v", err)
	}

	type test struct {
		description      string
		sessionExist     bool
		authSessionExist bool
		accessToken      string
		expiresAt        time.Time
		executeLogout    bool
		LogoutErr        error
		isErr            bool
		err              error
		statusCode       int
	}

	testCases := []test{
		{
			description:      "特に問題ないのでエラーなし",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "accessToken",
			expiresAt:        time.Now(),
			executeLogout:    true,
		},
		{
			// middlewareで弾かれるので、この状況にはならないはずだが、念のためテスト
			description:  "sessionが存在しないので400",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		{
			description:      "authSessionが存在しないので400",
			sessionExist:     true,
			authSessionExist: false,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		{
			description:      "Logoutがエラーなので500",
			sessionExist:     true,
			authSessionExist: true,
			accessToken:      "accessToken",
			expiresAt:        time.Now(),
			executeLogout:    true,
			LogoutErr:        errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c, req, rec := setupTestRequest(t, http.MethodPost, "/oauth2/logout", nil)

			if testCase.sessionExist {
				var authSession *domain.OIDCSession
				if testCase.authSessionExist {
					authSession = domain.NewOIDCSession(values.NewOIDCAccessToken(testCase.accessToken), testCase.expiresAt)
				}
				setTestSession(t, c, req, rec, session, authSession)
			}

			if testCase.executeLogout {
				mockOIDCService.
					EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(testCase.LogoutErr)
			}

			err := oauth.PostLogout(c)

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

			setCookieHeader(c)

			sess, err := session.get(c)
			if err != nil {
				t.Fatalf("failed to get session: %v", err)
			}
			assert.Less(t, sess.Options.MaxAge, 0)
		})
	}
}
