package service

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameAuth interface {
	AddGameCollaborators(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userIDs []values.TraPMemberID) error
	UpdateGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error
	RemoveGameCollaborator(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error
	GetGameManagers(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) ([]*GameManager, error)
	UpdateGameAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
	UpdateGameManagementRoleAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error
}

type GameManager struct {
	UserID     values.TraPMemberID
	UserName   values.TraPMemberName
	UserStatus values.TraPMemberStatus
	Role       values.GameManagementRole
}
