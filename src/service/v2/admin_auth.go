package v2

import (
	"context"
	"fmt"
	"log"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type AdminAuth struct {
	db                  repository.DB
	adminAuthRepository repository.AdminAuthV2
	user                *User
}

func NewAdminAuth(
	db repository.DB,
	adminAuthRepository repository.AdminAuthV2,
	user *User,
) *AdminAuth {
	return &AdminAuth{
		db:                  db,
		adminAuthRepository: adminAuthRepository,
		user:                user,
	}
}

func (aa *AdminAuth) AddAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*service.UserInfo, error) {
	activeUsers, err := aa.user.getActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	activeUsersMap := make(map[values.TraPMemberID]values.TraPMemberName, len(activeUsers))
	for _, activeUser := range activeUsers {
		activeUsersMap[activeUser.GetID()] = activeUser.GetName()
	}
	if _, ok := activeUsersMap[userID]; !ok {
		return nil, service.ErrInvalidUserID
	}

	adminIDs, err := aa.adminAuthRepository.GetAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %v", err)
	}

	adminsMap := make(map[values.TraPMemberID]struct{})
	for _, adminID := range adminIDs {
		adminsMap[adminID] = struct{}{}
	}
	if _, ok := adminsMap[userID]; ok { //ユーザーがすでにadmin
		return nil, service.ErrNoAdminsUpdated
	}

	err = aa.adminAuthRepository.AddAdmin(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to add admin: %v", err)
	}

	adminInfos := make([]*service.UserInfo, len(adminIDs))
	for _, adminID := range adminIDs {
		adminInfos = append(adminInfos, service.NewUserInfo(
			adminID,
			activeUsersMap[adminID],
			values.TrapMemberStatusActive,
		))
	}
	adminInfos = append(adminInfos, service.NewUserInfo(
		userID,
		activeUsersMap[userID],
		values.TrapMemberStatusActive,
	))

	return adminInfos, nil
}

func (aa *AdminAuth) GetAdmins(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	activeUsers, err := aa.user.getActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	activeUsersMap := make(map[values.TraPMemberID]values.TraPMemberName, len(activeUsers))
	for _, activeUser := range activeUsers {
		activeUsersMap[activeUser.GetID()] = activeUser.GetName()
	}

	adminIDs, err := aa.adminAuthRepository.GetAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %v", err)
	}

	adminsInfo := make([]*service.UserInfo, len(adminIDs))
	for _, adminID := range adminIDs {
		if adminName, ok := activeUsersMap[adminID]; ok {
			adminsInfo = append(adminsInfo, service.NewUserInfo(
				adminID,
				adminName,
				values.TrapMemberStatusActive,
			))
		} else {
			//adminが凍結されているとき、一応ログを残す。
			log.Printf("not active user: %v\n", adminID)
		}
	}

	return adminsInfo, nil
}
