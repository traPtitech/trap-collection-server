package gorm2

type LauncherVersion struct {
	db *DB
}

func NewLauncherVersion(db *DB) *LauncherVersion {
	return &LauncherVersion{
		db: db,
	}
}
