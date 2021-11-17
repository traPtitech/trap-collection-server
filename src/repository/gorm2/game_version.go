package gorm2

type GameVersion struct {
	db *DB
}

func NewGameVersion(db *DB) *GameVersion {
	return &GameVersion{
		db: db,
	}
}
