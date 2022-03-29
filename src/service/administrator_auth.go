package service

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

/*
	AdministratorAuth
	アプリケーション全体の管理者認証のサービス。
	ランチャーのバージョン作成など、ランチャー関連の操作の権限を持つ。
	現在は管理者を環境変数から読み込むことで対応。
	要改善。
*/
type AdministratorAuth interface {
	AdministratorAuth(ctx context.Context, session *domain.OIDCSession) error
}
