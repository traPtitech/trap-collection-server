package traq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestGetMe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	mockConfig := mock.NewMockAuthTraQ(ctrl)
	mockConfig.
		EXPECT().
		HTTPClient().
		Return(ts.Client(), nil)
	mockConfig.
		EXPECT().
		BaseURL().
		Return(baseURL, nil)
	userAuth, err := NewUser(mockConfig)
	if err != nil {
		t.Fatalf("Error creating user auth: %v", err)
	}

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
				false,
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
				false,
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
				false,
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

func TestGetAllActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type mockHandlerParam struct {
		isTraQBroken     bool
		accessToken      string
		accessTokenValid bool
		response         []*getUsersResponse
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
		if r.URL.Path != "/users" {
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

	mockConfig := mock.NewMockAuthTraQ(ctrl)
	mockConfig.
		EXPECT().
		HTTPClient().
		Return(ts.Client(), nil)
	mockConfig.
		EXPECT().
		BaseURL().
		Return(baseURL, nil)
	userAuth, err := NewUser(mockConfig)
	if err != nil {
		t.Fatalf("Error creating user auth: %v", err)
		return
	}

	type test struct {
		description      string
		isTraQBroken     bool
		session          *domain.OIDCSession
		accessTokenValid bool
		response         []*getUsersResponse
		users            []*service.UserInfo
		isErr            bool
		err              error
	}

	id1 := uuid.New()
	id2 := uuid.New()

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 1,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
					false,
				),
			},
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
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 0,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusDeactivated,
					false,
				),
			},
		},
		{
			description:  "stateが2なのでsuspended",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 2,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusSuspended,
					false,
				),
			},
		},
		{
			description:  "stateが0~2でないのでエラー",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 3,
				},
			},
			isErr: true,
		},
		{
			description:  "userが0人でもエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response:         []*getUsersResponse{},
			users:            []*service.UserInfo{},
		},
		{
			description:  "userが複数人でもエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 1,
				},
				{
					ID:    id2,
					Name:  "mazrean2",
					State: 0,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
					false,
				),
				service.NewUserInfo(
					values.NewTrapMemberID(id2),
					"mazrean2",
					values.TrapMemberStatusDeactivated,
					false,
				),
			},
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

			users, err := userAuth.GetAllActiveUsers(ctx, testCase.session)

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

			assert.Equal(t, len(testCase.users), len(users))
			for i, user := range testCase.users {
				assert.Equal(t, user.GetID(), users[i].GetID())
				assert.Equal(t, user.GetName(), users[i].GetName())
				assert.Equal(t, user.GetStatus(), users[i].GetStatus())
			}
		})
	}
}

func TestGetActiveUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type mockHandlerParam struct {
		isTraQBroken     bool
		accessToken      string
		accessTokenValid bool
		response         []*getUsersResponse
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
		if r.URL.Path != "/users" {
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

	mockConfig := mock.NewMockAuthTraQ(ctrl)
	mockConfig.
		EXPECT().
		HTTPClient().
		Return(ts.Client(), nil)
	mockConfig.
		EXPECT().
		BaseURL().
		Return(baseURL, nil)
	userAuth, err := NewUser(mockConfig)
	if err != nil {
		t.Fatalf("Error creating user auth: %v", err)
		return
	}

	type test struct {
		description      string
		isTraQBroken     bool
		session          *domain.OIDCSession
		accessTokenValid bool
		response         []*getUsersResponse
		users            []*service.UserInfo
		isErr            bool
		err              error
	}

	id1 := uuid.New()
	id2 := uuid.New()

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 1,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
					false,
				),
			},
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
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 0,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusDeactivated,
					false,
				),
			},
		},
		{
			description:  "stateが2なのでsuspended",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 2,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusSuspended,
					false,
				),
			},
		},
		{
			description:  "stateが0~2でないのでエラー",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 3,
				},
			},
			isErr: true,
		},
		{
			description:  "userが0人でもエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response:         []*getUsersResponse{},
			users:            []*service.UserInfo{},
		},
		{
			description:  "userが複数人でもエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			),
			accessTokenValid: true,
			response: []*getUsersResponse{
				{
					ID:    id1,
					Name:  "mazrean",
					State: 1,
				},
				{
					ID:    id2,
					Name:  "mazrean2",
					State: 0,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(id1),
					"mazrean",
					values.TrapMemberStatusActive,
					false,
				),
				service.NewUserInfo(
					values.NewTrapMemberID(id2),
					"mazrean2",
					values.TrapMemberStatusDeactivated,
					false,
				),
			},
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

			users, err := userAuth.GetActiveUsers(ctx, testCase.session)

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

			assert.Equal(t, len(testCase.users), len(users))
			for i, user := range testCase.users {
				assert.Equal(t, user.GetID(), users[i].GetID())
				assert.Equal(t, user.GetName(), users[i].GetName())
				assert.Equal(t, user.GetStatus(), users[i].GetStatus())
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	userName := "mazrean"

	testCases := map[string]struct {
		statusCode   int
		traqResponse []*getUsersResponse
		users        []*service.UserInfo
		err          error
	}{
		"traQが500を返すので、ErrIdpBroken": {
			statusCode: http.StatusInternalServerError,
			err:        auth.ErrIdpBroken,
		},
		"traQが401を返すので、ErrInvalidSession": {
			statusCode: http.StatusUnauthorized,
			err:        auth.ErrInvalidSession,
		},
		"deactivatedユーザー": {
			statusCode: http.StatusOK,
			traqResponse: []*getUsersResponse{
				{
					ID:    userID,
					Name:  userName,
					State: 0,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(userID),
					values.NewTrapMemberName(userName),
					values.TrapMemberStatusDeactivated,
					false,
				),
			},
		},
		"activeユーザー": {
			statusCode: http.StatusOK,
			traqResponse: []*getUsersResponse{
				{
					ID:    userID,
					Name:  userName,
					State: 1,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(userID),
					values.NewTrapMemberName(userName),
					values.TrapMemberStatusActive,
					false,
				),
			},
		},
		"suspendedユーザー": {
			statusCode: http.StatusOK,
			traqResponse: []*getUsersResponse{
				{
					ID:    userID,
					Name:  userName,
					State: 2,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(userID),
					values.NewTrapMemberName(userName),
					values.TrapMemberStatusSuspended,
					false,
				),
			},
		},
		"botユーザー": {
			statusCode: http.StatusOK,
			traqResponse: []*getUsersResponse{
				{
					ID:    userID,
					Name:  userName,
					State: 1,
					Bot:   true,
				},
			},
			users: []*service.UserInfo{
				service.NewUserInfo(
					values.NewTrapMemberID(userID),
					values.NewTrapMemberName(userName),
					values.TrapMemberStatusActive,
					true,
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewUnstartedServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/users" {
						w.WriteHeader(http.StatusNotFound)
						return
					}

					if r.Method != http.MethodGet {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}

					includeSuspendedStr := r.URL.Query().Get("include-suspended")
					includeSuspended, err := strconv.ParseBool(includeSuspendedStr)
					assert.NoError(t, err)
					assert.True(t, includeSuspended)

					body, err := json.Marshal(testCase.traqResponse)
					require.NoError(t, err)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(testCase.statusCode)
					_, err = w.Write(body)
					require.NoError(t, err)
				},
			))
			ts.EnableHTTP2 = true
			ts.StartTLS()
			defer ts.Close()

			baseURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			ctrl := gomock.NewController(t)

			mockConfig := mock.NewMockAuthTraQ(ctrl)
			mockConfig.
				EXPECT().
				HTTPClient().
				Return(ts.Client(), nil)
			mockConfig.
				EXPECT().
				BaseURL().
				Return(baseURL, nil)
			userAuth, err := NewUser(mockConfig)
			require.NoError(t, err)

			session := domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(5*time.Second),
			)

			users, err := userAuth.GetAllUsers(t.Context(), session)

			assert.ErrorIs(t, err, testCase.err)

			if testCase.err != nil {
				return
			}

			assert.Equal(t, testCase.users, users)
		})
	}
}
