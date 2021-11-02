package v1

import (
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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
