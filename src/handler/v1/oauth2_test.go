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
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestCallback(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	oauth := NewOAuth2(session, mockOIDCService)

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
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/oauth2/callback", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.codeVerifierExist {
					sess.Values[codeVerifierSessionKey] = testCase.codeVerifier
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)
			}

			if testCase.executeCallback {
				mockOIDCService.
					EXPECT().
					Callback(gomock.Any(), gomock.Any(), testCase.code).
					Return(testCase.authSession, testCase.CallbackErr)
			}

			err := oauth.Callback(testCase.strCode, c)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			sess, err := session.getSession(c)
			if err != nil {
				t.Fatalf("failed to get session: %v", err)
			}

			_, err = session.getAuthSession(sess)
			assert.NoError(t, err)
		})
	}
}

func TestGetGeneratedCode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDC(ctrl)
	session := NewSession("key", "secret")

	oauth := NewOAuth2(session, mockOIDCService)

	type test struct {
		description  string
		client       *domain.OIDCClient
		authState    *domain.OIDCAuthState
		AuthorizeErr error
		sessionExist bool
		response     openapi.InlineResponse200
		isErr        bool
		err          error
		statusCode   int
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
			sessionExist: true,
			response: openapi.InlineResponse200{
				CodeChallenge:       string(codeChallenge),
				CodeChallengeMethod: "S256",
				ClientId:            "clientID",
				ResponseType:        "code",
			},
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
			sessionExist: false,
			response: openapi.InlineResponse200{
				CodeChallenge:       string(codeChallenge),
				CodeChallengeMethod: "S256",
				ClientId:            "clientID",
				ResponseType:        "code",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/oauth2/callback", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)
			}

			mockOIDCService.
				EXPECT().
				Authorize(gomock.Any()).
				Return(testCase.client, testCase.authState, testCase.AuthorizeErr)

			response, err := oauth.GetGeneratedCode(c)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.response, *response)

			setCookieHeader(c)

			sess, err := session.getSession(c)
			if err != nil {
				t.Fatalf("failed to get session: %v", err)
			}

			_, err = session.getCodeVerifier(sess)
			assert.NoError(t, err)
		})
	}
}
