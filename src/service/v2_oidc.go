package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// OIDCV2
// v2用のOIDC・OAuth2.0(Authorization Code Flow)を使い認証を行うサービス
// ref: https://trap.jp/post/1007/#:~:text=%E3%81%91%E3%81%BE%E3%81%9B%E3%82%93%E3%80%82-,Authorization%20Code%20Flow,-%E6%B5%81%E3%82%8C
type OIDCV2 interface {
	// GenerateAuthState
	// AuthorizationCodeを認可サーバーにリクエストし、
	// ブラウザに返すために必要なOIDCClientとOIDCAuthStateを返す
	// refの「1. ブラウザで認証画面を開く」に相当
	GenerateAuthState(ctx context.Context) (*domain.OIDCClient, *domain.OIDCAuthState, error)
	// Callback
	// ブラウザからAuthorization Codeなどを受け取り、
	// traQからトークンを取得して、
	// ログインセッションを発行する
	// refの「6. Authorization Codeをアプリケーション（Back-End）へ渡す」から
	// 「9. ログインセッションを発行する」に相当
	Callback(ctx context.Context, authState *domain.OIDCAuthState, code values.OIDCAuthorizationCode) (*domain.OIDCSession, error)
	// Logout
	// traQにトークンは気のリクエストをした上で、
	// ログインセッションを破棄する
	Logout(ctx context.Context, session *domain.OIDCSession) error
	// Authenticate
	// ログインセッションを検証する
	// セッションの有効期限が切れている場合、ErrOIDCSessionExpiredを返す
	Authenticate(ctx context.Context, session *domain.OIDCSession) error
	// GetMe
	// sessionに対応するtraQのユーザー情報を取得する
	GetMe(ctx context.Context, session *domain.OIDCSession) (*UserInfo, error)
	// GetActiveUsers
	// traQの全アクティブユーザー(凍結されていないユーザー)情報を取得する
	GetActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*UserInfo, error)
}
