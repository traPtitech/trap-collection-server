package v2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func setEchoContext(ctx context.Context, c echo.Context) context.Context {
	// nolint:staticcheck
	return context.WithValue(ctx, oapiMiddleware.EchoContextKey, c)
}

func TestTrapMemberAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)
	mockEditionService := mock.NewMockEdition(ctrl)
	mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
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

	checker := NewChecker(
		NewContext(),
		session,
		mockOIDCService,
		mockEditionService,
		mockEditionAuthService,
	)

	type test struct {
		description        string
		isOk               bool
		isCheckTraPAuthErr bool
		isErr              bool
		err                error
		statusCode         int
	}

	testCases := []test{
		{
			description: "okかつエラーなしなので通す",
			isOk:        true,
		},
		{
			description: "okでないので401",
			isErr:       true,
			statusCode:  http.StatusUnauthorized,
		},
		{
			description:        "CheckLauncherAuthがエラーなので401",
			isCheckTraPAuthErr: true,
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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
			sess, err := session.New(req)
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

			ctx := setEchoContext(context.Background(), c)

			mockOIDCService.
				EXPECT().
				Authenticate(gomock.Any(), gomock.Any()).
				Return(traPAuthErr)

			err = checker.TrapMemberAuthChecker(ctx, nil)

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
		})
	}
}

func TestCheckTrapMemberAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)
	mockEditionService := mock.NewMockEdition(ctrl)
	mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
	mockConf := mockConfig.NewMockHandler(ctrl)
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

	checker := NewChecker(
		NewContext(),
		session,
		mockOIDCService,
		mockEditionService,
		mockEditionAuthService,
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
				sess, err := session.New(req)
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
					Authenticate(gomock.Any(), gomock.Any()).
					Return(testCase.TraPAuthErr)
			}

			ok, message, err := checker.checkTrapMemberAuth(c)

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
