package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// LauncherUser
// ランチャー使用者を表すドメイン。
// 1プロダクトキーでのランチャー起動数に制限がないため、
// 漏れたときにRevoke可能なようにする。
type LauncherUser struct {
	id         values.LauncherUserID
	productKey values.LauncherUserProductKey
	status     values.LauncherUserStatus
	createdAt  time.Time
}

func NewLauncherUser(
	id values.LauncherUserID,
	productKey values.LauncherUserProductKey,
) *LauncherUser {
	return &LauncherUser{
		id:         id,
		productKey: productKey,
		status:     values.LauncherUserStatusActive,
	}
}

func NewProductKey(
	id values.LauncherUserID,
	productKey values.LauncherUserProductKey,
	status values.LauncherUserStatus,
	createdAt time.Time,
) *LauncherUser {
	return &LauncherUser{
		id:         id,
		productKey: productKey,
		status:     status,
		createdAt:  createdAt,
	}
}

func (lu *LauncherUser) GetID() values.LauncherUserID {
	return lu.id
}

func (lu *LauncherUser) GetProductKey() values.LauncherUserProductKey {
	return lu.productKey
}

func (lu *LauncherUser) GetStatus() values.LauncherUserStatus {
	return lu.status
}

func (lu *LauncherUser) SetStatus(status values.LauncherUserStatus) {
	lu.status = status
}

func (lu *LauncherUser) GetCreatedAt() time.Time {
	return lu.createdAt
}
