package v2

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
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
			c, req, rec := setupTestRequest(t, http.MethodGet, "/", nil)

			var traPAuthErr error
			if testCase.isOk {
				traPAuthErr = nil
			} else if testCase.isCheckTraPAuthErr {
				traPAuthErr = errors.New("error")
			} else {
				traPAuthErr = service.ErrOIDCSessionExpired
			}

			accessToken := "access token"
			authSession := domain.NewOIDCSession(values.NewOIDCAccessToken(accessToken), time.Now())
			setTestSession(t, c, req, rec, session, authSession)

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
			c, req, rec := setupTestRequest(t, http.MethodGet, "/", nil)

			if testCase.sessionExist {
				var authSession *domain.OIDCSession
				if testCase.authSessionExist {
					authSession = domain.NewOIDCSession(values.NewOIDCAccessToken(testCase.accessToken), testCase.expiresAt)
				}
				setTestSession(t, c, req, rec, session, authSession)
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
		gameID              string
		noToken             bool
		executeAuthenticate bool
		AuthenticateErr     error
		executeGetGame      bool
		gameVisibility      values.GameVisibility
		GetGameErr          error
		isErr               bool
		expectedErr         error
		statusCode          int
	}

	testCases := map[string]test{
		"部員なので問題なし": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
		},
		"Authenticateがエラーなのでエラー": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
		"publicなのでエラー無し": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			executeGetGame:      true,
			gameVisibility:      values.GameVisibilityTypePublic,
		},
		"limitedなのでエラー無し": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			executeGetGame:      true,
			gameVisibility:      values.GameVisibilityTypeLimited,
		},
		"privateなので403": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			executeGetGame:      true,
			gameVisibility:      values.GameVisibilityTypePrivate,
			isErr:               true,
			statusCode:          http.StatusUnauthorized,
		},
		"トークンが無いがpublicなので問題なし": {
			gameID:         uuid.NewString(),
			noToken:        true,
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypePublic,
		},
		"トークンが無いがlimitedなので問題なし": {
			gameID:         uuid.NewString(),
			noToken:        true,
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypeLimited,
		},
		"トークンが無いがprivateなので403": {
			gameID:         uuid.NewString(),
			noToken:        true,
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypePrivate,
			isErr:          true,
			statusCode:     http.StatusUnauthorized,
		},
		"ゲームが存在しないのでエラー": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			executeGetGame:      true,
			gameVisibility:      values.GameVisibilityTypePrivate,
			GetGameErr:          service.ErrNoGame,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		"gameIDがuuidでないのでエラー": {
			executeAuthenticate: true,
			gameID:              "invalid",
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		"GetGameがエラーなのでエラー": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			AuthenticateErr:     service.ErrOIDCSessionExpired,
			executeGetGame:      true,
			GetGameErr:          errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {

			c, req, rec := setupTestRequest(t, http.MethodGet, "/api/v2/games/"+testCase.gameID, nil)

			var authSession *domain.OIDCSession
			if !testCase.noToken {
				authSession = domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))
			}
			setTestSession(t, c, req, rec, session, authSession)

			if testCase.executeAuthenticate {
				mockOIDCService.
					EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(testCase.AuthenticateErr)
			}

			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Nil(), gomock.Any()).
					Return(&service.GameInfoV2{
						Game: domain.NewGame(values.NewGameIDFromUUID(uuid.MustParse(testCase.gameID)), "name", "description", testCase.gameVisibility, time.Now()),
					}, testCase.GetGameErr)
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

func TestGameFileVisibilityChecker(t *testing.T) {
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
		gameID              string
		hasToken            bool
		executeAuthenticate bool
		AuthenticateErr     error
		executeGetGame      bool
		gameVisibility      values.GameVisibility
		GetGameErr          error
		isError             bool
		statusCode          int
	}

	testCases := map[string]test{
		"部員なのでOK": {
			gameID:              uuid.NewString(),
			executeAuthenticate: true,
			hasToken:            true,
		},
		"トークンが無いが、publicなのでOK": {
			gameID:         uuid.NewString(),
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypePublic,
		},
		"部員でなく、limitedなので401": {
			gameID:         uuid.NewString(),
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypeLimited,
			isError:        true,
			statusCode:     http.StatusUnauthorized,
		},
		"部員でなく、privateなので401": {
			gameID:         uuid.NewString(),
			executeGetGame: true,
			gameVisibility: values.GameVisibilityTypePrivate,
			isError:        true,
			statusCode:     http.StatusUnauthorized,
		},
		"ゲームが存在しないので404": {
			gameID:         uuid.NewString(),
			executeGetGame: true,
			GetGameErr:     service.ErrNoGame,
			isError:        true,
			statusCode:     http.StatusNotFound,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, req, rec := setupTestRequest(t, http.MethodGet, "/api/v2/games/"+testCase.gameID, nil)

			var authSession *domain.OIDCSession
			if testCase.hasToken {
				authSession = domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))
			}
			setTestSession(t, c, req, rec, session, authSession)

			if testCase.executeAuthenticate {
				mockOIDCService.
					EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(testCase.AuthenticateErr)
			}

			if testCase.executeGetGame {
				gameID := values.NewGameIDFromUUID(uuid.MustParse(testCase.gameID))
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Nil(), gameID).
					Return(&service.GameInfoV2{
						Game: domain.NewGame(gameID, "name", "description", testCase.gameVisibility, time.Now()),
					}, testCase.GetGameErr)
			}

			ctx := setEchoContext(context.Background(), c)

			ai := openapi3filter.AuthenticationInput{
				RequestValidationInput: &openapi3filter.RequestValidationInput{
					PathParams: map[string]string{"gameID": testCase.gameID},
				},
			}

			err := checker.GameFileVisibilityChecker(ctx, &ai)

			if !testCase.isError {
				assert.NoError(t, err)
			} else {
				var httpErr *echo.HTTPError
				assert.ErrorAs(t, err, &httpErr)
				assert.Equal(t, testCase.statusCode, httpErr.Code)
			}

		})
	}

}
