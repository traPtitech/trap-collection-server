package v1

import (
	"context"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type AdministratorAuth struct {
	administrators []values.TraPMemberName
	userUtils      *UserUtils
}

func NewAdministratorAuth(conf config.ServiceV1, userUtils *UserUtils) (*AdministratorAuth, error) {
	strAdministrators, err := conf.Administrators()
	if err != nil {
		return nil, fmt.Errorf("failed to get administrators: %w", err)
	}

	administrators := make([]values.TraPMemberName, 0, len(strAdministrators))
	for _, strAdministrator := range strAdministrators {
		administrator := values.NewTrapMemberName(strAdministrator)

		err := administrator.Validate()
		if err != nil {
			return nil, fmt.Errorf("failed to validate administrator: %w", err)
		}

		administrators = append(administrators, values.NewTrapMemberName(strAdministrator))
	}

	return &AdministratorAuth{
		administrators: administrators,
		userUtils:      userUtils,
	}, nil
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
