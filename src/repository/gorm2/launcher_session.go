package gorm2

type LauncherSession struct{
	db *DB
}

func NewLauncherSession(db *DB) *LauncherSession {
	return &LauncherSession{
		db: db,
	}
}
