package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameManagementRole interface {
	AddGameManagementRoles(ctx context.Context, gameID values.GameID, userIDs []values.TraPMemberID, role values.GameManagementRole) error
	UpdateGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error
	RemoveGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error
	GetGameManagersByGameID(ctx context.Context, gameID values.GameID) ([]*UserIDAndManagementRole, error)
	GetGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, lockType LockType) (values.GameManagementRole, error)
}

type UserIDAndManagementRole struct {
	UserID values.TraPMemberID
	Role   values.GameManagementRole
}
