package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGameFeedbackListTrapMemberAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	config := mockConfig.NewMockHandler(ctrl)
	config.EXPECT().SessionKey().Return("key", nil)
	config.EXPECT().SessionSecret().Return("secret", nil)
	store, err := session.NewSession(config)
	require.NoError(t, err)
	sessions, err := NewSession(store)
	require.NoError(t, err)
	oidc := mock.NewMockOIDCV2(ctrl)
	checker := NewChecker(
		NewContext(),
		sessions,
		oidc,
		mock.NewMockEdition(ctrl),
		mock.NewMockEditionAuth(ctrl),
		mock.NewMockGameRoleV2(ctrl),
		mock.NewMockAdminAuthV2(ctrl),
		mock.NewMockGameV2(ctrl),
	)
	serviceMock := mock.NewMockGameFeedback(ctrl)
	api := &API{Checker: checker, GameFeedback: NewGameFeedback(serviceMock)}
	e := echo.New()
	require.NoError(t, api.SetRoutes(e))
	gameID := values.NewGameID()

	t.Run("未認証なのでserviceを呼び出さず401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/games/"+uuid.UUID(gameID).String()+"/feedbacks", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("edition専用のBearer認証なのでserviceを呼び出さず401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v2/games/"+uuid.UUID(gameID).String()+"/feedbacks", nil)
		req.Header.Set("Authorization", "Bearer edition-access-token")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("member認証済みなのでserviceを呼び出して200", func(t *testing.T) {
		accessToken := "member token"
		oidc.
			EXPECT().
			Authenticate(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, authSession *domain.OIDCSession) error {
				assert.Equal(t, values.NewOIDCAccessToken(accessToken), authSession.GetAccessToken())
				return nil
			})
		serviceMock.
			EXPECT().
			GetGameFeedbacks(gomock.Any(), gameID, 50, 0).
			Return(nil, 0, nil)
		c, req, rec := setupTestRequest(t, http.MethodGet, "/api/v2/games/"+uuid.UUID(gameID).String()+"/feedbacks", nil)
		setTestSession(t, c, req, rec, sessions, domain.NewOIDCSession(values.NewOIDCAccessToken(accessToken), time.Now()))
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
