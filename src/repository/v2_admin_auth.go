package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type AdminAuthV2 interface {
	//AddAdmin
	//adminを追加
	AddAdmin(ctx context.Context, userID values.TraPMemberID) error
	//GetAdmins
	//adminを全員取得してuserIDを返す。
	GetAdmins(ctx context.Context) ([]values.TraPMemberID, error)
	//DeleteAdmin
	//adminを削除
	//ユーザーが存在しない場合、ErrNoRecordDeletedを返す。
	DeleteAdmin(ctx context.Context, userID values.TraPMemberID) error
}
