package v2

import (
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
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestFeedbackQuestionsMaintainerAuth(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	conf := mockConfig.NewMockHandler(ctrl)
	conf.EXPECT().SessionKey().Return("key", nil)
	conf.EXPECT().SessionSecret().Return("secret", nil)
	store, err := session.NewSession(conf)
	require.NoError(t, err)
	sessions, err := NewSession(store)
	require.NoError(t, err)
	oidc := mock.NewMockOIDCV2(ctrl)
	roles := mock.NewMockGameRoleV2(ctrl)
	admin := mock.NewMockAdminAuthV2(ctrl)
	checker := NewChecker(
		NewContext(),
		sessions,
		oidc,
		mock.NewMockEdition(ctrl),
		mock.NewMockEditionAuth(ctrl),
		roles,
		admin,
		mock.NewMockGameV2(ctrl),
	)
	feedback := mock.NewMockGameFeedback(ctrl)
	api := &API{
		Checker:      checker,
		GameFeedback: NewGameFeedback(feedback),
	}
	e := echo.New()
	require.NoError(t, api.SetRoutes(e))
	gameID := values.NewGameID()

	t.Run("未認証なのでserviceを呼び出さず401", func(t *testing.T) {
		url := "/api/v2/games/" + uuid.UUID(gameID).String() + "/feedback-questions"
		request := httptest.NewRequest(http.MethodPut, url, nil)
		recorder := httptest.NewRecorder()

		e.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("maintainerなのでserviceを呼び出す", func(t *testing.T) {
		access := "token"
		sessionValue := domain.NewOIDCSession(values.NewOIDCAccessToken(access), time.Now())
		admin.EXPECT().AdminAuthorize(gomock.Any(), gomock.Any()).Return(service.ErrForbidden)
		roles.EXPECT().UpdateGameAuth(gomock.Any(), gomock.Any(), gameID).Return(nil)
		feedback.
			EXPECT().
			PutFeedbackQuestions(gomock.Any(), gameID, []service.FeedbackQuestionInput{}).
			Return(nil, nil)

		url := "/api/v2/games/" + uuid.UUID(gameID).String() + "/feedback-questions"
		c, request, recorder := setupTestRequest(
			t,
			http.MethodPut,
			url,
			withJSONBody(t, map[string]any{"questions": []any{}}),
		)
		setTestSession(t, c, request, recorder, sessions, sessionValue)

		e.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}
