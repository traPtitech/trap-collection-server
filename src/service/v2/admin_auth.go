package v2

import (
	"context"
	"errors"
	"fmt"

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
	var adminInfos []*service.UserInfo
	err := aa.db.Transaction(ctx, nil, func(ctx context.Context) error {
		activeUsers, err := aa.user.getActiveUsers(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to get active users: %w", err)
		}
		activeUsersMap := make(map[values.TraPMemberID]*service.UserInfo, len(activeUsers))
		for _, activeUser := range activeUsers {
			activeUsersMap[activeUser.GetID()] = activeUser
		}
		if _, ok := activeUsersMap[userID]; !ok {
			return service.ErrInvalidUserID
		}

		adminIDs, err := aa.adminAuthRepository.GetAdmins(ctx)
		if err != nil {
			return fmt.Errorf("failed to get admins: %w", err)
		}

		for _, adminID := range adminIDs {
			if adminID == userID { //ユーザーがすでにadmin
				return service.ErrNoAdminsUpdated
			}
		}

		err = aa.adminAuthRepository.AddAdmin(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to add admin: %w", err)
		}

		adminInfos = make([]*service.UserInfo, 0, len(adminIDs)+1)
		for _, adminID := range adminIDs {
			if activeUsersMap[adminID].GetStatus() == values.TrapMemberStatusActive {
				adminInfos = append(adminInfos, activeUsersMap[adminID])
			}
		}
		adminInfos = append(adminInfos, activeUsersMap[userID])
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}
	return adminInfos, nil
}

func (aa *AdminAuth) GetAdmins(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	activeUsers, err := aa.user.getActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	activeUsersMap := make(map[values.TraPMemberID]*service.UserInfo, len(activeUsers))
	for _, activeUser := range activeUsers {
		activeUsersMap[activeUser.GetID()] = activeUser
	}

	adminIDs, err := aa.adminAuthRepository.GetAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	adminsInfo := make([]*service.UserInfo, 0, len(adminIDs))
	for _, adminID := range adminIDs {
		if adminInfo, ok := activeUsersMap[adminID]; ok {
			adminsInfo = append(adminsInfo, adminInfo)
		}
	}

	return adminsInfo, nil
}

func (aa *AdminAuth) DeleteAdmin(ctx context.Context, session *domain.OIDCSession, userID values.TraPMemberID) ([]*service.UserInfo, error) {
	activeUsers, err := aa.user.getActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	activeUsersMap := make(map[values.TraPMemberID]*service.UserInfo, len(activeUsers))
	for _, activeUser := range activeUsers {
		activeUsersMap[activeUser.GetID()] = activeUser
	}
	if _, ok := activeUsersMap[userID]; !ok {
		return nil, service.ErrInvalidUserID
	}

	err = aa.adminAuthRepository.DeleteAdmin(ctx, userID)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return nil, service.ErrNotAdmin
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete admin: %w", err)
	}

	adminIDs, err := aa.adminAuthRepository.GetAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	adminsInfo := make([]*service.UserInfo, 0, len(adminIDs))
	for _, adminID := range adminIDs {
		if adminInfo, ok := activeUsersMap[adminID]; ok {
			adminsInfo = append(adminsInfo, adminInfo)
		}
	}
	return adminsInfo, nil
}

func (aa *AdminAuth) AdminAuthorize(ctx context.Context, session *domain.OIDCSession) error {
	if session.IsExpired() {
		return service.ErrOIDCSessionExpired
	}

	userInfo, err := aa.user.getMe(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	adminsID, err := aa.adminAuthRepository.GetAdmins(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admins: %w", err)
	}
	for _, adminID := range adminsID {
		if adminID == userInfo.GetID() {
			return nil
		}
	}
	return service.ErrForbidden
}
