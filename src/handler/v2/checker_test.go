package v2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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
	mockGameRoleService := mock.NewMockGameRoleV2(ctrl)
	mockAdministratorAuthService := mock.NewMockAdminAuthV2(ctrl)
	mockGameService := mock.NewMockGameV2(ctrl)
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
		mockGameRoleService,
		mockAdministratorAuthService,
		mockGameService,
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
	mockGameRoleService := mock.NewMockGameRoleV2(ctrl)
	mockAdministratorAuthService := mock.NewMockAdminAuthV2(ctrl)
	mockGameService := mock.NewMockGameV2(ctrl)
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
		mockGameRoleService,
		mockAdministratorAuthService,
		mockGameService,
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

func TestGameInfoVisibilityChecker(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOIDCService := mock.NewMockOIDCV2(ctrl)
	mockEditionService := mock.NewMockEdition(ctrl)
	mockEditionAuthService := mock.NewMockEditionAuth(ctrl)
	mockGameRoleService := mock.NewMockGameRoleV2(ctrl)
	mockAdministratorAuthService := mock.NewMockAdminAuthV2(ctrl)
	mockGameService := mock.NewMockGameV2(ctrl)
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
		mockGameRoleService,
		mockAdministratorAuthService,
		mockGameService,
	)

	type test struct {
		gameID                   string
		AuthenticateErr          error
		executeGetGameVisibility bool
		gameVisibility           values.GameVisibility
		GetGameVisibilityErr     error
		isErr                    bool
		expectedErr              error
		statusCode               int
	}

	testCases := map[string]test{
		"部員なので問題なし": {
			gameID: uuid.NewString(),
		},
		"Authenticateがエラーなのでエラー": {
			gameID:          uuid.NewString(),
			AuthenticateErr: errors.New("error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		"publicなのでエラー無し": {
			gameID:                   uuid.NewString(),
			AuthenticateErr:          service.ErrOIDCSessionExpired,
			executeGetGameVisibility: true,
			gameVisibility:           values.GameVisibilityTypePublic,
		},
		"limitedなのでエラー無し": {
			gameID:                   uuid.NewString(),
			AuthenticateErr:          service.ErrOIDCSessionExpired,
			executeGetGameVisibility: true,
			gameVisibility:           values.GameVisibilityTypeLimited,
		},
		"privateなので403": {
			gameID:                   uuid.NewString(),
			AuthenticateErr:          service.ErrOIDCSessionExpired,
			executeGetGameVisibility: true,
			gameVisibility:           values.GameVisibilityTypePrivate,
			isErr:                    true,
			statusCode:               http.StatusUnauthorized,
		},
		"ゲームが存在しないのでエラー": {
			gameID:                   uuid.NewString(),
			AuthenticateErr:          service.ErrOIDCSessionExpired,
			executeGetGameVisibility: true,
			gameVisibility:           values.GameVisibilityTypePrivate,
			GetGameVisibilityErr:     service.ErrNoGame,
			isErr:                    true,
			statusCode:               http.StatusNotFound,
		},
		"gameIDがuuidでないのでエラー": {
			gameID:          "invalid",
			AuthenticateErr: service.ErrOIDCSessionExpired,
			isErr:           true,
			statusCode:      http.StatusBadRequest,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v2/games/"+testCase.gameID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			sess, err := session.New(req)
			if err != nil {
				t.Fatal(err)
			}

			sess.Values[accessTokenSessionKey] = "token"
			sess.Values[expiresAtSessionKey] = time.Now().Add(time.Hour)

			err = sess.Save(req, rec)
			if err != nil {
				t.Fatalf("failed to save session: %v", err)
			}

			setCookieHeader(c)

			mockOIDCService.
				EXPECT().
				Authenticate(gomock.Any(), gomock.Any()).
				Return(testCase.AuthenticateErr)

			if testCase.executeGetGameVisibility {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Nil(), gomock.Any()).
					Return(&service.GameInfoV2{
						Game: domain.NewGame(values.NewGameIDFromUUID(uuid.MustParse(testCase.gameID)), "name", "description", testCase.gameVisibility, time.Now()),
					}, testCase.GetGameVisibilityErr)
			}

			ctx := setEchoContext(context.Background(), c)

			ai := openapi3filter.AuthenticationInput{
				RequestValidationInput: &openapi3filter.RequestValidationInput{
					PathParams: map[string]string{"gameID": testCase.gameID},
				},
			}

			err = checker.GameInfoVisibilityChecker(ctx, &ai)

			if testCase.isErr {
				if testCase.expectedErr == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			if testCase.statusCode != 0 {
				httpErr, ok := err.(*echo.HTTPError)
				if !ok {
					t.Errorf("error is not *echo.HTTPError: %v", err)
				}
				assert.Equal(t, testCase.statusCode, httpErr.Code)
			}
		})
	}
}
