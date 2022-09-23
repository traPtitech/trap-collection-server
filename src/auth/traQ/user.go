package traq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User struct {
	client  *http.Client
	baseURL *url.URL
}

func NewUser(conf config.AuthTraQ) (*User, error) {
	client, err := conf.HTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	baseURL, err := conf.BaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get base url: %w", err)
	}

	return &User{
		client:  client,
		baseURL: baseURL,
	}, nil
}

type getUsersMeResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	State int       `json:"state"`
}

func (u *User) GetMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error) {
	path := *u.baseURL
	path.Path += "/users/me"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.GetAccessToken()))

	res, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, auth.ErrInvalidSession
	case http.StatusInternalServerError:
		return nil, auth.ErrIdpBroken
	default:
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response getUsersMeResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var status values.TraPMemberStatus
	switch response.State {
	case 0:
		status = values.TrapMemberStatusDeactivated
	case 1:
		status = values.TrapMemberStatusActive
	case 2:
		status = values.TrapMemberStatusSuspended
	default:
		return nil, fmt.Errorf("unexpected state: %d", response.State)
	}

	return service.NewUserInfo(
		values.NewTrapMemberID(response.ID),
		values.NewTrapMemberName(response.Name),
		status,
	), nil
}

type getUsersResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	State int       `json:"state"`
}

func (u *User) GetAllActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	path := *u.baseURL
	path.Path += "/users"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.GetAccessToken()))

	res, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, auth.ErrInvalidSession
	case http.StatusInternalServerError:
		return nil, auth.ErrIdpBroken
	default:
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response []*getUsersResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	users := make([]*service.UserInfo, 0, len(response))
	for _, user := range response {
		var status values.TraPMemberStatus
		switch user.State {
		case 0:
			status = values.TrapMemberStatusDeactivated
		case 1:
			status = values.TrapMemberStatusActive
		case 2:
			status = values.TrapMemberStatusSuspended
		default:
			return nil, fmt.Errorf("unexpected state: %d", user.State)
		}

		users = append(users, service.NewUserInfo(
			values.NewTrapMemberID(user.ID),
			values.NewTrapMemberName(user.Name),
			status,
		))
	}

	return users, nil
}

func (u *User) GetActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	path := *u.baseURL
	path.Path += "/users"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.GetAccessToken()))

	res, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, auth.ErrInvalidSession
	case http.StatusInternalServerError:
		return nil, auth.ErrIdpBroken
	default:
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response []*getUsersResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	users := make([]*service.UserInfo, 0, len(response))
	for _, user := range response {
		var status values.TraPMemberStatus
		switch user.State {
		case 0:
			status = values.TrapMemberStatusDeactivated
		case 1:
			status = values.TrapMemberStatusActive
		case 2:
			status = values.TrapMemberStatusSuspended
		default:
			return nil, fmt.Errorf("unexpected state: %d", user.State)
		}

		users = append(users, service.NewUserInfo(
			values.NewTrapMemberID(user.ID),
			values.NewTrapMemberName(user.Name),
			status,
		))
	}

	return users, nil
}
