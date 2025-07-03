package v2

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

// OIDC構造体がservice.OIDCV2インターフェイスを満たすことを示すおまじない
var _ service.OIDCV2 = &OIDC{}

type OIDC struct {
	client   *domain.OIDCClient
	user     *User
	oidcAuth auth.OIDC
}

func NewOIDC(conf config.ServiceV2, user *User, oidc auth.OIDC) (*OIDC, error) {
	strClientID, err := conf.ClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ID: %w", err)
	}

	clientID := values.NewOIDCClientID(strClientID)
	client := domain.NewOIDCClient(clientID)

	return &OIDC{
		client:   client,
		user:     user,
		oidcAuth: oidc,
	}, nil
}

func (o *OIDC) GenerateAuthState(_ context.Context) (*domain.OIDCClient, *domain.OIDCAuthState, error) {
	codeChallengeMethod := values.OIDCCodeChallengeMethodSha256
	codeChallenge, err := values.NewOIDCCodeVerifier()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	state := domain.NewOIDCAuthState(codeChallengeMethod, codeChallenge)

	return o.client, state, nil
}

func (o *OIDC) Callback(ctx context.Context, authState *domain.OIDCAuthState, code values.OIDCAuthorizationCode) (*domain.OIDCSession, error) {
	session, err := o.oidcAuth.GetOIDCSession(ctx, o.client, code, authState)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return nil, service.ErrInvalidAuthStateOrCode
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC session: %w", err)
	}

	return session, nil
}

func (o *OIDC) Logout(ctx context.Context, session *domain.OIDCSession) error {
	err := o.oidcAuth.RevokeOIDCSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to revoke OIDC session: %w", err)
	}

	return nil
}

func (o *OIDC) Authenticate(_ context.Context, session *domain.OIDCSession) error {
	// traQで凍結された場合の反映が遅れるのは許容しているので、sessionの有効期限確認のみ
	if session.IsExpired() {
		return service.ErrOIDCSessionExpired
	}

	return nil
}

func (o *OIDC) GetMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error) {
	return o.user.getMe(ctx, session)
}

func (o *OIDC) GetActiveUsers(ctx context.Context, session *domain.OIDCSession, includeBot bool) ([]*service.UserInfo, error) {
	users, err := o.user.getActiveUsers(ctx, session)
	if err != nil {
		return nil, err
	}
	if includeBot {
		return users, nil
	}
	filteredUsers := make([]*service.UserInfo, 0, len(users))
	for _, user := range users {
		if user.GetBot() {
			continue
		}
		filteredUsers = append(filteredUsers, user)
	}
	return filteredUsers, nil
}

// User
// traPメンバーの情報取得周りをキャッシュの使用も含めて行う。
type User struct {
	userAuth  auth.User
	userCache cache.User
}

func NewUser(userAuth auth.User, userCache cache.User) *User {
	return &User{
		userAuth:  userAuth,
		userCache: userCache,
	}
}

// getMe
// セッションから対応するtraQのユーザー情報を取得する。
// traQでの凍結・凍結解除の反映までに最大1時間の遅延が発生する点に注意。
func (uu *User) getMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error) {
	user, err := uu.userCache.GetMe(ctx, session.GetAccessToken())
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// cacheからの取り出しに失敗してもauthからとって来れれば良いので、returnはしない
		log.Printf("error: failed to get user info: %v\n", err)
	}
	// cacheから取り出した場合はそれを返す
	if err == nil {
		return user, nil
	}

	user, err = uu.userAuth.GetMe(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	err = uu.userCache.SetMe(ctx, session, user)
	if err != nil {
		// cacheの設定に失敗してもreturnはしない
		log.Printf("error: failed to set user info: %v\n", err)
	}

	return user, nil
}

// getActiveUsers
// traQのアクティブユーザー(凍結されていないユーザー)一覧を取得する。
func (uu *User) getActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	users, err := uu.userCache.GetActiveUsers(ctx)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// cacheからの取り出しに失敗してもauthからとって来れれば良いので、returnはしない
		log.Printf("error: failed to get user info: %v\n", err)
	}
	// cacheから取り出した場合はそれを返す
	if err == nil {
		return users, nil
	}

	users, err = uu.userAuth.GetActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	err = uu.userCache.SetActiveUsers(ctx, users)
	if err != nil {
		// cacheの設定に失敗してもreturnはしない
		log.Printf("error: failed to set user info: %v\n", err)
	}

	return users, nil
}
