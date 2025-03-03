package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type User interface {
	GetMe(ctx context.Context, session *domain.OIDCSession) (*UserInfo, error)
	GetAllActiveUser(ctx context.Context, session *domain.OIDCSession, includeBot bool) ([]*UserInfo, error)
}

// UserInfo 簡易的なtraP部員の情報
type UserInfo struct {
	id     values.TraPMemberID
	name   values.TraPMemberName
	status values.TraPMemberStatus
	bot    bool
}

func NewUserInfo(id values.TraPMemberID, name values.TraPMemberName, status values.TraPMemberStatus, bot bool) *UserInfo {
	return &UserInfo{
		id:     id,
		name:   name,
		status: status,
		bot:    bot,
	}
}

func (ui *UserInfo) GetID() values.TraPMemberID {
	return ui.id
}

func (ui *UserInfo) GetName() values.TraPMemberName {
	return ui.name
}

func (ui *UserInfo) GetStatus() values.TraPMemberStatus {
	return ui.status
}

func (ui *UserInfo) GetBot() bool {
	return ui.bot
}
