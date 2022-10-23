package v2

import "github.com/traPtitech/trap-collection-server/src/repository"

type AdminAuth struct {
	db              repository.DB
	adminRepository repository.AdminAuthV2
	user            *User
}

func NewAdminAuth(
	db repository.DB,
	adminRepository repository.AdminAuthV2,
	user *User,
) *AdminAuth {
	return &AdminAuth{
		db:              db,
		adminRepository: adminRepository,
		user:            user,
	}
}
