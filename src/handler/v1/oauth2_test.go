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
