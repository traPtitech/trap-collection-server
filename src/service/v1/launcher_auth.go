package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherAuth struct {
	db                        repository.DB
	launcherVersionRepository repository.LauncherVersion
	launcherUserRepository    repository.LauncherUser
	launcherSessionRepository repository.LauncherSession
}

func NewLauncherAuth(
	db repository.DB,
	launcherVersionRepository repository.LauncherVersion,
	launcherUserRepository repository.LauncherUser,
	launcherSessionRepository repository.LauncherSession,
) *LauncherAuth {
	return &LauncherAuth{
		db:                        db,
		launcherVersionRepository: launcherVersionRepository,
		launcherUserRepository:    launcherUserRepository,
		launcherSessionRepository: launcherSessionRepository,
	}
}

func (la *LauncherAuth) CreateLauncherUser(ctx context.Context, launcherVersionID values.LauncherVersionID, userNum int) ([]*domain.LauncherUser, error) {
	_, err := la.launcherVersionRepository.GetLauncherVersion(ctx, launcherVersionID)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidLauncherVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	launcherUsers := make([]*domain.LauncherUser, 0, userNum)
	for i := 0; i < userNum; i++ {
		productKey, err := values.NewLauncherUserProductKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create product key: %w", err)
		}

		launcherUsers = append(launcherUsers, domain.NewLauncherUser(
			values.NewLauncherUserID(),
			productKey,
		))
	}

	launcherUsers, err = la.launcherUserRepository.CreateLauncherUsers(ctx, launcherVersionID, launcherUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher users: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return launcherUsers, nil
}
