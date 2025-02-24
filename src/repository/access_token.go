package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type AccessToken interface {
	// SaveAccessToken
	// アクセストークンの保存。
	SaveAccessToken(ctx context.Context, productKeyID values.LauncherUserID, accessToken *domain.LauncherSession) error
	// GetAccessTokenInfo
	// アクセストークンの情報の取得。
	GetAccessTokenInfo(ctx context.Context, accessToken values.LauncherSessionAccessToken, lockType LockType) (*AccessTokenInfo, error)
}

type AccessTokenInfo struct {
	AccessToken *domain.LauncherSession
	ProductKey  *domain.LauncherUser
	Edition     *domain.LauncherVersion
}
