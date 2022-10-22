package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GameRoleV2 interface {
	//EditGameManagementRole
	//指定されたユーザーを指定されたゲームの管理者に指定したり変更したりする。
	//管理者でない場合は、roleが追加される。
	//既に管理者の場合は、指定されたroleが現状と異なれば変更し、現状と同じ場合はErrNoGameManagementRoleUpdatedを返す。
	//ゲームIDに当てはまるゲームが存在しないとき、ErrInvalidGameIDを返す。
	//ユーザーが存在しない、またはアクティブではない場合、ErrInvalidUserIDを返す。
	EditGameManagementRole(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error
	//RemoveGameManagementRole
	//指定されたユーザーの指定されたゲームでのroleを削除する。
	//そのゲームのroleを持っていない場合は、ErrInvalidRoleを返す。
	//そのユーザーを消すとowners(administraitors)がいなくなってしまう場合は、ErrCannotDeleteOwnerを返す。
	//ゲームIDに当てはまるゲームが存在しないとき、ErrInvalidGameIDを返す。
	RemoveGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error
	//UpdateGameAuth
	//ログイン中のユーザーがゲームの情報を更新する権限を持っているか、
	//すなわちowners(administraitors)とmaintainers(collaborators)のどちらかであるかを調べる。
	//権限を持っていない場合、ErrForbiddenを返す。持っている場合はnilを返す。
	//ゲームIDに当てはまるゲームが存在しないとき、ErrInvalidGameIDを返す。
	UpdateGameAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
	//UpdateGameManagementRoleAuth
	//ログイン中のユーザーがゲームの管理者たちを編集する権限を持っているか、
	//すなわちowners(administrators)であるかを調べる。
	//権限を持っていない場合、ErrForbiddenを返す。持っている場合はnilを返す。
	//ゲームIDに当てはまるゲームが存在しないとき、ErrInvalidGameIDを返す。
	UpdateGameManagementRoleAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
}

type GameManagerV2 struct {
	UserID     values.TraPMemberID
	UserName   values.TraPMemberName
	UserStatus values.TraPMemberStatus
	Role       values.GameManagementRole
}
