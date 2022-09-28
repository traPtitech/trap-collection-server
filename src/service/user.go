package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type User interface {
	GetMe(ctx context.Context, session *domain.OIDCSession) (*UserInfo, error)
	GetAllActiveUser(ctx context.Context, session *domain.OIDCSession) ([]*UserInfo, error)
}

// UserInfo 簡易的なtraP部員の情報
type UserInfo struct {
	ID     values.TraPMemberID
	Name   values.TraPMemberName
	Status values.TraPMemberStatus
}

func NewUserInfo(id values.TraPMemberID, name values.TraPMemberName, status values.TraPMemberStatus) *UserInfo {
	return &UserInfo{
		ID:     id,
		Name:   name,
		Status: status,
	}
}

func (ui *UserInfo) GetID() values.TraPMemberID {
	return ui.ID
}

func (ui *UserInfo) GetName() values.TraPMemberName {
	return ui.Name
}

func (ui *UserInfo) GetStatus() values.TraPMemberStatus {
	return ui.Status
}
