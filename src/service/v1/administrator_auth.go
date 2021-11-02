package v1

import (
	"context"
	"fmt"

	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type AdministratorAuth struct {
	administrators []values.TraPMemberName
	userUtils      *UserUtils
}

func NewAdministratorAuth(administrators common.Administrators, userUtils *UserUtils) *AdministratorAuth {
	return &AdministratorAuth{
		administrators: administrators,
		userUtils:      userUtils,
	}
}

func (aa *AdministratorAuth) AdministratorAuth(ctx context.Context, session *domain.OIDCSession) error {
	user, err := aa.userUtils.getMe(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	for _, administrator := range aa.administrators {
		if user.GetName() == administrator {
			return nil
		}
	}

	return service.ErrForbidden
}
