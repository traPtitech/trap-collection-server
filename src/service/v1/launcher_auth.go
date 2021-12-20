package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

const expiresIn = 86400

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
	_, err := la.launcherVersionRepository.GetLauncherVersion(ctx, launcherVersionID, repository.LockTypeNone)
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

	return launcherUsers, nil
}

func (la *LauncherAuth) GetLauncherUsers(ctx context.Context, launcherVersionID values.LauncherVersionID) ([]*domain.LauncherUser, error) {
	_, err := la.launcherVersionRepository.GetLauncherVersion(ctx, launcherVersionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidLauncherVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	launcherUsers, err := la.launcherVersionRepository.GetLauncherUsersByLauncherVersionID(ctx, launcherVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher users: %w", err)
	}

	return launcherUsers, nil
}

func (la *LauncherAuth) RevokeProductKey(ctx context.Context, user values.LauncherUserID) error {
	err := la.launcherUserRepository.DeleteLauncherUser(ctx, user)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return service.ErrInvalidLauncherUser
	}
	if err != nil {
		return fmt.Errorf("failed to delete launcher user: %w", err)
	}

	return nil
}

func (la *LauncherAuth) LoginLauncher(ctx context.Context, productKey values.LauncherUserProductKey) (*domain.LauncherSession, error) {
	launcherUser, err := la.launcherUserRepository.GetLauncherUserByProductKey(ctx, productKey)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidLauncherUserProductKey
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	launcherSession := domain.NewLauncherSession(
		values.NewLauncherSessionID(),
		accessToken,
		getExpiresAt(),
	)

	launcherSession, err = la.launcherSessionRepository.CreateLauncherSession(ctx, launcherUser.GetID(), launcherSession)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher session: %w", err)
	}

	return launcherSession, nil
}

func getExpiresAt() time.Time {
	return time.Now().Add(expiresIn * time.Second)
}

func (la *LauncherAuth) LauncherAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken) (*domain.LauncherUser, *domain.LauncherVersion, error) {
	launcherVersion, launcherUser, launcherSession, err := la.launcherVersionRepository.GetLauncherVersionAndUserAndSessionByAccessToken(ctx, accessToken)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrInvalidLauncherSessionAccessToken
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get launcher version and user and session: %w", err)
	}

	if launcherSession.IsExpired() {
		return nil, nil, service.ErrLauncherSessionAccessTokenExpired
	}

	return launcherUser, launcherVersion, nil
}
