package traq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

func TestGetOIDCSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// ref: https://github.com/traPtitech/traQ/blob/master/router/oauth2/token_endpoint.go#L21-L27
	type tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}
	type mockHandlerParam struct {
		isTraQBroken      bool
		code              string
		codeValid         bool
		clientID          string
		clientIDValid     bool
		codeVerifier      string
		codeVerifierValid bool
		*tokenResponse
	}

	var (
		param      *mockHandlerParam
		handlerErr error
		callCount  int

		errNoParamSet             = errors.New("param is not set")
		errUnexpectedCode         = errors.New("unexpected code")
		errUnexpectedClientID     = errors.New("unexpected clientID")
		errUnexpectedCodeVerifier = errors.New("unexpected codeVerifier")
	)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/oauth2/token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
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

		code := r.FormValue("code")
		if code != param.code {
			handlerErr = errUnexpectedCode
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !param.codeValid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clientID := r.FormValue("client_id")
		if clientID != param.clientID {
			handlerErr = errUnexpectedClientID
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !param.clientIDValid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		codeVerifier := r.FormValue("code_verifier")
		if codeVerifier != param.codeVerifier {
			handlerErr = errUnexpectedCodeVerifier
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !param.codeVerifierValid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := json.NewEncoder(w).Encode(param.tokenResponse)
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
	oidcAuth, err := NewOIDC(mockConfig)
	if err != nil {
		t.Fatalf("Error creating OIDC: %v", err)
		return
	}

	type test struct {
		description       string
		isTraQBroken      bool
		code              values.OIDCAuthorizationCode
		codeValid         bool
		client            *domain.OIDCClient
		clientIDValid     bool
		authState         *domain.OIDCAuthState
		codeVerifierValid bool
		tokenResponse     *tokenResponse
		session           *domain.OIDCSession
		isErr             bool
		err               error
	}

	codeVerifier, err := values.NewOIDCCodeVerifier()
	if err != nil {
		t.Errorf("failed to create code verifier: %v", err)
	}

	testCases := []test{
		{
			description:   "特に問題ないのでエラーなし",
			isTraQBroken:  false,
			code:          values.NewOIDCAuthorizationCode("code"),
			codeValid:     true,
			client:        domain.NewOIDCClient(values.NewOIDCClientID("clientID")),
			clientIDValid: true,
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			tokenResponse: &tokenResponse{
				AccessToken: "accessToken",
				TokenType:   "tokenType",
				ExpiresIn:   1,
			},
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("accessToken"),
				time.Now().Add(time.Second).Add(-5*time.Second),
			),
			codeVerifierValid: true,
		},
		{
			description:   "traQが壊れているのでエラー",
			isTraQBroken:  true,
			code:          values.NewOIDCAuthorizationCode("code"),
			codeValid:     true,
			client:        domain.NewOIDCClient(values.NewOIDCClientID("clientID")),
			clientIDValid: true,
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			codeVerifierValid: true,
			isErr:             true,
			err:               auth.ErrIdpBroken,
		},
		{
			description:   "clientIDが誤っているのでエラー",
			isTraQBroken:  false,
			code:          values.NewOIDCAuthorizationCode("code"),
			codeValid:     true,
			client:        domain.NewOIDCClient(values.NewOIDCClientID("")),
			clientIDValid: false,
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			codeVerifierValid: true,
			isErr:             true,
			err:               auth.ErrInvalidClient,
		},
		{
			description:   "codeが誤っているのでエラー",
			isTraQBroken:  false,
			code:          values.NewOIDCAuthorizationCode("code"),
			codeValid:     false,
			client:        domain.NewOIDCClient(values.NewOIDCClientID("clientID")),
			clientIDValid: true,
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			codeVerifierValid: true,
			isErr:             true,
			err:               auth.ErrInvalidCredentials,
		},
		{
			description:   "codeVerifierが誤っているのでエラー",
			isTraQBroken:  false,
			code:          values.NewOIDCAuthorizationCode("code"),
			codeValid:     true,
			client:        domain.NewOIDCClient(values.NewOIDCClientID("clientID")),
			clientIDValid: true,
			authState: domain.NewOIDCAuthState(
				values.OIDCCodeChallengeMethodSha256,
				codeVerifier,
			),
			codeVerifierValid: false,
			isErr:             true,
			err:               auth.ErrInvalidCredentials,
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
				isTraQBroken:      testCase.isTraQBroken,
				code:              string(testCase.code),
				codeValid:         testCase.codeValid,
				clientID:          string(testCase.client.GetClientID()),
				clientIDValid:     testCase.clientIDValid,
				codeVerifier:      string(testCase.authState.GetCodeVerifier()),
				codeVerifierValid: testCase.codeVerifierValid,
				tokenResponse:     testCase.tokenResponse,
			}

			session, err := oidcAuth.GetOIDCSession(ctx, testCase.client, testCase.code, testCase.authState)

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

			assert.Equal(t, testCase.session.GetAccessToken(), session.GetAccessToken())
			assert.WithinDuration(t, testCase.session.GetExpiresAt(), session.GetExpiresAt(), time.Second)
		})
	}
}

func TestRevokeOIDCSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type mockHandlerParam struct {
		isTraQBroken bool
		token        string
	}

	var (
		param      *mockHandlerParam
		handlerErr error
		callCount  int

		errNoParamSet      = errors.New("param is not set")
		errUnexpectedToken = errors.New("unexpected token")
	)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/oauth2/revoke" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
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

		token := r.FormValue("token")
		if token != param.token {
			handlerErr = errUnexpectedToken
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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
	oidcAuth, err := NewOIDC(mockConfig)
	if err != nil {
		t.Fatalf("Error creating OIDC: %v", err)
		return
	}

	type test struct {
		description  string
		isTraQBroken bool
		session      *domain.OIDCSession
		isErr        bool
		err          error
	}
	testCases := []test{
		{
			description:  "問題がないのでエラーなし",
			isTraQBroken: false,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("token"),
				time.Now().Add(time.Hour),
			),
		},
		{
			description:  "traQが壊れているのでエラー",
			isTraQBroken: true,
			session: domain.NewOIDCSession(
				values.NewOIDCAccessToken("token"),
				time.Now().Add(time.Hour),
			),
			isErr: true,
			err:   auth.ErrIdpBroken,
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
				isTraQBroken: testCase.isTraQBroken,
				token:        string(testCase.session.GetAccessToken()),
			}

			err := oidcAuth.RevokeOIDCSession(ctx, testCase.session)

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
		})
	}
}
