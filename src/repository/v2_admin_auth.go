package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type AdminAuthV2 interface {
	AddAdmin(ctx context.Context, userID values.TraPMemberID) ([]*values.TraPMemberID, error)
	GetAdmins(ctx context.Context) ([]*values.TraPMemberID, error)
	Delete(ctx context.Context, userID values.TraPMemberID) ([]*values.TraPMemberID, error)
}
