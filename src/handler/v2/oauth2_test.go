package v2

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
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
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/oauth2/callback", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
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
