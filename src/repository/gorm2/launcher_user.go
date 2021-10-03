package gorm2

type LauncherUser struct {
	db *DB
}

func NewLauncherUser(db *DB) *LauncherUser {
	return &LauncherUser{
		db: db,
	}
}
