package traq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type mockHandlerParam struct {
		isTraQBroken     bool
		accessToken      string
		accessTokenValid bool
		response         *getUsersMeResponse
	}

	var (
		param      *mockHandlerParam
		handlerErr error
		callCount  int

		errNoParamSet            = errors.New("param is not set")
		errUnexpectedAccessToken = errors.New("unexpected access token")
	)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/users/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if param.isTraQBroken {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if param == nil {
			handlerErr = errNoParamSet
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		authorizationHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
		if accessToken != param.accessToken {
			handlerErr = errUnexpectedAccessToken
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !param.accessTokenValid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err := json.NewEncoder(w).Encode(param.response)
		if err != nil {
			handlerErr = err
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Errorf("Error parsing base URL: %v", err)
	}
	userAuth := NewUser(ts.Client(), common.TraQBaseURL(baseURL))

	type test struct {
		description      string
		isTraQBroken     bool
		session          *domain.OIDCSession
		accessTokenValid bool
		response         *getUsersMeResponse
		user             *service.UserInfo
		isErr            bool
		err              error
	}

	id := uuid.New()

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: &getUsersMeResponse{
				ID:    id,
				Name:  "mazrean",
				State: 1,
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(id),
				"mazrean",
				values.TrapMemberStatusActive,
			),
		},
		{
			description:  "traQが壊れているのでエラー",
			isTraQBroken: true,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			isErr: true,
			err:   auth.ErrIdpBroken,
		},
		{
			description:  "access tokenが誤っているのでエラー",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken(""),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: false,
			isErr:            true,
			err:              auth.ErrInvalidSession,
		},
		{
			description:  "stateが0なのでdeactivated",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: &getUsersMeResponse{
				ID:    id,
				Name:  "mazrean",
				State: 0,
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(id),
				"mazrean",
				values.TrapMemberStatusDeactivated,
			),
		},
		{
			description:  "stateが2なのでsuspended",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: &getUsersMeResponse{
				ID:    id,
				Name:  "mazrean",
				State: 2,
			},
			user: service.NewUserInfo(
				values.NewTrapMemberID(id),
				"mazrean",
				values.TrapMemberStatusSuspended,
			),
		},
		{
			description:  "stateが0~2でないのでエラー",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: &getUsersMeResponse{
				ID:    id,
				Name:  "mazrean",
				State: 3,
			},
			isErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				param = nil
				handlerErr = nil
				callCount = 0
			}()
			param = &mockHandlerParam{
				isTraQBroken:     testCase.isTraQBroken,
				accessToken:      string(testCase.session.GetAccessToken()),
				accessTokenValid: testCase.accessTokenValid,
				response:         testCase.response,
			}

			user, err := userAuth.GetMe(ctx, testCase.session)

			assert.NoError(t, handlerErr)
			assert.Equal(t, 1, callCount)

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

			assert.Equal(t, testCase.user.GetID(), user.GetID())
			assert.Equal(t, testCase.user.GetName(), user.GetName())
			assert.Equal(t, testCase.user.GetStatus(), user.GetStatus())
		})
	}
}
