package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type AdminAuthV2 interface {
	//AddAdmin
	//adminの追加
	//ユーザーが存在しないとき、ErrInvalidUserIDを返す。
	//既にユーザーがadminのとき、ErrNoAdminsUpdatedを返す。
	AddAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*UserInfo, error)
	//GetAdmins
	//adminを全取得
	GetAdmins(ctx context.Context, session *domain.OIDCSession) ([]*UserInfo, error)
	//DeleteAdmin
	//adminを削除
	//ユーザーが存在しないとき、ErrInvalidUserIDを返す。
	//ユーザーがadminでないとき、ErrNotAdminを返す。
	//自分自身を削除しようとしたとき、ErrCannotDeleteMeFromAdminを返す。
	DeleteAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*UserInfo, error)
	//AdminAuthorize
	//ログイン中のユーザーがadminかどうか判断する。
	//adminでなければErrForbiddenを返し、adminならnilを返す。
	//sessionが切れている場合はErrOIDCSessionExpiredを返す。
	AdminAuthorize(ctx context.Context, session *domain.OIDCSession) error
}
