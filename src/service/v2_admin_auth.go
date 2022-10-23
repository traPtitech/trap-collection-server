package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type AdminAuthV2 interface {
	//AddAdmin
	//adminの追加
	AddAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*UserInfo, error)
	//GetAdmins
	//adminを全取得
	GetAdmins(ctx context.Context, session *domain.OIDCSession) ([]*UserInfo, error)
	//DeleteAdmin
	//adminを削除
	DeleteAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*UserInfo, error)
}
