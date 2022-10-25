package gorm2

type AdminAuth struct {
	db *DB
}

func NewAdminAuth(db *DB) *AdminAuth {
	return &AdminAuth{
		db: db,
	}
}
