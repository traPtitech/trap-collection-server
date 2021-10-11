package traq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type OIDC struct {
	client  *http.Client
	baseURL *url.URL
}

func NewOIDC(client *http.Client, baseURL common.TraQBaseURL) *OIDC {
	return &OIDC{
		client:  client,
		baseURL: (*url.URL)(baseURL),
	}
}

type postOAuth2TokenResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (o *OIDC) GetOIDCSession(ctx context.Context, client *domain.OIDCClient, code values.OIDCAuthorizationCode, authState *domain.OIDCAuthState) (*domain.OIDCSession, error) {
	path := *o.baseURL
	path.Path += "/oauth2/token"

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", string(client.GetClientID()))
	form.Set("code", string(code))
	form.Set("code_verifier", string(authState.GetCodeVerifier()))
	reqBody := strings.NewReader(form.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusBadRequest:
		return nil, auth.ErrInvalidCredentials
	case http.StatusUnauthorized:
		return nil, auth.ErrInvalidClient
	case http.StatusInternalServerError:
		return nil, auth.ErrIdpBroken
	default:
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response postOAuth2TokenResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return domain.NewOIDCSession(
		values.NewOIDCAccessToken(response.AccessToken),
		// 期限切れでなければ確実に使用可能にするためにサーバー上での期限は少し短めにしている
		time.Now().Add(time.Duration(response.ExpiresIn-5)*time.Second),
	), nil
}

func (o *OIDC) RevokeOIDCSession(ctx context.Context, session *domain.OIDCSession) error {
	path := *o.baseURL
	path.Path += "/oauth2/revoke"

	form := url.Values{}
	form.Set("token", string(session.GetAccessToken()))
	reqBody := strings.NewReader(form.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusInternalServerError:
		return auth.ErrIdpBroken
	default:
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}
