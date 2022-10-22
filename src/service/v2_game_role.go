package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GameRoleV2 interface {
	EditGameManagementRole(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error
	RemoveGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error
	UpdateGameAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
	UpdateGameManagementRoleAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
}

type GameManagerV2 struct {
	UserID     values.TraPMemberID
	UserName   values.TraPMemberName
	UserStatus values.TraPMemberStatus
	Role       values.GameManagementRole
}
